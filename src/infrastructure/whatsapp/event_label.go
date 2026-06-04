package whatsapp

import (
	"context"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/proto/waSyncAction"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

const (
	eventTypeLabelEdit        = "label.edit"
	eventTypeLabelAssociation = "label.association"
)

func isLabelAppState(evt *events.AppState) bool {
	if evt == nil || len(evt.Index) == 0 {
		return false
	}

	// Note: IndexLabelAssociationMessage (per-message label) is intentionally excluded.
	// Only chat-level associations and label edits are forwarded.
	return evt.Index[0] == appstate.IndexLabelEdit || evt.Index[0] == appstate.IndexLabelAssociationChat
}

func forwardLabelAppStateToWebhook(ctx context.Context, evt *events.AppState, deviceID string, client *whatsmeow.Client) error {
	eventName, payload := buildLabelAppStatePayload(ctx, evt, client)
	if eventName == "" {
		return nil
	}

	body := map[string]any{
		"Event":     eventName,
		"Timestamp": labelAppStateTimestamp(evt),
		"Payload":   payload,
	}
	if deviceID != "" {
		body["DeviceID"] = deviceID
	}

	return forwardPayloadToConfiguredWebhooks(ctx, body, eventName)
}

func buildLabelAppStatePayload(ctx context.Context, evt *events.AppState, client *whatsmeow.Client) (string, map[string]any) {
	if evt == nil || evt.SyncActionValue == nil || len(evt.Index) == 0 {
		return "", nil
	}

	switch evt.Index[0] {
	case appstate.IndexLabelEdit:
		if len(evt.Index) < 2 || evt.LabelEditAction == nil {
			return "", nil
		}

		payload := map[string]any{
			"LabelID": evt.Index[1],
			"Type":    "LabelEditMessage",
		}
		addLabelEditActionFields(payload, evt.LabelEditAction)

		return eventTypeLabelEdit, payload
	case appstate.IndexLabelAssociationChat:
		if len(evt.Index) < 3 || evt.LabelAssociationAction == nil {
			return "", nil
		}

		payload := map[string]any{
			"LabelID": evt.Index[1],
			"Labeled": evt.LabelAssociationAction.GetLabeled(),
			"Type":    "LabelAssociationMessage",
		}
		addLabelChatFields(ctx, client, payload, evt.Index[2])

		return eventTypeLabelAssociation, payload
	default:
		return "", nil
	}
}

func addLabelEditActionFields(payload map[string]any, action *waSyncAction.LabelEditAction) {
	if action.Name != nil {
		payload["Name"] = action.GetName()
	}
	if action.Color != nil {
		payload["Color"] = action.GetColor()
	}
	if action.PredefinedID != nil {
		payload["PredefinedID"] = action.GetPredefinedID()
	}
	if action.Deleted != nil {
		payload["Deleted"] = action.GetDeleted()
	}
	if action.OrderIndex != nil {
		payload["OrderIndex"] = action.GetOrderIndex()
	}
	if action.IsActive != nil {
		payload["IsActive"] = action.GetIsActive()
	}
	if action.Type != nil {
		payload["LabelType"] = action.GetType().String()
	}
	if action.IsImmutable != nil {
		payload["IsImmutable"] = action.GetIsImmutable()
	}
	if action.MuteEndTimeMS != nil {
		payload["MuteEndTimeMS"] = action.GetMuteEndTimeMS()
	}
}

func addLabelChatFields(ctx context.Context, client *whatsmeow.Client, payload map[string]any, rawJID string) {
	jid, err := types.ParseJID(rawJID)
	if err != nil {
		payload["ChatID"] = rawJID

		return
	}

	chatJID := jid.ToNonAD()
	if chatJID.Server == "lid" {
		payload["ChatLid"] = chatJID.String()
		if client != nil {
			chatJID = NormalizeJIDFromLID(ctx, chatJID, client).ToNonAD()
		}
	}

	payload["ChatID"] = chatJID.String()
}

func labelAppStateTimestamp(evt *events.AppState) string {
	if ts := evt.GetTimestamp(); ts > 0 {
		return time.UnixMilli(ts).UTC().Format(time.RFC3339)
	}

	return time.Now().UTC().Format(time.RFC3339)
}
