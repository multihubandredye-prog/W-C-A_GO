package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/liststore"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// paginationRowID builds the row id of the navigation entry for a page.
func paginationRowID(page int) string {
	return domainSend.PaginationRowPrefix + strconv.Itoa(page)
}

// ParsePaginationRowID reports whether a row id is a navigation entry and, if
// so, which page it requests.
func ParsePaginationRowID(rowID string) (page int, ok bool) {
	if !strings.HasPrefix(rowID, domainSend.PaginationRowPrefix) {
		return 0, false
	}
	page, err := strconv.Atoi(strings.TrimPrefix(rowID, domainSend.PaginationRowPrefix))
	if err != nil || page < 1 {
		return 0, false
	}
	return page, true
}

// sendPaginatedList stores the catalogue and sends its first page.
func (service serviceSend) sendPaginatedList(ctx context.Context, client *whatsmeow.Client, recipient types.JID, request domainSend.ListRequest) (response domainSend.GenericResponse, err error) {
	pageSize := request.PageSize
	if pageSize <= 0 {
		pageSize = domainSend.DefaultPageSize
	}

	label := strings.TrimSpace(request.PaginationLabel)
	if label == "" {
		label = domainSend.DefaultPaginationLabel
	}

	buttonText := strings.TrimSpace(request.ButtonText)
	if buttonText == "" {
		buttonText = "Select"
	}

	// Validation guarantees exactly one section when paginating.
	section := request.Sections[0]
	rows := make([]liststore.Row, 0, len(section.Rows))
	for _, row := range section.Rows {
		rowID := strings.TrimSpace(row.RowID)
		title := strings.TrimSpace(row.Title)
		if rowID == "" {
			rowID = title
		}
		rows = append(rows, liststore.Row{
			RowID:       rowID,
			Title:       title,
			Description: strings.TrimSpace(row.Description),
		})
	}

	deviceID := ""
	if device, ok := whatsapp.DeviceFromContext(ctx); ok && device != nil {
		deviceID = device.ID()
	}

	stored := liststore.PagedList{
		Chat:              recipient.String(),
		DeviceID:          deviceID,
		Title:             request.Title,
		Description:       request.Description,
		Footer:            request.Footer,
		ButtonText:        buttonText,
		SectionName:       section.Title,
		Rows:              rows,
		PageSize:          pageSize,
		PaginationLabel:   label,
		ForwardPagination: request.ForwardPagination,
	}

	ts, err := service.sendListPage(ctx, client, recipient, stored, 1)
	if err != nil {
		return response, err
	}

	response.MessageID = ts.ID
	response.Status = fmt.Sprintf("Send list success %s page 1 of %d (server timestamp: %s)",
		request.Phone, stored.TotalPages(), ts.Timestamp.String())
	return response, nil
}

// SendListPage renders and sends a single page of a stored catalogue, keeping
// the catalogue reachable from the new message so the next tap can find it.
func (service serviceSend) SendListPage(ctx context.Context, client *whatsmeow.Client, recipient types.JID, stored liststore.PagedList, page int) (whatsmeow.SendResponse, error) {
	return service.sendListPage(ctx, client, recipient, stored, page)
}

func (service serviceSend) sendListPage(ctx context.Context, client *whatsmeow.Client, recipient types.JID, stored liststore.PagedList, page int) (whatsmeow.SendResponse, error) {
	pageRows, hasMore := stored.Page(page)
	if len(pageRows) == 0 {
		return whatsmeow.SendResponse{}, fmt.Errorf("list page %d is empty", page)
	}

	totalPages := stored.TotalPages()

	protoRows := make([]*waE2E.ListMessage_Row, 0, len(pageRows)+1)
	for _, row := range pageRows {
		protoRow := &waE2E.ListMessage_Row{
			RowID: proto.String(row.RowID),
			Title: proto.String(truncateRunes(row.Title, domainSend.MaxListRowTitleLength)),
		}
		if row.Description != "" {
			protoRow.Description = proto.String(truncateRunes(row.Description, domainSend.MaxListRowDescriptionLength))
		}
		protoRows = append(protoRows, protoRow)
	}

	// The navigation entry occupies the last slot of every page but the final
	// one, and carries the page it will open.
	if hasMore {
		nextPage := page + 1
		protoRows = append(protoRows, &waE2E.ListMessage_Row{
			RowID:       proto.String(paginationRowID(nextPage)),
			Title:       proto.String(truncateRunes(stored.PaginationLabel, domainSend.MaxListRowTitleLength)),
			Description: proto.String(truncateRunes(fmt.Sprintf("Página %d de %d", nextPage, totalPages), domainSend.MaxListRowDescriptionLength)),
		})
	}

	protoSection := &waE2E.ListMessage_Section{Rows: protoRows}
	if stored.SectionName != "" {
		protoSection.Title = proto.String(stored.SectionName)
	}

	listMsg := &waE2E.ListMessage{
		Description: proto.String(stored.Description),
		ButtonText:  proto.String(stored.ButtonText),
		ListType:    waE2E.ListMessage_SINGLE_SELECT.Enum(),
		Sections:    []*waE2E.ListMessage_Section{protoSection},
	}
	if stored.Title != "" {
		listMsg.Title = proto.String(stored.Title)
	}
	if stored.Footer != "" {
		listMsg.FooterText = proto.String(stored.Footer)
	}

	msg := &waE2E.Message{
		DocumentWithCaptionMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{ListMessage: listMsg},
		},
	}

	extraNodes := listBizNode()
	ts, err := service.wrapSendMessage(ctx, client, recipient, msg, stored.Description, whatsmeow.SendRequestExtra{
		ID:              client.GenerateMessageID(),
		AdditionalNodes: &extraNodes,
	})
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}

	// Link the catalogue to the message just sent so the reply, which quotes
	// it, can locate the remaining rows. Once the last page is out there is
	// nothing left to navigate to.
	if hasMore {
		liststore.Default.Save(ts.ID, stored)
	}

	return ts, nil
}
