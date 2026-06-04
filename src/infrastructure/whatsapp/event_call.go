package whatsapp

import (
	"context"
	"errors"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

// handleCallOffer handles incoming call events and optionally auto-rejects them
func handleCallOffer(ctx context.Context, evt *events.CallOffer, chatStorageRepo domainChatStorage.IChatStorageRepository, deviceID string, client *whatsmeow.Client) {
	logrus.Infof("Incoming call from %s (CallID: %s)", evt.CallCreator.String(), evt.CallID)

	// Auto-reject call if configured
	autoRejected := false
	if config.WhatsappAutoRejectCall {
		rejectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := client.RejectCall(rejectCtx, evt.CallCreator, evt.CallID); err != nil {
			logrus.Errorf("Failed to reject call from %s: %v", evt.CallCreator.String(), err)
		} else {
			autoRejected = true
			logrus.Infof("Auto-rejected call from %s (CallID: %s)", evt.CallCreator.String(), evt.CallID)
		}
	}

	if chatStorageRepo != nil {
		if err := chatStorageRepo.CreateIncomingCallRecord(ctx, evt, autoRejected); err != nil {
			switch {
			case errors.Is(err, domainChatStorage.ErrMissingDeviceContext),
				errors.Is(err, domainChatStorage.ErrCallOfferMissingPeerJID):
				logrus.Warnf("Skipping incoming call persistence: %v", err)
			default:
				logrus.Errorf("Failed to persist incoming call: %v", err)
			}
		}
	}

	// Forward call event to webhook if configured
	if len(config.WhatsappWebhook) > 0 {
		go func(e *events.CallOffer, c *whatsmeow.Client, rejected bool) {
			webhookCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := forwardCallOfferToWebhook(webhookCtx, e, deviceID, c, rejected); err != nil {
				logrus.Errorf("Failed to forward call event to webhook: %v", err)
			}
		}(evt, client, autoRejected)
	}
}

// createCallOfferPayload creates a webhook payload for incoming call events
func createCallOfferPayload(ctx context.Context, evt *events.CallOffer, deviceID string, client *whatsmeow.Client, autoRejected bool) map[string]any {
	body := make(map[string]any)
	payload := make(map[string]any)

	// Determine call type
	isVideo := false
	if evt.Data != nil {
		for _, child := range evt.Data.GetChildren() {
			if child.Tag == "video" {
				isVideo = true
				break
			}
		}
	}

	// Add call details
	payload["CallID"] = evt.CallID
	payload["From"] = NormalizeJIDFromLID(ctx, evt.CallCreator, client).ToNonAD().String()
	payload["AutoRejected"] = autoRejected
	payload["IsGroup"] = !evt.GroupJID.IsEmpty()
	payload["CallStatus"] = "incoming"

	if isVideo {
		payload["CallType"] = "video"
		payload["Type"] = "VideoCallOfferMessage"
	} else {
		payload["CallType"] = "audio"
		payload["Type"] = "AudioCallOfferMessage"
	}

	// Add group JID if this is a group call
	if !evt.GroupJID.IsEmpty() {
		payload["GroupJID"] = evt.GroupJID.ToNonAD().String()
	}

	// Wrap in body structure
	body["Event"] = "call.offer"
	body["Timestamp"] = evt.Timestamp.Format(time.RFC3339)
	if deviceID != "" {
		body["DeviceID"] = deviceID
	}
	body["Payload"] = payload

	return body
}

// forwardCallOfferToWebhook forwards incoming call events to the configured webhook URLs
func forwardCallOfferToWebhook(ctx context.Context, evt *events.CallOffer, deviceID string, client *whatsmeow.Client, autoRejected bool) error {
	payload := createCallOfferPayload(ctx, evt, deviceID, client, autoRejected)
	return forwardPayloadToConfiguredWebhooks(ctx, payload, "call.offer")
}

// handleCallTerminate handles incoming call termination events
func handleCallTerminate(ctx context.Context, evt *events.CallTerminate, deviceID string, client *whatsmeow.Client) {
	logrus.Infof("Call terminated: %s", evt.CallID)

	// Forward call terminate event to webhook
	if len(config.WhatsappWebhook) > 0 {
		
		// Determine shutdown causer
		shutdownCauser := evt.Reason
		if evt.Reason == "rejected_elsewhere" {
			shutdownCauser = "MySelf"
		} else if evt.Reason == "" {
			shutdownCauser = "SenderHungUp"
		}

		payload := map[string]any{
			"CallID":       evt.CallID,
			"Type":         "CallTerminateMessage",
			"CallStatus":   "terminated",
			"TerminatedBy": NormalizeJIDFromLID(ctx, evt.From, client).ToNonAD().String(),
			"ShutdownCauser": shutdownCauser,
		}
		
		body := map[string]any{
			"Event":     "call.terminate",
			"Timestamp": time.Now().Format(time.RFC3339),
			"Payload":   payload,
		}
		if deviceID != "" {
			body["DeviceID"] = deviceID
		}

		_ = forwardPayloadToConfiguredWebhooks(ctx, body, "call.terminate")
	}
}
