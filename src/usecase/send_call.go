package usecase

import (
	"context"
	"fmt"
	"io"
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
	if request.Audio != nil {
		// Uploaded file (multipart): persist to a temp file so meowcaller's
		// MP3 decoder can stream it, exactly like the URL/base64 paths.
		uploaded, errOpen := request.Audio.Open()
		if errOpen != nil {
			return response, pkgError.ValidationError(fmt.Sprintf("failed to open the uploaded audio file: %v", errOpen))
		}
		audioBytes, errRead := io.ReadAll(uploaded)
		_ = uploaded.Close()
		if errRead != nil {
			return response, pkgError.ValidationError(fmt.Sprintf("failed to read the uploaded audio file: %v", errRead))
		}
		tempAudioPath = fmt.Sprintf("%s/temp_call_%s.mp3", config.PathMedia, fiberUtils.UUIDv4())
		if errWrite := os.WriteFile(tempAudioPath, audioBytes, 0644); errWrite != nil {
			return response, pkgError.InternalServerError(fmt.Sprintf("failed to write temp audio file: %v", errWrite))
		}
		deleteTempFile = true
		if mp3, errMP3 := meowcaller.MP3File(tempAudioPath); errMP3 == nil {
			mp3Source = mp3
		} else {
			// Fail fast with a clear message: the upload is a user action,
			// unlike the legacy paths which only warn and keep the call.
			return response, pkgError.ValidationError("the uploaded audio file must be a valid MP3")
		}
	} else if request.AudioURL != "" {
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

	// The meowcaller client must be the one created before the whatsmeow
	// client connected (it carries the low-level call interception that
	// learns the relay endpoint when the peer answers). Creating it here,
	// per request, silently breaks answered calls into an eternal
	// "Conectando...". Reuse the device's caller; the fallback exists only
	// for unexpected paths and logs loudly.
	var caller *meowcaller.Client
	if instance, ok := whatsapp.DeviceFromContext(ctx); ok && instance != nil {
		caller = instance.GetCaller()
	}
	if caller == nil {
		logrus.Warn("meowcaller client unavailable on device instance; creating it after connect (answered calls may stay on \"Conectando...\" until the API process restarts)")
		caller = meowcaller.NewClient(client)
	}

	call, err := caller.Call(ctx, recipient.String())
	if err != nil {
		if mp3Source != nil {
			_ = mp3Source.Close()
		}
		return response, pkgError.InternalServerError(fmt.Sprintf("Failed to initiate call: %v", err))
	}

	// Outbound media must flow from the moment the call is placed: the relay
	// only bridges the peer's media after seeing our stream, which is what
	// takes an answered call out of "Conectando...". Play explicit silence
	// right away (empirically required: without it answered calls stayed on
	// "Conectando..."), then swap it for the real audio when the call becomes
	// active so the peer hears the file from the very beginning.
	warmupSilence := call.Play(meowcaller.PCMStream(endlessSilence{}))

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

	call.OnReady(func() {
		logrus.Infof("Call %s is ready (media flowing); starting duration timer (%ds max)", call.ID(), durationSec)

		if mp3Source != nil {
			// Swap the warm-up silence for the real audio: the peer hears the
			// file from the very beginning.
			warmupSilence.Stop()
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
		warmupSilence.Stop()
		if mp3Source != nil {
			// AudioSource.Close is safe to call more than once, so this also
			// covers calls that ended before the audio was ever played.
			_ = mp3Source.Close()
		}
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

// endlessSilence is an io.ReadCloser that yields an infinite stream of zero
// bytes — raw s16le silence for meowcaller.PCMStream. It keeps outbound RTP
// flowing from dial time until the real audio takes over, without ever
// surfacing sound to the peer.
type endlessSilence struct{}

func (endlessSilence) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func (endlessSilence) Close() error { return nil }
