package usecase

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/validations"
	fiberUtils "github.com/gofiber/fiber/v2/utils"
	"github.com/purpshell/meowcaller"
	"github.com/sirupsen/logrus"
)

// SendCall initiates a WhatsApp VoIP audio call to the recipient.
// If AudioPath (local path or Base64) or AudioURL is provided, the audio file is streamed when the peer answers.
func (service serviceSend) SendCall(ctx context.Context, request domainSend.CallRequest) (response domainSend.CallResponse, err error) {
	err = validations.ValidateSendCall(ctx, request)
	if err != nil {
		return response, err
	}

	client := whatsapp.ClientFromContext(ctx)
	if client == nil {
		return response, pkgError.ErrWaCLI
	}

	recipient, err := utils.ValidateJidWithLogin(client, request.Phone)
	if err != nil {
		return response, err
	}

	var tempAudioPath string
	var deleteTempFile bool
	defer func() {
		if deleteTempFile && tempAudioPath != "" {
			_ = os.Remove(tempAudioPath)
		}
	}()

	var mp3Source meowcaller.AudioSource
	if request.AudioURL != "" {
		audioBytes, _, errDownload := utils.DownloadAudioFromURL(request.AudioURL)
		if errDownload != nil {
			return response, pkgError.ValidationError(fmt.Sprintf("failed to download audio from URL: %v", errDownload))
		}
		tempAudioPath = fmt.Sprintf("%s/temp_call_%s.mp3", config.PathMedia, fiberUtils.UUIDv4())
		if errWrite := os.WriteFile(tempAudioPath, audioBytes, 0644); errWrite != nil {
			return response, pkgError.InternalServerError(fmt.Sprintf("failed to write temp audio file: %v", errWrite))
		}
		deleteTempFile = true
		if mp3, errMP3 := meowcaller.MP3File(tempAudioPath); errMP3 == nil {
			mp3Source = mp3
		} else {
			logrus.Warnf("Failed to open downloaded MP3 file %s for call: %v", tempAudioPath, errMP3)
		}
	} else if request.AudioPath != "" {
		// 1) First check if it is an existing file on the server filesystem
		if _, errStat := os.Stat(request.AudioPath); errStat == nil {
			if mp3, errMP3 := meowcaller.MP3File(request.AudioPath); errMP3 == nil {
				mp3Source = mp3
			} else {
				logrus.Warnf("Failed to open local MP3 file %s for call: %v", request.AudioPath, errMP3)
			}
		} else {
			// 2) Otherwise try decoding as Base64 string
			audioBytes, errBase64 := utils.Base64ToBytes(request.AudioPath)
			if errBase64 != nil {
				return response, pkgError.ValidationError("audio_path: file does not exist on server filesystem and is not a valid base64 audio string. For remote files, use audio_url or send a Base64 string.")
			}
			tempAudioPath = fmt.Sprintf("%s/temp_call_%s.mp3", config.PathMedia, fiberUtils.UUIDv4())
			if errWrite := os.WriteFile(tempAudioPath, audioBytes, 0644); errWrite != nil {
				return response, pkgError.InternalServerError(fmt.Sprintf("failed to write temp base64 audio file: %v", errWrite))
			}
			deleteTempFile = true
			if mp3, errMP3 := meowcaller.MP3File(tempAudioPath); errMP3 == nil {
				mp3Source = mp3
			} else {
				logrus.Warnf("Failed to open base64 MP3 file %s for call: %v", tempAudioPath, errMP3)
			}
		}
	}

	caller := meowcaller.NewClient(client)
	call, err := caller.Call(ctx, recipient.String())
	if err != nil {
		if mp3Source != nil {
			_ = mp3Source.Close()
		}
		return response, pkgError.InternalServerError(fmt.Sprintf("Failed to initiate call: %v", err))
	}

	// durationSec is the maximum wall time the call may last. When audio is
	// provided the call also ends as soon as the audio finishes, whichever
	// comes first.
	durationSec := 15
	if request.Duration != nil && *request.Duration > 0 {
		durationSec = *request.Duration
	}

	call.OnPeerAccept(func() {
		logrus.Infof("Peer accepted call %s; relay negotiation in progress...", call.ID())
	})

	// meowcaller keeps the relay bridged with automatic silence frames while no
	// player is attached, so the audio only needs to start when the call is
	// actually ready: the peer hears the file from the very beginning instead
	// of joining mid-playback.
	call.OnReady(func() {
		logrus.Infof("Call %s is ready (media flowing); starting duration timer (%ds max)", call.ID(), durationSec)

		if mp3Source != nil {
			logrus.Infof("Call %s: playing audio for the peer", call.ID())
			player := call.Play(mp3Source)
			// Hang up as soon as the audio ends: no dead silence until the
			// duration timer expires.
			player.OnFinish(func() {
				logrus.Infof("Call %s: audio finished; hanging up", call.ID())
				_ = call.Hangup()
			})
		}

		go func() {
			time.Sleep(time.Duration(durationSec) * time.Second)
			_ = call.Hangup()
		}()
	})

	call.OnEnd(func(reason string) {
		logrus.Infof("Call %s ended with reason: %s", call.ID(), reason)
	})

	// Fallback timeout in case the call rings forever and is never answered or ready
	go func() {
		time.Sleep(time.Duration(durationSec+45) * time.Second)
		_ = call.Hangup()
	}()

	response.CallID = call.ID()
	response.Status = fmt.Sprintf("Call initiated successfully to %s (max duration %ds; audio starts when the peer answers)", request.Phone, durationSec)
	return response, nil
}
