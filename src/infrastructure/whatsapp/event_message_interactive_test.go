package whatsapp

import (
	"testing"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// buttonReply builds the message a client sends when a NativeFlow button is tapped.
func buttonReply(id, text string) *waE2E.Message {
	return &waE2E.Message{
		InteractiveResponseMessage: &waE2E.InteractiveResponseMessage{
			Body: &waE2E.InteractiveResponseMessage_Body{Text: proto.String("Já paguei")},
			InteractiveResponseMessage: &waE2E.InteractiveResponseMessage_NativeFlowResponseMessage_{
				NativeFlowResponseMessage: &waE2E.InteractiveResponseMessage_NativeFlowResponseMessage{
					Name:       proto.String("quick_reply"),
					ParamsJSON: proto.String(`{"display_text":"` + text + `","id":"` + id + `"}`),
				},
			},
		},
	}
}

// TestUnwrapMessageReachesInteractiveReply covers the bug where a button tap
// wrapped in DeviceSentMessage/EphemeralMessage never reached the extractor,
// leaving the webhook without the selected option.
func TestUnwrapMessageReachesInteractiveReply(t *testing.T) {
	inner := buttonReply("ja_paguei", "Já paguei")

	tests := []struct {
		name string
		msg  *waE2E.Message
	}{
		{
			name: "bare interactive response",
			msg:  inner,
		},
		{
			name: "wrapped in DeviceSentMessage",
			msg: &waE2E.Message{
				DeviceSentMessage: &waE2E.DeviceSentMessage{Message: inner},
			},
		},
		{
			name: "wrapped in EphemeralMessage",
			msg: &waE2E.Message{
				EphemeralMessage: &waE2E.FutureProofMessage{Message: inner},
			},
		},
		{
			name: "DeviceSent inside Ephemeral",
			msg: &waE2E.Message{
				EphemeralMessage: &waE2E.FutureProofMessage{
					Message: &waE2E.Message{
						DeviceSentMessage: &waE2E.DeviceSentMessage{Message: inner},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unwrapped := utils.UnwrapMessage(tt.msg)
			require.NotNil(t, unwrapped.GetInteractiveResponseMessage(),
				"unwrap must expose the interactive response")

			payload := map[string]any{}
			buildInteractiveReplyFields(unwrapped, payload)

			reply, ok := payload["InteractiveReply"].(map[string]any)
			require.True(t, ok, "InteractiveReply must be present")
			assert.Equal(t, "buttons", reply["Type"])
			assert.Equal(t, "ja_paguei", reply["SelectedID"])
			assert.Equal(t, "Já paguei", reply["SelectedText"])
		})
	}
}

// TestGetMessagePascalTypeInteractive covers the bug where interactive traffic
// was reported as "Unknown", which surfaced in Tasker as unknown_message.
func TestGetMessagePascalTypeInteractive(t *testing.T) {
	listMsg := &waE2E.ListMessage{
		Description: proto.String("Escolha"),
		ButtonText:  proto.String("Ver"),
		Sections: []*waE2E.ListMessage_Section{{
			Rows: []*waE2E.ListMessage_Row{{
				RowID: proto.String("a"), Title: proto.String("A"),
			}},
		}},
	}

	tests := []struct {
		name     string
		msg      *waE2E.Message
		expected string
	}{
		{
			name:     "button tap",
			msg:      buttonReply("ja_paguei", "Já paguei"),
			expected: "ButtonsResponseMessage",
		},
		{
			name: "list selection",
			msg: &waE2E.Message{
				ListResponseMessage: &waE2E.ListResponseMessage{
					Title: proto.String("Margherita"),
					SingleSelectReply: &waE2E.ListResponseMessage_SingleSelectReply{
						SelectedRowID: proto.String("pizza_marg"),
					},
				},
			},
			expected: "ListResponseMessage",
		},
		{
			name: "outgoing buttons",
			msg: &waE2E.Message{
				InteractiveMessage: &waE2E.InteractiveMessage{
					Body: &waE2E.InteractiveMessage_Body{Text: proto.String("Pague o PIX")},
				},
			},
			expected: "ButtonsMessage",
		},
		{
			name:     "outgoing list",
			msg:      &waE2E.Message{ListMessage: listMsg},
			expected: "ListMessage",
		},
		{
			name: "outgoing list wrapped in DocumentWithCaption",
			msg: utils.UnwrapMessage(&waE2E.Message{
				DocumentWithCaptionMessage: &waE2E.FutureProofMessage{
					Message: &waE2E.Message{ListMessage: listMsg},
				},
			}),
			expected: "ListMessage",
		},
		{
			name:     "plain text stays Message",
			msg:      &waE2E.Message{Conversation: proto.String("oi")},
			expected: "Message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getMessagePascalType(tt.msg))
		})
	}
}

// TestBuildSentButtonsPayload checks the options we offered are echoed back,
// so the webhook shows the buttons that were sent.
func TestBuildSentButtonsPayload(t *testing.T) {
	interactive := &waE2E.InteractiveMessage{
		Body:   &waE2E.InteractiveMessage_Body{Text: proto.String("Pague via PIX")},
		Footer: &waE2E.InteractiveMessage_Footer{Text: proto.String("Clínica Bem Estar")},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
					{
						Name:             proto.String("cta_copy"),
						ButtonParamsJSON: proto.String(`{"display_text":"Copiar PIX","copy_code":"558184752564"}`),
					},
					{
						Name:             proto.String("quick_reply"),
						ButtonParamsJSON: proto.String(`{"display_text":"Já paguei","id":"ja_paguei"}`),
					},
				},
			},
		},
	}

	sent := buildSentButtonsPayload(interactive)
	assert.Equal(t, "Pague via PIX", sent["Body"])
	assert.Equal(t, "Clínica Bem Estar", sent["Footer"])
	assert.Equal(t, 2, sent["ButtonsCount"])

	buttons, ok := sent["Buttons"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, buttons, 2)

	assert.Equal(t, "Copiar PIX", buttons[0]["Title"])
	assert.Equal(t, "558184752564", buttons[0]["CopyCode"])
	assert.Equal(t, "Já paguei", buttons[1]["Title"])
	assert.Equal(t, "ja_paguei", buttons[1]["ID"])
}

// TestBuildSentListPayload checks sections and rows are flattened for the webhook.
func TestBuildSentListPayload(t *testing.T) {
	list := &waE2E.ListMessage{
		Description: proto.String("Escolha seu pedido"),
		ButtonText:  proto.String("Ver cardápio"),
		Title:       proto.String("Pizzaria"),
		Sections: []*waE2E.ListMessage_Section{{
			Title: proto.String("Pizzas"),
			Rows: []*waE2E.ListMessage_Row{
				{RowID: proto.String("pizza_marg"), Title: proto.String("Margherita"), Description: proto.String("R$ 45")},
				{RowID: proto.String("pizza_cala"), Title: proto.String("Calabresa")},
			},
		}},
	}

	sent := buildSentListPayload(list)
	assert.Equal(t, "Escolha seu pedido", sent["Description"])
	assert.Equal(t, "Ver cardápio", sent["ButtonText"])
	assert.Equal(t, 2, sent["RowsCount"])

	sections, ok := sent["Sections"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, sections, 1)

	rows, ok := sections[0]["Rows"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, rows, 2)
	assert.Equal(t, "pizza_marg", rows[0]["RowID"])
	assert.Equal(t, "R$ 45", rows[0]["Description"])
}

// TestUnwrapKeepsRealDocuments guards against unwrapping a genuine document
// message, which would break document handling.
func TestUnwrapKeepsRealDocuments(t *testing.T) {
	msg := &waE2E.Message{
		DocumentWithCaptionMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				DocumentMessage: &waE2E.DocumentMessage{
					FileName: proto.String("contrato.pdf"),
				},
			},
		},
	}

	unwrapped := utils.UnwrapMessage(msg)
	assert.NotNil(t, unwrapped.GetDocumentWithCaptionMessage(),
		"a real document must not be unwrapped")
}

// TestFindInteractiveReplyDigsDeep covers replies nested deeper than the
// generic unwrap reaches, which arrived at the webhook as an empty payload
// typed "Unknown" on LID-migrated accounts.
func TestFindInteractiveReplyDigsDeep(t *testing.T) {
	inner := buttonReply("ja_paguei", "Já paguei")

	tests := []struct {
		name  string
		msg   *waE2E.Message
		found bool
	}{
		{
			name:  "bare reply",
			msg:   inner,
			found: true,
		},
		{
			name: "inside DocumentWithCaption",
			msg: &waE2E.Message{
				DocumentWithCaptionMessage: &waE2E.FutureProofMessage{Message: inner},
			},
			found: true,
		},
		{
			name: "DeviceSent inside DocumentWithCaption",
			msg: &waE2E.Message{
				DocumentWithCaptionMessage: &waE2E.FutureProofMessage{
					Message: &waE2E.Message{
						DeviceSentMessage: &waE2E.DeviceSentMessage{Message: inner},
					},
				},
			},
			found: true,
		},
		{
			name: "three levels deep",
			msg: &waE2E.Message{
				EphemeralMessage: &waE2E.FutureProofMessage{
					Message: &waE2E.Message{
						ViewOnceMessage: &waE2E.FutureProofMessage{
							Message: &waE2E.Message{
								DeviceSentMessage: &waE2E.DeviceSentMessage{Message: inner},
							},
						},
					},
				},
			},
			found: true,
		},
		{
			name:  "plain text has no reply",
			msg:   &waE2E.Message{Conversation: proto.String("oi")},
			found: false,
		},
		{
			name: "real document has no reply",
			msg: &waE2E.Message{
				DocumentWithCaptionMessage: &waE2E.FutureProofMessage{
					Message: &waE2E.Message{
						DocumentMessage: &waE2E.DocumentMessage{FileName: proto.String("a.pdf")},
					},
				},
			},
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := utils.FindInteractiveReply(tt.msg)
			if !tt.found {
				assert.Nil(t, found)
				return
			}

			require.NotNil(t, found, "the reply must be located at any depth")
			assert.Equal(t, "ButtonsResponseMessage", getMessagePascalType(found))

			payload := map[string]any{}
			buildInteractiveReplyFields(found, payload)
			reply, ok := payload["InteractiveReply"].(map[string]any)
			require.True(t, ok)
			assert.Equal(t, "ja_paguei", reply["SelectedID"])
		})
	}
}

// TestIsEmptyMessage guards the RawMessage fallback trigger.
func TestIsEmptyMessage(t *testing.T) {
	assert.True(t, isEmptyMessage(nil))
	assert.True(t, isEmptyMessage(&waE2E.Message{}))

	// MessageContextInfo alone still counts as empty: it carries no content.
	assert.True(t, isEmptyMessage(&waE2E.Message{
		MessageContextInfo: &waE2E.MessageContextInfo{},
	}))

	assert.False(t, isEmptyMessage(&waE2E.Message{Conversation: proto.String("oi")}))
	assert.False(t, isEmptyMessage(buttonReply("ja_paguei", "Já paguei")))
}

// TestRecoverReplyFromRawMessage reproduces the reported failure: a button tap
// from a LID account arrives with an emptied Message while the NativeFlow
// response is still present in RawMessage.
func TestRecoverReplyFromRawMessage(t *testing.T) {
	raw := &waE2E.Message{
		DeviceSentMessage: &waE2E.DeviceSentMessage{
			Message: buttonReply("ja_paguei", "Já paguei"),
		},
	}

	// The processed message is empty, so the type would be Unknown.
	emptied := &waE2E.Message{MessageContextInfo: &waE2E.MessageContextInfo{}}
	require.True(t, isEmptyMessage(emptied))
	assert.Equal(t, "Unknown", getMessagePascalType(emptied))

	// The fallback must locate the reply in RawMessage.
	recovered := utils.FindInteractiveReply(raw)
	require.NotNil(t, recovered, "reply must be recoverable from RawMessage")
	assert.Equal(t, "ButtonsResponseMessage", getMessagePascalType(recovered))

	payload := map[string]any{}
	buildInteractiveReplyFields(recovered, payload)
	reply, ok := payload["InteractiveReply"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "ja_paguei", reply["SelectedID"])
	assert.Equal(t, "Já paguei", reply["SelectedText"])
}

// TestTemplateButtonReply reproduces the payload captured in production:
// tapping a NativeFlow button produced a templateButtonReplyMessage, a shape
// the event builder did not recognise, so the webhook reported Unknown with
// no selection even though the data was fully present and unencrypted.
func TestTemplateButtonReply(t *testing.T) {
	msg := &waE2E.Message{
		TemplateButtonReplyMessage: &waE2E.TemplateButtonReplyMessage{
			SelectedID:          proto.String("ja_paguei"),
			SelectedDisplayText: proto.String("Já paguei"),
			SelectedIndex:       proto.Uint32(1),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID: proto.String("3EB02C095FB5F81D2023AE"),
			},
		},
		MessageContextInfo: &waE2E.MessageContextInfo{},
	}

	assert.Equal(t, "ButtonsResponseMessage", getMessagePascalType(msg),
		"a template reply must not fall through to Unknown")

	payload := map[string]any{}
	buildInteractiveReplyFields(msg, payload)

	reply, ok := payload["InteractiveReply"].(map[string]any)
	require.True(t, ok, "InteractiveReply must be populated")
	assert.Equal(t, "buttons", reply["Type"])
	assert.Equal(t, "ja_paguei", reply["SelectedID"])
	assert.Equal(t, "Já paguei", reply["SelectedText"])
	assert.Equal(t, "3EB02C095FB5F81D2023AE", reply["RepliedToID"])
}

// TestTemplateButtonReplyWrapped ensures the same reply is still found when it
// arrives nested, as happens on LID-migrated accounts.
func TestTemplateButtonReplyWrapped(t *testing.T) {
	inner := &waE2E.Message{
		TemplateButtonReplyMessage: &waE2E.TemplateButtonReplyMessage{
			SelectedID:          proto.String("enviar_comprovante"),
			SelectedDisplayText: proto.String("Comprovante"),
		},
	}

	wrapped := &waE2E.Message{
		EphemeralMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				DeviceSentMessage: &waE2E.DeviceSentMessage{Message: inner},
			},
		},
	}

	found := utils.FindInteractiveReply(wrapped)
	require.NotNil(t, found)
	assert.Equal(t, "ButtonsResponseMessage", getMessagePascalType(found))

	payload := map[string]any{}
	buildInteractiveReplyFields(found, payload)
	reply, ok := payload["InteractiveReply"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "enviar_comprovante", reply["SelectedID"])
}
