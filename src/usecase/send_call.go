package usecase

import (
	"context"
	"fmt"
	"time"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/validations"
	"github.com/purpshell/meowcaller"
	"github.com/sirupsen/logrus"
)

// SendCall initiates a WhatsApp VoIP audio call to the recipient.
// If AudioPath is provided, the audio file is streamed when the peer answers.
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

	caller := meowcaller.NewClient(client)
	call, err := caller.Call(ctx, recipient.String())
	if err != nil {
		return response, pkgError.InternalServerError(fmt.Sprintf("Failed to initiate call: %v", err))
	}

	durationSec := 15
	if request.Duration != nil && *request.Duration > 0 {
		durationSec = *request.Duration
	}

	if request.AudioPath != "" {
		if mp3, audioErr := meowcaller.MP3File(request.AudioPath); audioErr == nil {
			call.Play(mp3)
		} else {
			logrus.Warnf("Failed to open MP3 file %s for call: %v", request.AudioPath, audioErr)
		}
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
