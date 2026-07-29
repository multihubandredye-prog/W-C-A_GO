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
		return response, pkgError.InternalServerError(fmt.Sprintf("Failed to initiate call: %v", err))
	}

	durationSec := 15
	if request.Duration != nil && *request.Duration > 0 {
		durationSec = *request.Duration
	}

	if mp3Source != nil {
		call.Play(mp3Source)
	}

	// Schedule call termination after the duration
	go func() {
		time.Sleep(time.Duration(durationSec) * time.Second)
		_ = call.Hangup()
	}()

	response.CallID = call.ID()
	response.Status = fmt.Sprintf("Call initiated successfully to %s (duration %ds)", request.Phone, durationSec)
	return response, nil
}
