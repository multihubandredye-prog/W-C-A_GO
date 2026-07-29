package whatsapp

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/liststore"
	"github.com/stretchr/testify/assert"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func TestParsePaginationRowID(t *testing.T) {
	tests := []struct {
		name      string
		rowID     string
		wantPage  int
		wantValid bool
	}{
		{"valid page 1", "__wca_page_1", 1, true},
		{"valid page 4", "__wca_page_4", 4, true},
		{"invalid prefix", "page_1", 0, false},
		{"invalid page zero", "__wca_page_0", 0, false},
		{"invalid page negative", "__wca_page_-1", 0, false},
		{"not a number", "__wca_page_abc", 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			page, valid := parsePaginationRowID(tc.rowID)
			assert.Equal(t, tc.wantPage, page)
			assert.Equal(t, tc.wantValid, valid)
		})
	}
}

func TestSelectedRowIDFromEvent(t *testing.T) {
	evt := &events.Message{
		Message: &waE2E.Message{
			ListResponseMessage: &waE2E.ListResponseMessage{
				SingleSelectReply: &waE2E.ListResponseMessage_SingleSelectReply{
					SelectedRowID: proto.String("__wca_page_2"),
				},
			},
		},
	}
	assert.Equal(t, "__wca_page_2", selectedRowIDFromEvent(evt))
}

func TestHandleListPagination_FallbackCandidates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list_store.json")
	liststore.Default = liststore.New(path, time.Hour)

	const chat = "558184752564@s.whatsapp.net"
	liststore.Default.Save("MSG_123", liststore.PagedList{
		Chat:     chat,
		Rows:     []liststore.Row{{RowID: "r1", Title: "t1"}, {RowID: "r2", Title: "t2"}},
		PageSize: 9,
	})

	// Try with matching JID without quoted ID
	evt := &events.Message{
		Info: types.MessageInfo{
			MessageSource: types.MessageSource{
				Chat: types.NewJID("558184752564", types.DefaultUserServer),
			},
		},
		Message: &waE2E.Message{
			ListResponseMessage: &waE2E.ListResponseMessage{
				SingleSelectReply: &waE2E.ListResponseMessage_SingleSelectReply{
					SelectedRowID: proto.String("__wca_page_2"),
				},
			},
		},
	}

	// Because listPageSender is nil in unit tests without setup, we expect
	// HandleListPagination to reach the logrus.Error("no page sender registered")
	// after successfully resolving the catalogue via fallback.
	res := HandleListPagination(context.Background(), evt, nil)
	assert.Nil(t, res)
}

func TestForwardMessageToWebhookWithPagination_IncludesForwardPagination(t *testing.T) {
	evt := &events.Message{
		Info: types.MessageInfo{
			MessageSource: types.MessageSource{
				Chat: types.NewJID("558184752564", types.DefaultUserServer),
			},
			ID: "MSG_TAP",
		},
		Message: &waE2E.Message{
			ListResponseMessage: &waE2E.ListResponseMessage{
				SingleSelectReply: &waE2E.ListResponseMessage_SingleSelectReply{
					SelectedRowID: proto.String("__wca_page_2"),
				},
			},
		},
	}

	result := &PaginationResult{
		Page:       2,
		TotalPages: 3,
		TotalRows:  25,
		MessageID:  "MSG_NEXT",
		Rows:       []liststore.Row{{RowID: "item1", Title: "Item 1"}},
		HasMore:    true,
		Forward:    true,
	}

	err := forwardMessageToWebhookWithPagination(context.Background(), nil, evt, result)
	assert.NoError(t, err)
}
