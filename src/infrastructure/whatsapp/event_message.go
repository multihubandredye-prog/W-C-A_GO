package whatsapp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"encoding/base64"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/pollstore"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types/events"
)

var reMention = regexp.MustCompile(`\B@\w+`)

// Event types for webhook payload
const (
	EventTypeMessage         = "message"
	EventTypeMessageReaction = "message.reaction"
	EventTypeMessageRevoked  = "message.revoked"
	EventTypeMessageEdited   = "message.edited"
	EventTypeMessagePollVote = "message.poll_vote"
	EventTypeStatusResponseMessage = "status.response"
)

// WebhookEvent is the top-level structure for webhook payloads
type WebhookEvent struct {
	Event    string         `json:"Event"`
	DeviceID string         `json:"DeviceID"`
	Payload  map[string]any `json:"Payload"`
}

type webhookContactPayload struct {
	DisplayName string `json:"DisplayName"`
	VCard       string `json:"VCard"`
	PhoneNumber string `json:"PhoneNumber,omitempty"`
}

// TriggerWebhookForSentMessage manually triggers a webhook event for a message that was just sent.
// This is useful for ensuring immediate outgoing notifications for API calls without waiting for server echos.
func TriggerWebhookForSentMessage(ctx context.Context, client *whatsmeow.Client, recipient types.JID, msg *waE2E.Message, ts whatsmeow.SendResponse) {
	evt := &events.Message{
		Info: types.MessageInfo{
			ID:        ts.ID,
			Timestamp: ts.Timestamp,
			MessageSource: types.MessageSource{
				IsFromMe: true,
				Sender:   *client.Store.ID,
				Chat:     recipient,
			},
		},
		Message: msg,
	}

	// We wrap in a goroutine to not block the main API response
	go func() {
		webhookCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := forwardMessageToWebhook(webhookCtx, client, evt); err != nil {
			logrus.Warnf("Failed to trigger manual outgoing webhook for %s: %v", ts.ID, err)
		}
	}()
}

// forwardMessageToWebhook is a helper function to forward message event to webhook url
func forwardMessageToWebhook(ctx context.Context, client *whatsmeow.Client, evt *events.Message) error {
	webhookEvent, err := createWebhookEvent(ctx, client, evt)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"Event":    webhookEvent.Event,
		"DeviceID": webhookEvent.DeviceID,
		"Payload":  webhookEvent.Payload,
	}

	// Always use the internal Event type for filtering and consistency
	return forwardPayloadToConfiguredWebhooks(ctx, payload, webhookEvent.Event)
}

func isReactionMessage(evt *events.Message) bool {
	if evt == nil || evt.Message == nil {
		return false
	}

	return utils.UnwrapMessage(evt.Message).GetReactionMessage() != nil
}

func createWebhookEvent(ctx context.Context, client *whatsmeow.Client, evt *events.Message) (*WebhookEvent, error) {
	webhookEvent := &WebhookEvent{
		Event:   EventTypeMessage,
		Payload: make(map[string]any),
	}

	// Set DeviceID
	if client != nil && client.Store != nil && client.Store.ID != nil {
		deviceJID := NormalizeJIDFromLID(ctx, client.Store.ID.ToNonAD(), client)
		webhookEvent.DeviceID = deviceJID.ToNonAD().String()
	}

	// Determine event type and build payload
	eventType, payload, err := buildEventPayload(ctx, client, evt)
	if err != nil {
		return nil, err
	}

	webhookEvent.Event = eventType
	webhookEvent.Payload = payload

	return webhookEvent, nil
}

func buildEventPayload(ctx context.Context, client *whatsmeow.Client, evt *events.Message) (string, map[string]any, error) {
	payload := make(map[string]any)

	msg := utils.UnwrapMessage(evt.Message)

	// WhatsApp clients (especially LID-migrated accounts on recent app builds)
	// often wrap messages in a SecretEncryptedMessage envelope. Decrypt those
	// here using whatsmeow's helper to access the inner message content.
	if sem := msg.GetSecretEncryptedMessage(); sem != nil && client != nil {
		if decrypted, err := client.DecryptSecretEncryptedMessage(ctx, evt); err != nil {
			logrus.Warnf("Failed to decrypt SecretEncryptedMessage for %s: %v", evt.Info.ID, err)
		} else if decrypted != nil {
			msg = utils.UnwrapMessage(decrypted)
		}
	}

	// Determine general message type (e.g., Message, LinkMessage, StatusResponseMessage)
	messageType := getMessagePascalType(msg)
	payload["Type"] = messageType

	// If it's a Status Reply Message, extract specific quoted status details
	if messageType == "StatusResponseMessage" {
		if quotedMsgInfo := utils.ExtractContextInfo(msg); quotedMsgInfo != nil && quotedMsgInfo.GetQuotedMessage() != nil {
			// It's a reply to a status
			payload["QuotedStatusID"] = quotedMsgInfo.GetStanzaID()     // ID of the status message being replied to
			payload["QuotedStatusSender"] = quotedMsgInfo.GetParticipant() // Original sender of the status

			// Extract original status message content if available
			quotedMsg := quotedMsgInfo.GetQuotedMessage()
			if quotedMsg.GetExtendedTextMessage() != nil {
				payload["QuotedStatusText"] = quotedMsg.GetExtendedTextMessage().GetText()
				if quotedMsg.GetExtendedTextMessage().GetTitle() != "" {
					payload["QuotedStatusTitle"] = quotedMsg.GetExtendedTextMessage().GetTitle()
				}
				if quotedMsg.GetExtendedTextMessage().GetDescription() != "" {
					payload["QuotedStatusDescription"] = quotedMsg.GetExtendedTextMessage().GetDescription()
				}
			} else if quotedMsg.GetConversation() != "" {
				payload["QuotedStatusText"] = quotedMsg.GetConversation()
			} else if media := quotedMsg.GetImageMessage(); media != nil {
				payload["QuotedStatusType"] = "ImageStatus"
				payload["QuotedStatusCaption"] = media.GetCaption()
			} else if media := quotedMsg.GetVideoMessage(); media != nil {
				payload["QuotedStatusType"] = "VideoStatus"
				payload["QuotedStatusCaption"] = media.GetCaption()
			}
		}
	}

	// Common fields for all message types
	payload["ID"] = evt.Info.ID
	payload["Timestamp"] = evt.Info.Timestamp.Format(time.RFC3339)
	payload["IsFromMe"] = evt.Info.IsFromMe
	payload["IsGroup"] = evt.Info.Chat.Server == types.GroupServer

	// Build from/from_lid fields
	buildFromFields(ctx, client, evt, payload)

	// Set FromName (pushname)
	if pushname := evt.Info.PushName; pushname != "" {
		payload["FromName"] = pushname
	}

	// Check for protocol messages (revoke, edit)
	if protocolMessage := msg.GetProtocolMessage(); protocolMessage != nil {
		protocolType := protocolMessage.GetType().String()

		switch protocolType {
		case "REVOKE":
			if key := protocolMessage.GetKey(); key != nil {
				payload["RevokedMessageID"] = key.GetID()
				payload["RevokedFromMe"] = key.GetFromMe()
				if key.GetRemoteJID() != "" {
					payload["RevokedChat"] = key.GetRemoteJID()
				}
			}
			return EventTypeMessageRevoked, payload, nil

		case "MESSAGE_EDIT":
			if key := protocolMessage.GetKey(); key != nil {
				payload["OriginalMessageID"] = key.GetID()
			}
			if editedMessage := protocolMessage.GetEditedMessage(); editedMessage != nil {
				if editedText := editedMessage.GetExtendedTextMessage(); editedText != nil {
					payload["Body"] = editedText.GetText()
				} else if editedConv := editedMessage.GetConversation(); editedConv != "" {
					payload["Body"] = editedConv
				}
			}
			return EventTypeMessageEdited, payload, nil
		}
	}

	// Check for reaction message
	if reactionMessage := msg.GetReactionMessage(); reactionMessage != nil {
		payload["Reaction"] = reactionMessage.GetText()
		if key := reactionMessage.GetKey(); key != nil {
			payload["ReactedMessageID"] = key.GetID()
		}
		payload["Type"] = "ReactionMessage"
		return EventTypeMessageReaction, payload, nil
	}

	// Check for poll vote
	if pollUpdate := msg.GetPollUpdateMessage(); pollUpdate != nil {
		originalMsgID := pollUpdate.GetPollCreationMessageKey().GetID()
		payload["OriginalMessageID"] = originalMsgID
		payload["Type"] = "PollResponseMessage"

		pollData, found := pollstore.DefaultPollStore.GetPoll(originalMsgID)
		if !found || pollData.EncKey == nil {
			logrus.Warnf("Original poll message %s or its encKey not found in store, cannot decrypt votes", originalMsgID)
			payload["Votes"] = "could not decrypt, original poll data not found"
		} else {
			decryptedVote, err := manualDecryptPollVote(&evt.Info, pollUpdate, pollData.EncKey)
			if err != nil {
				logrus.Errorf("could not manually decrypt poll vote for message %s: %v", originalMsgID, err)
				payload["Votes"] = fmt.Sprintf("could not decrypt, decryption failed: %v", err)
			} else {
				selectedHashes := make(map[string]struct{})
				for _, hash := range decryptedVote.GetSelectedOptions() {
					selectedHashes[hex.EncodeToString(hash)] = struct{}{}
				}

				var decryptedVotes []string
				for _, option := range pollData.Options {
					hash := sha256.Sum256([]byte(option))
					hashStr := hex.EncodeToString(hash[:])
					if _, ok := selectedHashes[hashStr]; ok {
						decryptedVotes = append(decryptedVotes, option)
					}
				}
				payload["PollQuestion"] = pollData.Question
				payload["PollOptions"] = pollData.Options
				payload["PollVotes"] = decryptedVotes
			}
		}
		return EventTypeMessagePollVote, payload, nil
	}

	// Regular message - build body and media fields
	if err := buildMessageBody(ctx, client, evt, payload); err != nil {
		return "", nil, err
	}

	// Add optional fields
	if err := buildOptionalFields(ctx, client, evt, msg, payload); err != nil {
		return "", nil, err
	}

	if messageType == "StatusResponseMessage" {
		return EventTypeStatusResponseMessage, payload, nil
	}

	return EventTypeMessage, payload, nil
}

func getMessagePascalType(msg *waE2E.Message) string {
	if ci := utils.ExtractContextInfo(msg); ci != nil && ci.GetRemoteJID() == "status@broadcast" {
		return "StatusResponseMessage"
	}
	switch {
	case msg.GetExtendedTextMessage() != nil:
		extendedText := msg.GetExtendedTextMessage()
		if extendedText != nil {
			// Check for link preview via Title/Description directly on ExtendedTextMessage
			if extendedText.GetTitle() != "" || extendedText.GetDescription() != "" {
				return "LinkMessage"
			}
		}
		return "Message" // Default for ExtendedTextMessage without link preview
	case msg.GetConversation() != "":
		return "Message"
	case msg.GetImageMessage() != nil:
		return "ImageMessage"
	case msg.GetVideoMessage() != nil:
		return "VideoMessage"
	case msg.GetAudioMessage() != nil:
		return "AudioMessage"
	case msg.GetDocumentMessage() != nil:
		return "DocumentMessage"
	case msg.GetStickerMessage() != nil:
		return "StickerMessage"
	case msg.GetPollCreationMessage() != nil:
		return "PollMessage"
	case msg.GetContactMessage() != nil:
		return "ContactMessage"
	case msg.GetLocationMessage() != nil:
		return "LocationMessage"
	case msg.GetLiveLocationMessage() != nil: // Added missing case
		return "LiveLocationMessage"
	case msg.GetPtvMessage() != nil:
		return "VideoNoteMessage"
	default:
		return "Unknown"
	}
}

func buildFromFields(ctx context.Context, client *whatsmeow.Client, evt *events.Message, payload map[string]any) {
	chatJID := evt.Info.Chat.ToNonAD()
	if chatJID.Server == "lid" {
		payload["ChatLid"] = chatJID.String()
		chatJID = NormalizeJIDFromLID(ctx, chatJID, client).ToNonAD()
	}
	payload["ChatID"] = chatJID.String()

	senderJID := evt.Info.Sender
	if senderJID.Server == "lid" {
		payload["FromLid"] = senderJID.ToNonAD().String()
	}

	normalizedSenderJID := NormalizeJIDFromLID(ctx, senderJID, client)
	payload["From"] = normalizedSenderJID.ToNonAD().String()

	// Add group name for group messages
	if chatJID.Server == types.GroupServer {
		groupInfo, err := client.GetGroupInfo(ctx, chatJID)
		if err == nil && groupInfo != nil {
			payload["GroupName"] = groupInfo.Name
		}
	}
}

func buildMessageBody(ctx context.Context, client *whatsmeow.Client, evt *events.Message, payload map[string]any) error {
	message := utils.BuildEventMessage(evt)

	// Replace LID mentions with phone numbers in text
	if message.Text != "" && client != nil && client.Store != nil && client.Store.LIDs != nil {
		tags := reMention.FindAllString(message.Text, -1)
		tagsMap := make(map[string]bool)
		for _, tag := range tags {
			tagsMap[tag] = true
		}
		for tag := range tagsMap {
			lid, err := types.ParseJID(tag[1:] + "@lid")
			if err != nil {
				logrus.Errorf("Error when parse jid: %v", err)
			} else {
				pn, err := client.Store.LIDs.GetPNForLID(ctx, lid)
				if err != nil {
					logrus.Errorf("Error when get pn for lid %s: %v", lid.ToNonAD().String(), err)
				}
				if !pn.IsEmpty() {
					message.Text = strings.Replace(message.Text, tag, fmt.Sprintf("@%s", pn.User), -1)
				}
			}
		}
		payload["Body"] = message.Text
	} else if message.Text != "" {
		payload["Body"] = message.Text
	}

	// Fallback: extract caption from media messages if no text Body was set
	if _, hasBody := payload["Body"]; !hasBody {
		msg := utils.UnwrapMessage(evt.Message)
		if caption := utils.ExtractMediaCaption(msg); caption != "" {
			payload["Body"] = caption
		}
	}

	// Add reply context if present
	if message.RepliedId != "" {
		payload["RepliedToID"] = message.RepliedId
		if payload["Type"] != "StatusResponseMessage" {
			payload["Type"] = "QuoteMessage"
		}
	}
	if message.QuotedMessage != "" {
		payload["QuotedBody"] = message.QuotedMessage
	}

	if body, ok := payload["Body"]; ok {
		payload["Message"] = body
	}

	return nil
}

func buildOptionalFields(ctx context.Context, client *whatsmeow.Client, evt *events.Message, msg *waE2E.Message, payload map[string]any) error {
	if evt.IsViewOnce {
		payload["ViewOnce"] = true
	}

	if utils.BuildForwarded(evt) {
		payload["Forwarded"] = true
	}

	if referral := utils.ExtractExternalAdReply(msg); referral != nil {
		payload["Referral"] = referral
	}

    if payload["Type"] == "LinkMessage" {
        if extendedText := msg.GetExtendedTextMessage(); extendedText != nil {
            if v := extendedText.GetTitle(); v != "" {
                payload["LinkTitle"] = v
            }
            if v := extendedText.GetDescription(); v != "" {
                payload["LinkDescription"] = v
            }

            // Try to get the high-quality image from the URL
            url := extendedText.GetMatchedText()
            if url != "" {
                meta, err := utils.GetMetaDataFromURL(url)
                if err == nil && len(meta.JPEGThumb) > 0 {
                    payload["LinkThumbnailBase64"] = base64.StdEncoding.EncodeToString(meta.JPEGThumb)
                    payload["LinkImageHighQuality"] = true
                }
            }
            
            // Fallback to WhatsApp's low-quality thumbnail if high-quality image wasn't found or added
            if _, exists := payload["LinkThumbnailBase64"]; !exists {
                if v := extendedText.GetJPEGThumbnail(); len(v) > 0 {
                    payload["LinkThumbnailBase64"] = base64.StdEncoding.EncodeToString(v)
                    payload["LinkImageHighQuality"] = false
                }
            }
        }
    }

	if err := buildMediaFields(ctx, client, msg, payload); err != nil {
		return err
	}

	buildOtherMessageTypes(msg, payload)

	return nil
}

func buildMediaFields(ctx context.Context, client *whatsmeow.Client, msg *waE2E.Message, payload map[string]any) error {
	if audioMedia := msg.GetAudioMessage(); audioMedia != nil {
		ext := ""
		if mt := audioMedia.GetMimetype(); mt != "" {
			if extensions, err := filepath.Match("*/*", mt); err == nil && extensions {
				// Fallback to basic extraction if needed, but let's use what we can
			}
		}
		
		payload["Extension"] = filepath.Ext(audioMedia.GetURL()) // Usually empty for WA URLs
		if mt := audioMedia.GetMimetype(); mt != "" {
			parts := strings.Split(mt, "/")
			if len(parts) > 1 {
				ext = "." + strings.Split(parts[1], ";")[0]
				payload["Extension"] = ext
			}
		}
		payload["FileSizeKB"] = float64(audioMedia.GetFileLength()) / 1024

		if config.WhatsappAutoDownloadMedia {
			extracted, err := utils.ExtractMedia(ctx, client, config.PathMedia, audioMedia)
			if err != nil {
				logrus.Errorf("Failed to download audio: %v", err)
				return pkgError.WebhookError(fmt.Sprintf("Failed to download audio: %v", err))
			}
			payload["Audio"] = extracted.MediaPath
		} else {
			payload["Audio"] = map[string]any{
				"URL": audioMedia.GetURL(),
			}
		}
	}

	if documentMedia := msg.GetDocumentMessage(); documentMedia != nil {
		payload["Extension"] = filepath.Ext(documentMedia.GetFileName())
		payload["FileSizeKB"] = float64(documentMedia.GetFileLength()) / 1024
		
		if config.WhatsappAutoDownloadMedia {
			extracted, err := utils.ExtractMedia(ctx, client, config.PathMedia, documentMedia)
			if err != nil {
				logrus.Errorf("Failed to download document: %v", err)
				return pkgError.WebhookError(fmt.Sprintf("Failed to download document: %v", err))
			}
			payload["Document"] = buildAutoDownloadPayload(extracted)
		} else {
			payload["Document"] = map[string]any{
				"URL":      documentMedia.GetURL(),
				"FileName": documentMedia.GetFileName(),
			}
		}
	}

	if imageMedia := msg.GetImageMessage(); imageMedia != nil {
		ext := ".jpg"
		if mt := imageMedia.GetMimetype(); mt != "" {
			parts := strings.Split(mt, "/")
			if len(parts) > 1 {
				ext = "." + strings.Split(parts[1], ";")[0]
			}
		}
		payload["Extension"] = ext
		payload["FileSizeKB"] = float64(imageMedia.GetFileLength()) / 1024

		if config.WhatsappAutoDownloadMedia {
			extracted, err := utils.ExtractMedia(ctx, client, config.PathMedia, imageMedia)
			if err != nil {
				logrus.Errorf("Failed to download image: %v", err)
				return pkgError.WebhookError(fmt.Sprintf("Failed to download image: %v", err))
			}
			payload["Image"] = buildAutoDownloadPayload(extracted)
		} else {
			payload["Image"] = map[string]any{
				"URL":     imageMedia.GetURL(),
				"Caption": imageMedia.GetCaption(),
			}
		}
	}

	if stickerMedia := msg.GetStickerMessage(); stickerMedia != nil {
		payload["Extension"] = ".webp"
		payload["FileSizeKB"] = float64(stickerMedia.GetFileLength()) / 1024
		
		if config.WhatsappAutoDownloadMedia {
			extracted, err := utils.ExtractMedia(ctx, client, config.PathMedia, stickerMedia)
			if err != nil {
				logrus.Errorf("Failed to download sticker: %v", err)
				return pkgError.WebhookError(fmt.Sprintf("Failed to download sticker: %v", err))
			}
			payload["Sticker"] = extracted.MediaPath
		} else {
			payload["Sticker"] = map[string]any{
				"URL": stickerMedia.GetURL(),
			}
		}
	}

	if videoMedia := msg.GetVideoMessage(); videoMedia != nil {
		ext := ".mp4"
		if mt := videoMedia.GetMimetype(); mt != "" {
			parts := strings.Split(mt, "/")
			if len(parts) > 1 {
				ext = "." + strings.Split(parts[1], ";")[0]
			}
		}
		payload["Extension"] = ext
		payload["FileSizeKB"] = float64(videoMedia.GetFileLength()) / 1024

		if config.WhatsappAutoDownloadMedia {
			extracted, err := utils.ExtractMedia(ctx, client, config.PathMedia, videoMedia)
			if err != nil {
				logrus.Errorf("Failed to download video: %v", err)
				return pkgError.WebhookError(fmt.Sprintf("Failed to download video: %v", err))
			}
			payload["Video"] = buildAutoDownloadPayload(extracted)
		} else {
			payload["Video"] = map[string]any{
				"URL":     videoMedia.GetURL(),
				"Caption": videoMedia.GetCaption(),
			}
		}
	}

	if ptvMedia := msg.GetPtvMessage(); ptvMedia != nil {
		payload["Extension"] = ".mp4"
		payload["FileSizeKB"] = float64(ptvMedia.GetFileLength()) / 1024
		
		if config.WhatsappAutoDownloadMedia {
			extracted, err := utils.ExtractMedia(ctx, client, config.PathMedia, ptvMedia)
			if err != nil {
				logrus.Errorf("Failed to download video note: %v", err)
				return pkgError.WebhookError(fmt.Sprintf("Failed to download video note: %v", err))
			}
			payload["VideoNote"] = buildAutoDownloadPayload(extracted)
		} else {
			payload["VideoNote"] = map[string]any{
				"URL":     ptvMedia.GetURL(),
				"Caption": ptvMedia.GetCaption(),
			}
		}
	}

	return nil
}

// buildAutoDownloadPayload builds the media payload for auto-downloaded media.
// Returns just the path string if no caption (backward compatible), or a map with path+caption.
func buildAutoDownloadPayload(extracted utils.ExtractedMedia) any {
	if extracted.Caption != "" {
		return map[string]any{
			"Path":    extracted.MediaPath,
			"Caption": extracted.Caption,
		}
	}
	return extracted.MediaPath
}

func buildOtherMessageTypes(msg *waE2E.Message, payload map[string]any) {
	if contactMessage := msg.GetContactMessage(); contactMessage != nil {
		payload["Contact"] = buildWebhookContactPayload(contactMessage)
	}

	if contactsArrayMessage := msg.GetContactsArrayMessage(); contactsArrayMessage != nil {
		payload["ContactsArray"] = buildWebhookContactsArrayPayload(contactsArrayMessage.GetContacts())
	}

	if listMessage := msg.GetListMessage(); listMessage != nil {
		payload["List"] = listMessage
	}

	if liveLocationMessage := msg.GetLiveLocationMessage(); liveLocationMessage != nil {
		payload["LiveLocation"] = liveLocationMessage
        payload["Latitude"] = liveLocationMessage.GetDegreesLatitude()
        payload["Longitude"] = liveLocationMessage.GetDegreesLongitude()
	}

	if locationMessage := msg.GetLocationMessage(); locationMessage != nil {
		payload["Location"] = locationMessage
        payload["Latitude"] = locationMessage.GetDegreesLatitude()
        payload["Longitude"] = locationMessage.GetDegreesLongitude()
	}

	if orderMessage := msg.GetOrderMessage(); orderMessage != nil {
		payload["Order"] = orderMessage
	}

	if pollCreation := msg.GetPollCreationMessage(); pollCreation != nil {
		var options []string
		for _, option := range pollCreation.GetOptions() {
			options = append(options, option.GetOptionName())
		}
		payload["Poll"] = map[string]any{
			"Question": pollCreation.GetName(),
			"Options":  options,
			"EncKey":   msg.GetMessageContextInfo().GetMessageSecret(),
		}
	}
}

func buildWebhookContactPayload(contact *waE2E.ContactMessage) webhookContactPayload {
	if contact == nil {
		return webhookContactPayload{}
	}

	vcard := contact.GetVcard()
	return webhookContactPayload{
		DisplayName: contact.GetDisplayName(),
		VCard:       vcard,
		PhoneNumber: utils.ExtractPhoneFromVCard(vcard),
	}
}

func buildWebhookContactsArrayPayload(contacts []*waE2E.ContactMessage) []webhookContactPayload {
	result := make([]webhookContactPayload, 0, len(contacts))
	for _, contact := range contacts {
		if contact == nil {
			continue
		}
		result = append(result, buildWebhookContactPayload(contact))
	}
	return result
}
