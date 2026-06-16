package whatsapp

import (
	"context"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

// handleNewsletterJoin handles when you join/subscribe to a newsletter
func handleNewsletterJoin(ctx context.Context, evt *events.NewsletterJoin, deviceID string, client *whatsmeow.Client) {
	log.Infof("Joined newsletter %s", evt.ID)

	if len(config.WhatsappWebhook) > 0 {
		go func(e *events.NewsletterJoin) {
			webhookCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := forwardNewsletterJoinToWebhook(webhookCtx, e, deviceID); err != nil {
				logrus.Errorf("Failed to forward newsletter join to webhook: %v", err)
			}
		}(evt)
	}
}

// handleNewsletterLeave handles when you leave/unsubscribe from a newsletter
func handleNewsletterLeave(ctx context.Context, evt *events.NewsletterLeave, deviceID string, client *whatsmeow.Client) {
	log.Infof("Left newsletter %s (role: %s)", evt.ID, evt.Role)

	if len(config.WhatsappWebhook) > 0 {
		go func(e *events.NewsletterLeave) {
			webhookCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := forwardNewsletterLeaveToWebhook(webhookCtx, e, deviceID); err != nil {
				logrus.Errorf("Failed to forward newsletter leave to webhook: %v", err)
			}
		}(evt)
	}
}

// handleNewsletterLiveUpdate handles new messages in newsletters
func handleNewsletterLiveUpdate(ctx context.Context, evt *events.NewsletterLiveUpdate, deviceID string, client *whatsmeow.Client) {
	log.Infof("Newsletter %s: %d new message(s)", evt.JID, len(evt.Messages))

	if len(config.WhatsappWebhook) > 0 {
		go func(e *events.NewsletterLiveUpdate) {
			webhookCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := forwardNewsletterLiveUpdateToWebhook(webhookCtx, e, deviceID); err != nil {
				logrus.Errorf("Failed to forward newsletter live update to webhook: %v", err)
			}
		}(evt)
	}
}

// handleNewsletterMuteChange handles newsletter mute setting changes
func handleNewsletterMuteChange(ctx context.Context, evt *events.NewsletterMuteChange, deviceID string, client *whatsmeow.Client) {
	log.Infof("Newsletter %s mute changed to: %s", evt.ID, evt.Mute)

	if len(config.WhatsappWebhook) > 0 {
		go func(e *events.NewsletterMuteChange) {
			webhookCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := forwardNewsletterMuteChangeToWebhook(webhookCtx, e, deviceID); err != nil {
				logrus.Errorf("Failed to forward newsletter mute change to webhook: %v", err)
			}
		}(evt)
	}
}

// Webhook forwarding functions

func forwardNewsletterJoinToWebhook(ctx context.Context, evt *events.NewsletterJoin, deviceID string) error {
	payload := map[string]any{
		"NewsletterID": evt.ID.String(),
		"Type":         "NewsletterJoinMessage",
	}
	if evt.ThreadMeta.Name.Text != "" {
		payload["Name"] = evt.ThreadMeta.Name.Text
	}
	if evt.ThreadMeta.Description.Text != "" {
		payload["Description"] = evt.ThreadMeta.Description.Text
	}

	body := map[string]any{
		"Event":     "newsletter.joined",
		"Payload":   payload,
		"Timestamp": time.Now().Format(time.RFC3339),
	}
	if deviceID != "" {
		body["DeviceID"] = deviceID
	}

	return forwardPayloadToConfiguredWebhooks(ctx, body, "newsletter.joined")
}

func forwardNewsletterLeaveToWebhook(ctx context.Context, evt *events.NewsletterLeave, deviceID string) error {
	payload := map[string]any{
		"NewsletterID": evt.ID.String(),
		"Role":         string(evt.Role),
		"Type":         "NewsletterLeaveMessage",
	}

	body := map[string]any{
		"Event":     "newsletter.left",
		"Payload":   payload,
		"Timestamp": time.Now().Format(time.RFC3339),
	}
	if deviceID != "" {
		body["DeviceID"] = deviceID
	}

	return forwardPayloadToConfiguredWebhooks(ctx, body, "newsletter.left")
}

func forwardNewsletterLiveUpdateToWebhook(ctx context.Context, evt *events.NewsletterLiveUpdate, deviceID string) error {
	messages := make([]map[string]any, 0, len(evt.Messages))
	for _, msg := range evt.Messages {
		m := map[string]any{
			"ServerID":  msg.MessageServerID,
			"MessageID": string(msg.MessageID),
			"Type":      msg.Type,
			"Timestamp": msg.Timestamp.Format(time.RFC3339),
		}
		if msg.ViewsCount > 0 {
			m["ViewsCount"] = msg.ViewsCount
		}
		if len(msg.ReactionCounts) > 0 {
			m["ReactionCounts"] = msg.ReactionCounts
		}
		messages = append(messages, m)
	}

	payload := map[string]any{
		"NewsletterID": evt.JID.String(),
		"Messages":     messages,
		"Type":         "NewsletterUpdateMessage",
	}

	body := map[string]any{
		"Event":     "newsletter.message",
		"Payload":   payload,
		"Timestamp": evt.Time.Format(time.RFC3339),
	}
	if deviceID != "" {
		body["DeviceID"] = deviceID
	}

	return forwardPayloadToConfiguredWebhooks(ctx, body, "newsletter.message")
}

func forwardNewsletterMuteChangeToWebhook(ctx context.Context, evt *events.NewsletterMuteChange, deviceID string) error {
	payload := map[string]any{
		"NewsletterID": evt.ID.String(),
		"Mute":         string(evt.Mute),
		"Type":         "NewsletterMuteMessage",
	}

	body := map[string]any{
		"Event":     "newsletter.mute",
		"Payload":   payload,
		"Timestamp": time.Now().Format(time.RFC3339),
	}
	if deviceID != "" {
		body["DeviceID"] = deviceID
	}

	return forwardPayloadToConfiguredWebhooks(ctx, body, "newsletter.mute")
}
