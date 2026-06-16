package whatsapp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/pollstore"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func getReceiptTypeDescription(evt types.ReceiptType) string {
	switch evt {
	case types.ReceiptTypeDelivered:
		return "A mensagem foi entregue ao dispositivo do destinatário."
	case types.ReceiptTypeSender:
		return "A mensagem foi entregue a outro dos seus dispositivos."
	case types.ReceiptTypeRetry:
		return "A mensagem foi entregue, mas houve falha na descriptografia."
	case types.ReceiptTypeRead:
		return "o usuário abriu a conversa e visualizou a mensagem"
	case types.ReceiptTypeReadSelf:
		return "A mensagem foi lida por você em outro dispositivo."
	case types.ReceiptTypePlayed:
		return "A mensagem de mídia (áudio/vídeo) foi reproduzida ou visualizada."
	case types.ReceiptTypePlayedSelf:
		return "A mensagem de mídia foi reproduzida ou visualizada por você em outro dispositivo."
	default:
		return "Tipo de confirmação desconhecido."
	}
}

// createReceiptPayload creates a webhook payload for message acknowledgement (receipt) events
func createReceiptPayload(ctx context.Context, evt *events.Receipt, deviceID string, client *whatsmeow.Client) map[string]any {
	body := make(map[string]any)
	payload := make(map[string]any)

	// Add message IDs
	if len(evt.MessageIDs) > 0 {
		payload["IDs"] = evt.MessageIDs
	}

	// Add ChatID
	payload["ChatID"] = evt.Chat.ToNonAD().String()
	payload["IsGroup"] = evt.Chat.Server == types.GroupServer

	// Build from/from_lid fields from sender
	senderJID := evt.Sender

	if senderJID.Server == "lid" {
		payload["FromLid"] = senderJID.ToNonAD().String()
	}

	// Resolve sender JID (convert LID to phone number if needed)
	normalizedSenderJID := NormalizeJIDFromLID(ctx, senderJID, client)
	payload["From"] = normalizedSenderJID.ToNonAD().String()

	// Receipt type
	if evt.Type == types.ReceiptTypeDelivered {
		payload["ReceiptType"] = "delivered"
	} else {
		payload["ReceiptType"] = string(evt.Type)
	}
	payload["ReceiptTypeDescription"] = getReceiptTypeDescription(evt.Type)
	payload["Type"] = "ReceiptMessage"

	// Enrich with poll data if available
	if len(evt.MessageIDs) > 0 {
		if pollData, found := pollstore.DefaultPollStore.GetPoll(string(evt.MessageIDs[0])); found {
			payload["Poll"] = pollData
		}
	}

	// Wrap in body structure
	body["Event"] = "message.ack"
	body["Timestamp"] = evt.Timestamp.Format(time.RFC3339)
	if deviceID != "" {
		body["DeviceID"] = deviceID
	}
	body["Payload"] = payload

	return body
}

var (
	receiptDebounceMu sync.Mutex
	receiptDebouncers = make(map[string]*receiptDebouncer)
)

type receiptDebouncer struct {
	evt        *events.Receipt
	timer      *time.Timer
	deviceID   string
	client     *whatsmeow.Client
	messageIDs map[string]struct{}
}

// forwardReceiptToWebhook forwards message acknowledgement events to the configured webhook URLs.
//
// IMPORTANT: We only forward receipts from the primary device (Device == 0).
// WhatsApp sends separate receipt events for each linked device (phone, web, desktop, etc.)
// of a user. For example, if a user has 3 devices, you would receive 3 "delivered" receipts
// for the same message. To avoid duplicate webhooks and simplify downstream processing,
// we only send the receipt from the primary device (Device == 0).
//
// If you need receipts from all devices in the future, remove the Device == 0 check below.
func forwardReceiptToWebhook(ctx context.Context, evt *events.Receipt, deviceID string, client *whatsmeow.Client) error {
	// Only forward receipts from the primary device to avoid duplicates.
	// See function comment above for detailed explanation.
	if evt.Sender.Device != 0 {
		logrus.Debugf("Skipping receipt webhook for linked device %d (only primary device receipts are forwarded)", evt.Sender.Device)
		return nil
	}

	// Debounce receipts to avoid flooding the webhook when many messages are read at once (e.g. opening a group)
	// We group by DeviceID, ChatID and ReceiptType within a 500ms window.
	key := fmt.Sprintf("%s-%s-%s", deviceID, evt.Chat.String(), evt.Type)

	receiptDebounceMu.Lock()
	defer receiptDebounceMu.Unlock()

	if d, ok := receiptDebouncers[key]; ok {
		// Add new IDs to the map for uniqueness
		for _, id := range evt.MessageIDs {
			d.messageIDs[id] = struct{}{}
		}
		// Update to the latest timestamp
		d.evt.Timestamp = evt.Timestamp
		// Reset the timer
		d.timer.Stop()
		d.timer.Reset(500 * time.Millisecond)
		return nil
	}

	// Create new debouncer
	ids := make(map[string]struct{})
	for _, id := range evt.MessageIDs {
		ids[id] = struct{}{}
	}

	d := &receiptDebouncer{
		evt:        evt,
		deviceID:   deviceID,
		client:     client,
		messageIDs: ids,
	}
	d.timer = time.AfterFunc(500*time.Millisecond, func() {
		receiptDebounceMu.Lock()
		delete(receiptDebouncers, key)

		// Convert IDs map back to slice
		finalIDs := make([]string, 0, len(d.messageIDs))
		for id := range d.messageIDs {
			finalIDs = append(finalIDs, id)
		}
		d.evt.MessageIDs = finalIDs
		currentEvt := d.evt
		receiptDebounceMu.Unlock()

		// Use a fresh background context as the original might have been cancelled
		webhookCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		payload := createReceiptPayload(webhookCtx, currentEvt, d.deviceID, d.client)
		if err := forwardPayloadToConfiguredWebhooks(webhookCtx, payload, "message.ack"); err != nil {
			logrus.Errorf("Failed to forward batched ack event to webhook: %v", err)
		}
	})
	receiptDebouncers[key] = d

	return nil
}

