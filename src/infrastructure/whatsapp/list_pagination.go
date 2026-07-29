package whatsapp

import (
	"context"
	"strconv"
	"strings"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/liststore"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// PageSender delivers one page of a stored catalogue. The usecase layer
// registers the real implementation at startup; keeping it behind a function
// variable avoids an import cycle between infrastructure and usecase.
type PageSender func(ctx context.Context, client *whatsmeow.Client, recipient types.JID, stored liststore.PagedList, page int) (whatsmeow.SendResponse, error)

var listPageSender PageSender

// RegisterListPageSender wires the paginated list sender used to answer
// navigation taps.
func RegisterListPageSender(sender PageSender) {
	listPageSender = sender
}

// PaginationResult describes the outcome of a navigation tap, so the webhook
// payload can report which page was opened and what it contained.
type PaginationResult struct {
	Page       int
	TotalPages int
	TotalRows  int
	MessageID  string
	Rows       []liststore.Row
	HasMore    bool
	// Forward reports whether the customer configured this catalogue to
	// surface navigation taps to the webhook.
	Forward bool
}

// parsePaginationRowID reports whether a row id requests another page.
func parsePaginationRowID(rowID string) (int, bool) {
	if !strings.HasPrefix(rowID, domainSend.PaginationRowPrefix) {
		return 0, false
	}
	page, err := strconv.Atoi(strings.TrimPrefix(rowID, domainSend.PaginationRowPrefix))
	if err != nil || page < 1 {
		return 0, false
	}
	return page, true
}

// selectedRowIDFromEvent extracts the row id a list reply carries.
func selectedRowIDFromEvent(evt *events.Message) string {
	msg := utils.UnwrapMessage(evt.Message)
	if reply := utils.FindInteractiveReply(msg); reply != nil {
		msg = reply
	}
	if listResponse := msg.GetListResponseMessage(); listResponse != nil {
		return listResponse.GetSingleSelectReply().GetSelectedRowID()
	}
	return ""
}

// quotedMessageID returns the id of the message the reply quotes, which is the
// page the customer was looking at and therefore the key of the catalogue.
func quotedMessageID(evt *events.Message) string {
	msg := utils.UnwrapMessage(evt.Message)
	if reply := utils.FindInteractiveReply(msg); reply != nil {
		msg = reply
	}
	if ctxInfo := utils.ExtractContextInfo(msg); ctxInfo != nil {
		return ctxInfo.GetStanzaID()
	}
	return ""
}

// HandleListPagination answers a navigation tap by sending the requested page.
//
// It returns a result when the tap was a page request, so the caller can
// enrich the webhook payload, and nil for ordinary selections. Sending the
// next page never depends on the forward setting: that only decides whether
// the tap is reported.
func HandleListPagination(ctx context.Context, evt *events.Message, client *whatsmeow.Client) *PaginationResult {
	if evt == nil || client == nil {
		return nil
	}

	selectedID := selectedRowIDFromEvent(evt)
	if selectedID == "" {
		return nil
	}

	page, isPagination := parsePaginationRowID(selectedID)
	if !isPagination {
		return nil
	}

	logrus.Infof("[LIST_PAGINATION] navigation tap %q from %s requesting page %d", selectedID, evt.Info.Chat, page)

	recipientJID := evt.Info.Chat.ToNonAD()

	// The tap normally quotes the page it came from, which identifies the
	// catalogue. Not every client sends that quote, so fall back to the newest
	// pending catalogue of this chat rather than leaving the tap unanswered.
	sourceID := quotedMessageID(evt)

	var stored liststore.PagedList
	var ok bool

	if sourceID != "" {
		stored, ok = liststore.Default.Get(sourceID)
	}

	if !ok {
		var fallbackKey string
		fallbackKey, stored, ok = liststore.Default.GetLatestForChat(recipientJID.String())
		if ok {
			logrus.Debugf("[LIST_PAGINATION] quoted id %q unusable, resolved catalogue by chat (%s)", sourceID, fallbackKey)
			sourceID = fallbackKey
		}
	}

	if !ok {
		logrus.Warnf("[LIST_PAGINATION] no catalogue pending for %s (quoted id %q); it may have expired", recipientJID, sourceID)
		return nil
	}

	if listPageSender == nil {
		logrus.Error("[LIST_PAGINATION] no page sender registered; cannot deliver the next page")
		return nil
	}

	recipient := recipientJID
	ts, err := listPageSender(ctx, client, recipient, stored, page)
	if err != nil {
		logrus.Errorf("[LIST_PAGINATION] failed to send page %d to %s: %v", page, recipient, err)
		return nil
	}

	rows, hasMore := stored.Page(page)
	logrus.Infof("[LIST_PAGINATION] sent page %d of %d to %s", page, stored.TotalPages(), recipient)

	// The previous page is no longer the head of the catalogue.
	liststore.Default.Delete(sourceID)

	return &PaginationResult{
		Page:       page,
		TotalPages: stored.TotalPages(),
		TotalRows:  len(stored.Rows),
		MessageID:  ts.ID,
		Rows:       rows,
		HasMore:    hasMore,
		Forward:    stored.ForwardPagination,
	}
}

// forwardMessageToWebhookWithPagination delivers the event enriched with the
// page that was just sent, so the consumer knows which page opened and which
// rows it carried.
func forwardMessageToWebhookWithPagination(ctx context.Context, client *whatsmeow.Client, evt *events.Message, result *PaginationResult) error {
	webhookEvent, err := createWebhookEvent(ctx, client, evt)
	if err != nil {
		return err
	}

	if webhookEvent.Payload != nil && result != nil {
		// Surface the page details on the unified reply object so a consumer
		// reading InteractiveReply sees everything in one place.
		for _, key := range []string{"InteractiveReply", "ListReply"} {
			if reply, ok := webhookEvent.Payload[key].(map[string]any); ok {
				reply["IsPagination"] = true
				reply["Page"] = result.Page
				reply["TotalPages"] = result.TotalPages
				reply["TotalRows"] = result.TotalRows
			}
		}

		rows := make([]map[string]any, 0, len(result.Rows))
		for _, row := range result.Rows {
			entry := map[string]any{
				"RowID": row.RowID,
				"Title": row.Title,
			}
			if row.Description != "" {
				entry["Description"] = row.Description
			}
			rows = append(rows, entry)
		}

		webhookEvent.Payload["PaginationSent"] = map[string]any{
			"MessageID":  result.MessageID,
			"Page":       result.Page,
			"TotalPages": result.TotalPages,
			"TotalRows":  result.TotalRows,
			"RowsCount":  len(rows),
			"HasMore":    result.HasMore,
			"Rows":       rows,
		}
	}

	payload := map[string]any{
		"Event":    webhookEvent.Event,
		"DeviceID": webhookEvent.DeviceID,
		"Payload":  webhookEvent.Payload,
	}

	return forwardPayloadToConfiguredWebhooks(ctx, payload, webhookEvent.Event)
}
