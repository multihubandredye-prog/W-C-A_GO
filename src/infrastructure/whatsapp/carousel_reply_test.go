package whatsapp

import (
	"context"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func TestCarouselResponseClassification(t *testing.T) {
	for _, tc := range []struct {
		id  string
		typ string
		idx int
	}{
		{"carousel_1", "CarouselResponseMessage", 1}, {"carousel_10", "CarouselResponseMessage", 10},
		{"carousel_0", "ButtonsResponseMessage", 0}, {"carousel_11", "ButtonsResponseMessage", 0},
		{"carousel_01", "ButtonsResponseMessage", 0}, {"normal", "ButtonsResponseMessage", 0},
	} {
		msg := buttonReply(tc.id, "Selecionar")
		assert.Equal(t, tc.typ, getMessagePascalType(msg))
		payload := map[string]any{}
		buildOtherMessageTypes(msg, payload)
		reply := payload["InteractiveReply"].(map[string]any)
		if tc.idx == 0 {
			assert.Equal(t, "buttons", reply["Type"])
			assert.NotContains(t, payload, "CarouselReply")
			continue
		}
		assert.Equal(t, "carousel", reply["Type"])
		assert.Equal(t, tc.id, reply["SelectedID"])
		assert.Equal(t, tc.idx, reply["CardIndex"])
		c, ok := payload["CarouselReply"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, tc.idx, c["CardIndex"])
	}
}
func TestCarouselLegacyAndContext(t *testing.T) {
	msgs := []*waE2E.Message{
		{ButtonsResponseMessage: &waE2E.ButtonsResponseMessage{SelectedButtonID: proto.String("carousel_2")}},
		{TemplateButtonReplyMessage: &waE2E.TemplateButtonReplyMessage{SelectedID: proto.String("carousel_2")}},
	}
	for _, msg := range msgs {
		assert.Equal(t, "CarouselResponseMessage", getMessagePascalType(msg))
		p := map[string]any{}
		buildOtherMessageTypes(msg, p)
		assert.Equal(t, 2, p["CarouselReply"].(map[string]any)["CardIndex"])
	}
	msg := buttonReply("carousel_3", "Selecionar")
	msg.InteractiveResponseMessage.ContextInfo = &waE2E.ContextInfo{StanzaID: proto.String("origin-id")}
	p := map[string]any{}
	buildOtherMessageTypes(msg, p)
	assert.Equal(t, "origin-id", p["CarouselReply"].(map[string]any)["RepliedToID"])
}

func TestCarouselFullWebhookEvent(t *testing.T) {
	msg := buttonReply("carousel_2", "Selecionar")
	msg.InteractiveResponseMessage.ContextInfo = &waE2E.ContextInfo{StanzaID: proto.String("sent-id"), QuotedMessage: &waE2E.Message{Conversation: proto.String("Texto anterior")}}
	evt := &events.Message{Message: msg, Info: types.MessageInfo{MessageSource: types.MessageSource{Chat: types.NewJID("5511999999999", types.DefaultUserServer), Sender: types.NewJID("5511999999999", types.DefaultUserServer)}}}
	event, payload, err := buildEventPayload(context.Background(), nil, evt)
	require.NoError(t, err)
	assert.Equal(t, "message", event)
	assert.Equal(t, "CarouselResponseMessage", payload["Type"])
	reply, ok := payload["CarouselReply"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, 2, reply["CardIndex"])
	assert.Equal(t, "sent-id", reply["RepliedToID"])
}
