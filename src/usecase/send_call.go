package usecase

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/multihubandredye-prog/W-C-A_GO/config"
	domainSend "github.com/multihubandredye-prog/W-C-A_GO/domains/send"
	"github.com/multihubandredye-prog/W-C-A_GO/infrastructure/whatsapp"
	pkgError "github.com/multihubandredye-prog/W-C-A_GO/pkg/error"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/utils"
	"github.com/multihubandredye-prog/W-C-A_GO/validations"
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

	var mp3Source meowcaller.AudioSource
	var audioPath string
	var temporaryInput string
	defer func() {
		if temporaryInput != "" {
			_ = os.Remove(temporaryInput)
		}
	}()
	if request.Audio != nil || request.AudioURL != "" {
		var audioBytes []byte
		if request.Audio != nil {
			uploaded, errOpen := request.Audio.Open()
			if errOpen != nil {
				return response, pkgError.ValidationError("Failed to open uploaded audio")
			}
			audioBytes, err = io.ReadAll(io.LimitReader(uploaded, 20_000_001))
			_ = uploaded.Close()
			if err != nil || len(audioBytes) == 0 || len(audioBytes) > 20_000_000 {
				return response, pkgError.ValidationError("Audio is empty or exceeds 20 MB")
			}
		} else {
			audioBytes, _, err = utils.DownloadAudioFromURL(request.AudioURL)
			if err != nil {
				return response, pkgError.ValidationError("Failed to download audio")
			}
			if len(audioBytes) == 0 || len(audioBytes) > 20_000_000 {
				return response, pkgError.ValidationError("Audio is empty or exceeds 20 MB")
			}
		}
		f, e := os.CreateTemp(config.PathMedia, ".wca-call-input-*")
		if e != nil {
			return response, pkgError.InternalServerError("Unable to prepare temporary audio")
		}
		temporaryInput = f.Name()
		_, e = f.Write(audioBytes)
		closeErr := f.Close()
		if e != nil || closeErr != nil {
			return response, pkgError.InternalServerError("Unable to write temporary audio")
		}
		audioPath = temporaryInput
	} else if request.AudioPath != "" {
		if _, errStat := os.Stat(request.AudioPath); errStat == nil {
			audioPath = request.AudioPath
		} else {
			audioBytes, errBase64 := utils.Base64ToBytes(request.AudioPath)
			if errBase64 != nil {
				return response, pkgError.ValidationError("audio_path must be an existing file or valid Base64 audio")
			}
			if len(audioBytes) == 0 || len(audioBytes) > 20_000_000 {
				return response, pkgError.ValidationError("Audio is empty or exceeds 20 MB")
			}
			f, e := os.CreateTemp(config.PathMedia, ".wca-call-input-*")
			if e != nil {
				return response, pkgError.InternalServerError("Unable to prepare temporary audio")
			}
			temporaryInput = f.Name()
			_, e = f.Write(audioBytes)
			closeErr := f.Close()
			if e != nil || closeErr != nil {
				return response, pkgError.InternalServerError("Unable to write temporary audio")
			}
			audioPath = temporaryInput
		}
	}
	if audioPath != "" {
		info, e := os.Stat(audioPath)
		if e != nil || !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > 20_000_000 {
			return response, pkgError.ValidationError("Audio is empty, inaccessible or exceeds 20 MB")
		}
		var cleanup func()
		mp3Source, cleanup, e = openCallAudio(ctx, audioPath, config.PathMedia)
		if e != nil {
			return response, pkgError.ValidationError(e.Error())
		}
		defer cleanup()
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
