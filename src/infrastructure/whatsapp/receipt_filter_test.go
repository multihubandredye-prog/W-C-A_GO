package whatsapp

import (
	"github.com/multihubandredye-prog/W-C-A_GO/config"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"testing"
)

func TestReceiptWebhookFilters(t *testing.T) {
	oldD, oldR, oldS := config.WhatsappWebhookAckDelivered, config.WhatsappWebhookAckRead, config.WhatsappWebhookAckReadSelf
	defer func() {
		config.WhatsappWebhookAckDelivered, config.WhatsappWebhookAckRead, config.WhatsappWebhookAckReadSelf = oldD, oldR, oldS
	}()
	own := types.JID{User: "5511", Server: types.DefaultUserServer, Device: 1}
	lid := types.JID{User: "99", Server: types.HiddenUserServer}
	client := &whatsmeow.Client{Store: &store.Device{ID: &own, LID: lid}}
	remote := types.JID{User: "5522", Server: types.DefaultUserServer}
	cases := []struct {
		name                        string
		evt                         *events.Receipt
		delivered, read, self, want bool
	}{
		{"entrega ativa", &events.Receipt{Type: types.ReceiptTypeDelivered}, true, false, false, true},
		{"entrega desativada", &events.Receipt{Type: types.ReceiptTypeDelivered}, false, true, true, false},
		{"destinatario leu", &events.Receipt{Type: types.ReceiptTypeRead, MessageSource: types.MessageSource{Sender: remote}}, false, true, false, true},
		{"destinatario bloqueado", &events.Receipt{Type: types.ReceiptTypeRead, MessageSource: types.MessageSource{Sender: remote}}, true, false, true, false},
		{"propria leitura pelo flag", &events.Receipt{Type: types.ReceiptTypeRead, MessageSource: types.MessageSource{IsFromMe: true}}, false, false, true, true},
		{"propria leitura privada", &events.Receipt{Type: types.ReceiptTypeReadSelf}, false, true, false, false},
		{"propria leitura por jid", &events.Receipt{Type: types.ReceiptTypeRead, MessageSource: types.MessageSource{Sender: own}}, false, true, false, false},
		{"propria leitura por lid", &events.Receipt{Type: types.ReceiptTypeRead, MessageSource: types.MessageSource{Sender: lid}}, false, false, true, true},
		{"grupo remetente alheio", &events.Receipt{Type: types.ReceiptTypeRead, MessageSource: types.MessageSource{IsGroup: true, Sender: remote}}, false, true, false, true},
		{"outro tipo ignorado", &events.Receipt{Type: types.ReceiptTypeRetry}, true, true, true, false},
		{"nil", nil, true, true, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config.WhatsappWebhookAckDelivered, config.WhatsappWebhookAckRead, config.WhatsappWebhookAckReadSelf = tc.delivered, tc.read, tc.self
			if got := receiptWebhookEnabled(tc.evt, client); got != tc.want {
				t.Fatalf("got %t want %t", got, tc.want)
			}
		})
	}
}
