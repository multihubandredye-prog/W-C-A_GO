package whatsapp

import (
	"context"
	"testing"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// TestGetMessagePascalTypeNewMappings locks the content variants that used to
// fall through to "Unknown", which surfaced in group webhooks as untyped
// payloads (the same ID arriving twice, first Unknown then typed, was one of
// the observed symptoms).
func TestGetMessagePascalTypeNewMappings(t *testing.T) {
	tests := []struct {
		name     string
		msg      *waE2E.Message
		expected string
	}{
		{"album de fotos", &waE2E.Message{AlbumMessage: &waE2E.AlbumMessage{}}, "AlbumMessage"},
		{"evento de calendário", &waE2E.Message{EventMessage: &waE2E.EventMessage{}}, "EventMessage"},
		{"convite para evento", &waE2E.Message{EventInviteMessage: &waE2E.EventInviteMessage{}}, "EventInviteMessage"},
		{"convite de grupo", &waE2E.Message{GroupInviteMessage: &waE2E.GroupInviteMessage{}}, "GroupInviteMessage"},
		{"convite de canal", &waE2E.Message{NewsletterAdminInviteMessage: &waE2E.NewsletterAdminInviteMessage{}}, "NewsletterAdminInviteMessage"},
		{"opção adicionada em enquete", &waE2E.Message{PollAddOptionMessage: &waE2E.PollAddOptionMessage{}}, "PollAddOptionMessage"},
		{"resultado de enquete", &waE2E.Message{PollResultSnapshotMessage: &waE2E.PollResultSnapshotMessage{}}, "PollResultSnapshotMessage"},
		{"pacote de figurinhas", &waE2E.Message{StickerPackMessage: &waE2E.StickerPackMessage{}}, "StickerPackMessage"},
		{"comentário", &waE2E.Message{CommentMessage: &waE2E.CommentMessage{}}, "CommentMessage"},
		{"mensagem fixada", &waE2E.Message{PinInChatMessage: &waE2E.PinInChatMessage{}}, "PinInChatMessage"},
		{"keep in chat", &waE2E.Message{KeepInChatMessage: &waE2E.KeepInChatMessage{}}, "KeepInChatMessage"},
		{"pedido", &waE2E.Message{OrderMessage: &waE2E.OrderMessage{}}, "OrderMessage"},
		{"produto do catálogo", &waE2E.Message{ProductMessage: &waE2E.ProductMessage{}}, "ProductMessage"},
		{"fatura", &waE2E.Message{InvoiceMessage: &waE2E.InvoiceMessage{}}, "InvoiceMessage"},
		{"vários contatos", &waE2E.Message{ContactsArrayMessage: &waE2E.ContactsArrayMessage{}}, "ContactsArrayMessage"},
		{"reação criptografada", &waE2E.Message{EncReactionMessage: &waE2E.EncReactionMessage{}}, "EncReactionMessage"},
		{"música", &waE2E.Message{MusicMessage: &waE2E.MusicMessage{}}, "MusicMessage"},
		{"chamada agendada", &waE2E.Message{ScheduledCallCreationMessage: &waE2E.ScheduledCallCreationMessage{}}, "ScheduledCallCreationMessage"},
		{"placeholder", &waE2E.Message{PlaceholderMessage: &waE2E.PlaceholderMessage{}}, "PlaceholderMessage"},
		{"envelope LID sem descriptografia", &waE2E.Message{SecretEncryptedMessage: &waE2E.SecretEncryptedMessage{}}, "SecretEncryptedMessage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMessagePascalType(tt.msg); got != tt.expected {
				t.Fatalf("%s: expected %q, got %q", tt.name, tt.expected, got)
			}
		})
	}
}

// TestGetMessagePascalTypeRegression keeps the previously mapped types stable.
func TestGetMessagePascalTypeRegression(t *testing.T) {
	link := &waE2E.Message{ExtendedTextMessage: &waE2E.ExtendedTextMessage{Title: protoString("t")}}

	tests := []struct {
		name     string
		msg      *waE2E.Message
		expected string
	}{
		{"texto simples", &waE2E.Message{Conversation: protoString("oi")}, "Message"},
		{"texto com quebra", &waE2E.Message{ExtendedTextMessage: &waE2E.ExtendedTextMessage{}}, "Message"},
		{"link", link, "LinkMessage"},
		{"imagem", &waE2E.Message{ImageMessage: &waE2E.ImageMessage{}}, "ImageMessage"},
		{"vídeo", &waE2E.Message{VideoMessage: &waE2E.VideoMessage{}}, "VideoMessage"},
		{"áudio", &waE2E.Message{AudioMessage: &waE2E.AudioMessage{}}, "AudioMessage"},
		{"documento", &waE2E.Message{DocumentMessage: &waE2E.DocumentMessage{}}, "DocumentMessage"},
		{"figurinha", &waE2E.Message{StickerMessage: &waE2E.StickerMessage{}}, "StickerMessage"},
		{"enquete", &waE2E.Message{PollCreationMessage: &waE2E.PollCreationMessage{}}, "PollMessage"},
		{"contato", &waE2E.Message{ContactMessage: &waE2E.ContactMessage{}}, "ContactMessage"},
		{"localização", &waE2E.Message{LocationMessage: &waE2E.LocationMessage{}}, "LocationMessage"},
		{"localização ao vivo", &waE2E.Message{LiveLocationMessage: &waE2E.LiveLocationMessage{}}, "LiveLocationMessage"},
		{"nota de vídeo", &waE2E.Message{PtvMessage: &waE2E.VideoMessage{}}, "VideoNoteMessage"},
		{"resposta de botão", &waE2E.Message{InteractiveResponseMessage: &waE2E.InteractiveResponseMessage{}}, "ButtonsResponseMessage"},
		{"resposta de lista", &waE2E.Message{ListResponseMessage: &waE2E.ListResponseMessage{}}, "ListResponseMessage"},
		{"botões enviados", &waE2E.Message{ButtonsMessage: &waE2E.ButtonsMessage{}}, "ButtonsMessage"},
		{"lista enviada", &waE2E.Message{ListMessage: &waE2E.ListMessage{}}, "ListMessage"},
		{"template", &waE2E.Message{TemplateMessage: &waE2E.TemplateMessage{}}, "TemplateMessage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMessagePascalType(tt.msg); got != tt.expected {
				t.Fatalf("%s: expected %q, got %q", tt.name, tt.expected, got)
			}
		})
	}
}

// TestGetMessagePascalTypeDocumentWithCaptionPeek: UnwrapMessage keeps genuine
// documents wrapped in DocumentWithCaptionMessage on purpose (guarded by
// TestUnwrapKeepsRealDocuments), so the type detector peeks inside instead of
// reporting "Unknown".
func TestGetMessagePascalTypeDocumentWithCaptionPeek(t *testing.T) {
	wrapper := &waE2E.Message{
		DocumentWithCaptionMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{DocumentMessage: &waE2E.DocumentMessage{}},
		},
	}
	if got := getMessagePascalType(wrapper); got != "DocumentMessage" {
		t.Fatalf("expected DocumentMessage via peek, got %q", got)
	}

	// A wrapper with nothing recognizable inside still reports Unknown.
	empty := &waE2E.Message{
		DocumentWithCaptionMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{},
		},
	}
	if got := getMessagePascalType(empty); got != "Unknown" {
		t.Fatalf("expected Unknown for empty wrapper, got %q", got)
	}
}

// TestBuildEventPayloadWrappedMedia reproduces the field case: media inside a
// FutureProof wrapper used to reach the webhook as Type "Unknown"; after the
// unwrap fix it arrives typed as the content it carries.
func TestBuildEventPayloadWrappedMedia(t *testing.T) {
	oldAutoDownload := config.WhatsappAutoDownloadMedia
	config.WhatsappAutoDownloadMedia = false
	t.Cleanup(func() {
		config.WhatsappAutoDownloadMedia = oldAutoDownload
	})

	media := func(wrap func(*waE2E.Message, *waE2E.FutureProofMessage)) *events.Message {
		return &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{
					Chat:   types.NewJID("555", types.DefaultUserServer),
					Sender: types.NewJID("555", types.DefaultUserServer),
				},
				ID:        "MSG-WRAP-1",
				Timestamp: time.Date(2026, time.September, 16, 12, 0, 0, 0, time.UTC),
			},
			Message: func() *waE2E.Message {
				msg := &waE2E.Message{}
				wrap(msg, &waE2E.FutureProofMessage{
					Message: &waE2E.Message{ImageMessage: &waE2E.ImageMessage{}},
				})
				return msg
			}(),
		}
	}

	tests := []struct {
		name     string
		wrap     func(*waE2E.Message, *waE2E.FutureProofMessage)
		expected string
	}{
		{"imagem dentro de editedMessage", func(m *waE2E.Message, f *waE2E.FutureProofMessage) { m.EditedMessage = f }, "ImageMessage"},
		{"imagem dentro de groupMentionedMessage", func(m *waE2E.Message, f *waE2E.FutureProofMessage) { m.GroupMentionedMessage = f }, "ImageMessage"},
		{"imagem dentro de documentWithCaption (peek)", func(m *waE2E.Message, f *waE2E.FutureProofMessage) { m.DocumentWithCaptionMessage = f }, "ImageMessage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventType, payload, err := buildEventPayload(context.Background(), nil, media(tt.wrap))
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if eventType != EventTypeMessage {
				t.Fatalf("expected event %q, got %q", EventTypeMessage, eventType)
			}
			if payload["Type"] != tt.expected {
				t.Fatalf("expected Type=%q, got %v", tt.expected, payload["Type"])
			}
		})
	}
}

// TestBuildEventPayloadAlbum: a photo album used to arrive as "Unknown"; it
// now reports its own type.
func TestBuildEventPayloadAlbum(t *testing.T) {
	oldAutoDownload := config.WhatsappAutoDownloadMedia
	config.WhatsappAutoDownloadMedia = false
	t.Cleanup(func() {
		config.WhatsappAutoDownloadMedia = oldAutoDownload
	})

	evt := &events.Message{
		Info: types.MessageInfo{
			MessageSource: types.MessageSource{
				Chat:   types.NewJID("555", types.DefaultUserServer),
				Sender: types.NewJID("555", types.DefaultUserServer),
			},
			ID:        "MSG-ALBUM-1",
			Timestamp: time.Date(2026, time.September, 16, 12, 0, 0, 0, time.UTC),
		},
		Message: &waE2E.Message{AlbumMessage: &waE2E.AlbumMessage{}},
	}

	eventType, payload, err := buildEventPayload(context.Background(), nil, evt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if eventType != EventTypeMessage {
		t.Fatalf("expected event %q, got %q", EventTypeMessage, eventType)
	}
	if payload["Type"] != "AlbumMessage" {
		t.Fatalf("expected Type=AlbumMessage, got %v", payload["Type"])
	}
}
