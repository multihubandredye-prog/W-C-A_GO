package whatsapp

import (
	"context"
	"testing"

	"github.com/multihubandredye-prog/W-C-A_GO/config"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func TestOutgoingWebhookDisabledSkipsMessage(t *testing.T) {
	previous := config.WhatsappWebhookIncludeOutgoing
	config.WhatsappWebhookIncludeOutgoing = false
	t.Cleanup(func() { config.WhatsappWebhookIncludeOutgoing = previous })

	evt := &events.Message{Info: types.MessageInfo{ID: "outgoing-test", MessageSource: types.MessageSource{IsFromMe: true}}}
	if err := forwardMessageToWebhook(context.Background(), nil, evt); err != nil {
		t.Fatalf("disabled outgoing webhook returned error: %v", err)
	}
}
