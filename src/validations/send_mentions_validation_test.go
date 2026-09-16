package validations

import (
	"context"
	"testing"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/stretchr/testify/assert"
)

// TestValidateSendMessage_WithMentions covers the ghost mentions field:
// phone numbers or JIDs that are notified without a visible "@phone" in the text,
// plus the special "@everyone" keyword that expands to every group participant.
func TestValidateSendMessage_WithMentions(t *testing.T) {
	const group = "120363000000000000@g.us"

	type args struct {
		request domainSend.MessageRequest
	}
	tests := []struct {
		name string
		args args
		err  any
	}{
		{
			name: "should success with a single ghost mention",
			args: args{request: domainSend.MessageRequest{
				BaseRequest: domainSend.BaseRequest{Phone: group},
				Message:     "Atencao, reunião às 14h.",
				Mentions:    []string{"5588999999999"},
			}},
			err: nil,
		},
		{
			name: "should success with multiple ghost mentions",
			args: args{request: domainSend.MessageRequest{
				BaseRequest: domainSend.BaseRequest{Phone: group},
				Message:     "Favor confirmar presença.",
				Mentions:    []string{"5588999999999", "5588988888888"},
			}},
			err: nil,
		},
		{
			name: "should success with @everyone keyword",
			args: args{request: domainSend.MessageRequest{
				BaseRequest: domainSend.BaseRequest{Phone: group},
				Message:     "Aviso para todo o grupo.",
				Mentions:    []string{"@everyone"},
			}},
			err: nil,
		},
		{
			name: "should success mixing @everyone with explicit numbers",
			args: args{request: domainSend.MessageRequest{
				BaseRequest: domainSend.BaseRequest{Phone: group},
				Message:     "Aviso geral.",
				Mentions:    []string{"@everyone", "5588977777777"},
			}},
			err: nil,
		},
		{
			name: "should success with full JID as mention",
			args: args{request: domainSend.MessageRequest{
				BaseRequest: domainSend.BaseRequest{Phone: group},
				Message:     "Mencionando por JID.",
				Mentions:    []string{"5588966666666@s.whatsapp.net"},
			}},
			err: nil,
		},
		{
			name: "should success without mentions (plain message)",
			args: args{request: domainSend.MessageRequest{
				BaseRequest: domainSend.BaseRequest{Phone: group},
				Message:     "Mensagem normal sem menções.",
			}},
			err: nil,
		},
		{
			name: "should error with local-format mention (leading zero)",
			args: args{request: domainSend.MessageRequest{
				BaseRequest: domainSend.BaseRequest{Phone: group},
				Message:     "Menção inválida.",
				Mentions:    []string{"088999999999"},
			}},
			err: pkgError.ValidationError("mention 088999999999: phone number must be in international format"),
		},
		{
			name: "should error with empty mention",
			args: args{request: domainSend.MessageRequest{
				BaseRequest: domainSend.BaseRequest{Phone: group},
				Message:     "Menção vazia.",
				Mentions:    []string{""},
			}},
			err: pkgError.ValidationError("mention : phone number must be in international format"),
		},
		{
			name: "should error when one mention in the list is invalid",
			args: args{request: domainSend.MessageRequest{
				BaseRequest: domainSend.BaseRequest{Phone: group},
				Message:     "Lista com item inválido.",
				Mentions:    []string{"5588999999999", "08999999999"},
			}},
			err: pkgError.ValidationError("mention 08999999999: phone number must be in international format"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSendMessage(context.Background(), tt.args.request)
			assert.Equal(t, tt.err, err)
		})
	}
}

// TestValidateSendMessage_MentionTextKeepsOriginalMessage guarantees the ghost
// mention feature does not rewrite the message body: the "@" stays out of the
// text exactly as the caller sent it, the notification comes from ContextInfo.
func TestValidateSendMessage_MentionTextKeepsOriginalMessage(t *testing.T) {
	request := domainSend.MessageRequest{
		BaseRequest: domainSend.BaseRequest{Phone: "120363000000000000@g.us"},
		Message:     "Atencao equipe, reunião às 14h.",
		Mentions:    []string{"5588999999999", "@everyone"},
	}

	assert.NoError(t, ValidateSendMessage(context.Background(), request))
	assert.Equal(t, "Atencao equipe, reunião às 14h.", request.Message)
	assert.NotContains(t, request.Message, "@")
}
