package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/validations"
	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// interactiveHeaderImageMaxBytes bounds the header image download so a hostile
// or mistyped URL cannot exhaust memory.
const interactiveHeaderImageMaxBytes = 5 * 1024 * 1024

// buttonsBizNode is the binary node WhatsApp requires for NativeFlow buttons.
//
// Without it the InteractiveMessage is delivered but the recipient's client
// renders it as plain text with no buttons at all. This node is what tells the
// server to process the payload as an interactive native flow.
func buttonsBizNode() []waBinary.Node {
	return []waBinary.Node{{
		Tag: "biz",
		Content: []waBinary.Node{{
			Tag:   "interactive",
			Attrs: waBinary.Attrs{"type": "native_flow", "v": "1"},
			Content: []waBinary.Node{{
				Tag:   "native_flow",
				Attrs: waBinary.Attrs{"v": "9", "name": "mixed"},
			}},
		}},
	}}
}

// listBizNode is the binary node WhatsApp requires for interactive lists.
//
// Note this differs from buttonsBizNode: lists use biz > list(product_list),
// buttons use biz > interactive(native_flow). Reusing the buttons node for a
// list makes the message fail to render.
func listBizNode() []waBinary.Node {
	return []waBinary.Node{{
		Tag: "biz",
		Content: []waBinary.Node{{
			Tag:   "list",
			Attrs: waBinary.Attrs{"type": "product_list", "v": "2"},
		}},
	}}
}

// truncateRunes shortens s to at most limit runes, counting runes rather than
// bytes so accented characters and emoji are never split mid-sequence.
func truncateRunes(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit])
}

// buildNativeFlowButtons converts the request buttons into NativeFlow buttons.
// Each button carries a JSON parameter blob whose shape depends on its type.
func buildNativeFlowButtons(buttons []domainSend.Button) ([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, error) {
	nativeButtons := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0, len(buttons))

	for i, button := range buttons {
		title := truncateRunes(strings.TrimSpace(button.Title), domainSend.MaxButtonTitleLength)

		buttonType := strings.ToLower(strings.TrimSpace(button.Type))
		if buttonType == "" {
			buttonType = domainSend.ButtonTypeReply
		}

		id := strings.TrimSpace(button.ID)
		if id == "" {
			id = title
		}

		var name string
		var params map[string]string

		switch buttonType {
		case domainSend.ButtonTypeReply:
			name = "quick_reply"
			params = map[string]string{"display_text": title, "id": id}
		case domainSend.ButtonTypeURL:
			name = "cta_url"
			params = map[string]string{
				"display_text": title,
				"url":          strings.TrimSpace(button.URL),
				"merchant_url": strings.TrimSpace(button.URL),
			}
		case domainSend.ButtonTypeCall:
			name = "cta_call"
			params = map[string]string{
				"display_text": title,
				"phone_number": strings.TrimSpace(button.PhoneNumber),
			}
		case domainSend.ButtonTypeCopy:
			name = "cta_copy"
			params = map[string]string{
				"display_text": title,
				"copy_code":    strings.TrimSpace(button.CopyCode),
			}
		default:
			return nil, pkgError.ValidationError(fmt.Sprintf("buttons[%d].type: %q is not supported.", i, button.Type))
		}

		paramsJSON, err := json.Marshal(params)
		if err != nil {
			return nil, pkgError.InternalServerError(fmt.Sprintf("failed to encode button params: %v", err))
		}

		nativeButtons = append(nativeButtons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String(name),
			ButtonParamsJSON: proto.String(string(paramsJSON)),
		})
	}

	return nativeButtons, nil
}

// resolveHeaderImage fetches the optional header image and uploads it to
// WhatsApp. Accepts a data:image base64 URI or an http(s) URL.
func (service serviceSend) resolveHeaderImage(ctx context.Context, client *whatsmeow.Client, source string) (*waE2E.ImageMessage, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, nil
	}

	var payload []byte

	switch {
	case strings.HasPrefix(source, "data:"):
		commaIdx := strings.Index(source, ",")
		if commaIdx < 0 {
			return nil, pkgError.ValidationError("image_url: malformed data URI.")
		}
		decoded, err := base64.StdEncoding.DecodeString(source[commaIdx+1:])
		if err != nil {
			return nil, pkgError.ValidationError("image_url: invalid base64 payload.")
		}
		payload = decoded
	case strings.HasPrefix(source, "http://"), strings.HasPrefix(source, "https://"):
		downloaded, _, err := utils.DownloadImageFromURL(source)
		if err != nil {
			return nil, pkgError.ValidationError(fmt.Sprintf("image_url: failed to download image: %v", err))
		}
		payload = downloaded
	default:
		return nil, pkgError.ValidationError("image_url: must be an http(s) URL or a data:image base64 URI.")
	}

	if len(payload) == 0 {
		return nil, pkgError.ValidationError("image_url: downloaded image is empty.")
	}
	if len(payload) > interactiveHeaderImageMaxBytes {
		return nil, pkgError.ValidationError("image_url: image exceeds the 5MB header limit.")
	}

	uploaded, err := client.Upload(ctx, payload, whatsmeow.MediaImage)
	if err != nil {
		return nil, pkgError.InternalServerError(fmt.Sprintf("failed to upload header image: %v", err))
	}

	return &waE2E.ImageMessage{
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(http.DetectContentType(payload)),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(payload))),
	}, nil
}

// SendButtons sends an interactive message with up to 3 NativeFlow buttons.
func (service serviceSend) SendButtons(ctx context.Context, request domainSend.ButtonsRequest) (response domainSend.GenericResponse, err error) {
	if err = validations.ValidateSendButtons(ctx, request); err != nil {
		return response, err
	}

	client := whatsapp.ClientFromContext(ctx)
	if client == nil {
		return response, pkgError.ErrWaCLI
	}

	recipient, err := utils.ValidateJidWithLogin(client, request.Phone)
	if err != nil {
		return response, err
	}

	nativeButtons, err := buildNativeFlowButtons(request.Buttons)
	if err != nil {
		return response, err
	}

	headerImage, err := service.resolveHeaderImage(ctx, client, request.ImageURL)
	if err != nil {
		return response, err
	}

	contextInfo := &waE2E.ContextInfo{}
	if request.IsForwarded {
		contextInfo.IsForwarded = proto.Bool(true)
	}
	if request.Duration != nil && *request.Duration > 0 {
		contextInfo.Expiration = proto.Uint32(uint32(*request.Duration))
	}

	interactive := &waE2E.InteractiveMessage{
		Header:      &waE2E.InteractiveMessage_Header{},
		Body:        &waE2E.InteractiveMessage_Body{Text: proto.String(request.Body)},
		ContextInfo: contextInfo,
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons:        nativeButtons,
				MessageVersion: proto.Int32(1),
			},
		},
	}

	if footer := strings.TrimSpace(request.Footer); footer != "" {
		interactive.Footer = &waE2E.InteractiveMessage_Footer{Text: proto.String(footer)}
	}

	// A media header takes precedence over a plain title header.
	if headerImage != nil {
		interactive.Header.HasMediaAttachment = proto.Bool(true)
		interactive.Header.Media = &waE2E.InteractiveMessage_Header_ImageMessage{ImageMessage: headerImage}
		if title := strings.TrimSpace(request.Title); title != "" {
			interactive.Header.Title = proto.String(title)
		}
	} else if title := strings.TrimSpace(request.Title); title != "" {
		interactive.Header.Title = proto.String(title)
	}

	msg := &waE2E.Message{InteractiveMessage: interactive}

	extraNodes := buttonsBizNode()
	ts, err := service.wrapSendMessage(ctx, client, recipient, msg, request.Body, whatsmeow.SendRequestExtra{
		ID:              client.GenerateMessageID(),
		AdditionalNodes: &extraNodes,
	})
	if err != nil {
		return response, err
	}

	response.MessageID = ts.ID
	response.Status = fmt.Sprintf("Send buttons success %s (server timestamp: %s)", request.Phone, ts.Timestamp.String())
	return response, nil
}

// SendList sends an interactive list message. Lists are the way to offer more
// than 3 options: rows are grouped in sections and shown inside a picker.
func (service serviceSend) SendList(ctx context.Context, request domainSend.ListRequest) (response domainSend.GenericResponse, err error) {
	if err = validations.ValidateSendList(ctx, request); err != nil {
		return response, err
	}

	client := whatsapp.ClientFromContext(ctx)
	if client == nil {
		return response, pkgError.ErrWaCLI
	}

	recipient, err := utils.ValidateJidWithLogin(client, request.Phone)
	if err != nil {
		return response, err
	}

	sections := make([]*waE2E.ListMessage_Section, 0, len(request.Sections))
	for _, section := range request.Sections {
		rows := make([]*waE2E.ListMessage_Row, 0, len(section.Rows))
		for _, row := range section.Rows {
			title := truncateRunes(strings.TrimSpace(row.Title), domainSend.MaxListRowTitleLength)

			rowID := strings.TrimSpace(row.RowID)
			if rowID == "" {
				rowID = title
			}

			protoRow := &waE2E.ListMessage_Row{
				RowID: proto.String(rowID),
				Title: proto.String(title),
			}
			if description := strings.TrimSpace(row.Description); description != "" {
				protoRow.Description = proto.String(truncateRunes(description, domainSend.MaxListRowDescriptionLength))
			}
			rows = append(rows, protoRow)
		}

		protoSection := &waE2E.ListMessage_Section{Rows: rows}
		if title := strings.TrimSpace(section.Title); title != "" {
			protoSection.Title = proto.String(title)
		}
		sections = append(sections, protoSection)
	}

	buttonText := strings.TrimSpace(request.ButtonText)
	if buttonText == "" {
		buttonText = "Select"
	}

	listMsg := &waE2E.ListMessage{
		Description: proto.String(request.Description),
		ButtonText:  proto.String(buttonText),
		ListType:    waE2E.ListMessage_SINGLE_SELECT.Enum(),
		Sections:    sections,
	}

	if title := strings.TrimSpace(request.Title); title != "" {
		listMsg.Title = proto.String(title)
	}
	if footer := strings.TrimSpace(request.Footer); footer != "" {
		listMsg.FooterText = proto.String(footer)
	}

	if request.IsForwarded || (request.Duration != nil && *request.Duration > 0) {
		listMsg.ContextInfo = &waE2E.ContextInfo{}
		if request.IsForwarded {
			listMsg.ContextInfo.IsForwarded = proto.Bool(true)
		}
		if request.Duration != nil && *request.Duration > 0 {
			listMsg.ContextInfo.Expiration = proto.Uint32(uint32(*request.Duration))
		}
	}

	// A ListMessage must travel inside DocumentWithCaptionMessage >
	// FutureProofMessage. Sent bare (or wrapped in ViewOnceMessage, a common
	// mistake) the list does not render and arrives as plain text.
	msg := &waE2E.Message{
		DocumentWithCaptionMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{ListMessage: listMsg},
		},
	}

	extraNodes := listBizNode()
	ts, err := service.wrapSendMessage(ctx, client, recipient, msg, request.Description, whatsmeow.SendRequestExtra{
		ID:              client.GenerateMessageID(),
		AdditionalNodes: &extraNodes,
	})
	if err != nil {
		return response, err
	}

	response.MessageID = ts.ID
	response.Status = fmt.Sprintf("Send list success %s (server timestamp: %s)", request.Phone, ts.Timestamp.String())
	return response, nil
}
