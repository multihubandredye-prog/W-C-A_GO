package whatsapp

import (
	"context"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// handleChatPresence handles incoming chat presence (typing notification) events.
// These events are emitted when a user starts or stops typing in a chat.
// Note: WhatsApp only sends these updates when the client is marked as online.
func handleChatPresence(ctx context.Context, evt *events.ChatPresence, deviceID string, client *whatsmeow.Client) {
	if evt.State == types.ChatPresenceComposing {
		if evt.Media == types.ChatPresenceMediaAudio {
			log.Infof("%s is recording audio in %s", evt.Sender.ToNonAD(), evt.Chat.ToNonAD())
		} else {
			log.Infof("%s is typing in %s", evt.Sender.ToNonAD(), evt.Chat.ToNonAD())
		}
	} else {
		log.Infof("%s stopped typing in %s", evt.Sender.ToNonAD(), evt.Chat.ToNonAD())
	}

	// Forward chat presence event to webhook if configured
	if len(config.WhatsappWebhook) > 0 {
		go func(e *events.ChatPresence, c *whatsmeow.Client) {
			webhookCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := forwardChatPresenceToWebhook(webhookCtx, e, deviceID, c); err != nil {
				logrus.Errorf("Failed to forward chat_presence event to webhook: %v", err)
			}
		}(evt, client)
	}
}

// createChatPresencePayload creates a webhook payload for chat presence (typing) events.
func createChatPresencePayload(ctx context.Context, evt *events.ChatPresence, deviceID string, client *whatsmeow.Client) map[string]any {
	body := make(map[string]any)
	payload := make(map[string]any)

	// Resolve sender JID (convert LID to phone number if needed)
	senderJID := evt.Sender
	if senderJID.Server == "lid" {
		payload["FromLid"] = senderJID.ToNonAD().String()
	}
	normalizedSenderJID := NormalizeJIDFromLID(ctx, senderJID, client)
	payload["From"] = normalizedSenderJID.ToNonAD().String()

	// Chat where the presence event occurred
	payload["ChatID"] = evt.Chat.ToNonAD().String()

	// Typing state: "composing" or "paused"
	payload["State"] = string(evt.State)

	// Media type: "" (text) or "audio" (recording voice message)
	payload["Media"] = string(evt.Media)

	// Whether this is a group chat
	payload["IsGroup"] = evt.IsGroup
	payload["Type"] = "ChatPresenceMessage"

	// Wrap in body structure
	body["Event"] = "chat_presence"
	body["Timestamp"] = time.Now().Format(time.RFC3339)
	if deviceID != "" {
		body["DeviceID"] = deviceID
	}
	body["Payload"] = payload

	return body
}

// forwardChatPresenceToWebhook forwards chat presence events to the configured webhook URLs.
func forwardChatPresenceToWebhook(ctx context.Context, evt *events.ChatPresence, deviceID string, client *whatsmeow.Client) error {
	payload := createChatPresencePayload(ctx, evt, deviceID, client)
	return forwardPayloadToConfiguredWebhooks(ctx, payload, "chat_presence")
}
