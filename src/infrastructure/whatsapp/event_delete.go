package whatsapp

import (
	"context"
	"time"

	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

// forwardDeleteToWebhook sends a delete event to webhook
func forwardDeleteToWebhook(ctx context.Context, evt *events.DeleteForMe, message *domainChatStorage.Message, deviceID string, client *whatsmeow.Client) error {
	payload, err := createDeletePayload(ctx, evt, message, deviceID, client)
	if err != nil {
		return err
	}

	return forwardPayloadToConfiguredWebhooks(ctx, payload, "message.deleted")
}

// createDeletePayload creates a webhook payload for delete events
func createDeletePayload(ctx context.Context, evt *events.DeleteForMe, message *domainChatStorage.Message, deviceID string, client *whatsmeow.Client) (map[string]any, error) {
	body := make(map[string]any)
	payload := make(map[string]any)

	payload["DeletedMessageID"] = evt.MessageID
	payload["Timestamp"] = time.Now().Format(time.RFC3339)
	payload["Type"] = "DeletedMessage"

	// Resolve sender JID (convert LID to phone number if needed)
	normalizedSenderJID := NormalizeJIDFromLID(ctx, evt.SenderJID, client)
	payload["From"] = normalizedSenderJID.ToNonAD().String()

	// Include original message information if available
	if message != nil {
		payload["ChatID"] = message.ChatJID
		payload["OriginalContent"] = message.Content
		payload["OriginalSender"] = message.Sender
		payload["OriginalTimestamp"] = message.Timestamp.Format(time.RFC3339)
		payload["WasFromMe"] = message.IsFromMe

		if message.MediaType != "" {
			payload["OriginalMediaType"] = message.MediaType
			payload["OriginalFileName"] = message.Filename
		}
	}

	body["Event"] = "message.deleted"
	if deviceID != "" {
		body["DeviceID"] = deviceID
	}
	body["Payload"] = payload

	return body, nil
}
