package validations_test

import (
	"context"
	"strings"
	"testing"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/validations"
	"github.com/stretchr/testify/assert"
)

const testPhone = "628912344551"

func baseButtons() []domainSend.Button {
	return []domainSend.Button{{Type: "reply", Title: "Yes", ID: "yes"}}
}

func TestValidateSendButtons(t *testing.T) {
	tests := []struct {
		name    string
		request domainSend.ButtonsRequest
		wantErr bool
		errPart string
	}{
		{
			name: "valid single reply button",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Body:        "Confirm?",
				Buttons:     baseButtons(),
			},
		},
		{
			name: "valid three mixed buttons",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Body:        "Pick one",
				Buttons: []domainSend.Button{
					{Type: "reply", Title: "Yes", ID: "yes"},
					{Type: "cta_url", Title: "Site", URL: "https://example.com"},
					{Type: "cta_call", Title: "Call", PhoneNumber: "628912344551"},
				},
			},
		},
		{
			name: "type defaults to reply when empty",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Body:        "Hi",
				Buttons:     []domainSend.Button{{Title: "Ok"}},
			},
		},
		{
			name: "rejects more than three buttons",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Body:        "Too many",
				Buttons: []domainSend.Button{
					{Title: "A"}, {Title: "B"}, {Title: "C"}, {Title: "D"},
				},
			},
			wantErr: true,
			errPart: "maximum 3 buttons",
		},
		{
			name: "rejects empty buttons",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Body:        "No buttons",
			},
			wantErr: true,
			errPart: "cannot be blank",
		},
		{
			name: "rejects blank body",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Buttons:     baseButtons(),
			},
			wantErr: true,
			errPart: "body",
		},
		{
			name: "rejects blank phone",
			request: domainSend.ButtonsRequest{
				Body:    "Hello",
				Buttons: baseButtons(),
			},
			wantErr: true,
			errPart: "phone",
		},
		{
			name: "rejects cta_url without url",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Body:        "Visit",
				Buttons:     []domainSend.Button{{Type: "cta_url", Title: "Go"}},
			},
			wantErr: true,
			errPart: "url",
		},
		{
			name: "rejects cta_call without phone_number",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Body:        "Call",
				Buttons:     []domainSend.Button{{Type: "cta_call", Title: "Ring"}},
			},
			wantErr: true,
			errPart: "phone_number",
		},
		{
			name: "rejects copy without copy_code",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Body:        "Coupon",
				Buttons:     []domainSend.Button{{Type: "copy", Title: "Copy"}},
			},
			wantErr: true,
			errPart: "copy_code",
		},
		{
			name: "rejects unknown type",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Body:        "Unknown",
				Buttons:     []domainSend.Button{{Type: "teleport", Title: "Beam"}},
			},
			wantErr: true,
			errPart: "not supported",
		},
		{
			name: "rejects duplicated reply ids",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Body:        "Dup",
				Buttons: []domainSend.Button{
					{Type: "reply", Title: "One", ID: "same"},
					{Type: "reply", Title: "Two", ID: "same"},
				},
			},
			wantErr: true,
			errPart: "duplicated",
		},
		{
			name: "rejects blank title",
			request: domainSend.ButtonsRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Body:        "Blank",
				Buttons:     []domainSend.Button{{Type: "reply", Title: "   "}},
			},
			wantErr: true,
			errPart: "title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validations.ValidateSendButtons(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errPart != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errPart))
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateSendList(t *testing.T) {
	validSection := domainSend.ListSection{
		Title: "Drinks",
		Rows: []domainSend.ListRow{
			{RowID: "coke", Title: "Coke", Description: "350ml"},
			{RowID: "water", Title: "Water"},
		},
	}

	tests := []struct {
		name    string
		request domainSend.ListRequest
		wantErr bool
		errPart string
	}{
		{
			name: "valid single section",
			request: domainSend.ListRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Description: "Choose a drink",
				Sections:    []domainSend.ListSection{validSection},
			},
		},
		{
			name: "valid list beyond the three button limit",
			request: domainSend.ListRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Description: "Menu",
				Sections: []domainSend.ListSection{{
					Title: "Items",
					Rows: []domainSend.ListRow{
						{RowID: "1", Title: "One"}, {RowID: "2", Title: "Two"},
						{RowID: "3", Title: "Three"}, {RowID: "4", Title: "Four"},
						{RowID: "5", Title: "Five"},
					},
				}},
			},
		},
		{
			name: "rejects empty sections",
			request: domainSend.ListRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Description: "Nothing",
			},
			wantErr: true,
			errPart: "sections",
		},
		{
			name: "rejects section without rows",
			request: domainSend.ListRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Description: "Empty section",
				Sections:    []domainSend.ListSection{{Title: "Empty"}},
			},
			wantErr: true,
			errPart: "rows",
		},
		{
			name: "rejects blank description",
			request: domainSend.ListRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Sections:    []domainSend.ListSection{validSection},
			},
			wantErr: true,
			errPart: "description",
		},
		{
			name: "rejects duplicated row ids across sections",
			request: domainSend.ListRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Description: "Dup",
				Sections: []domainSend.ListSection{
					{Title: "A", Rows: []domainSend.ListRow{{RowID: "same", Title: "One"}}},
					{Title: "B", Rows: []domainSend.ListRow{{RowID: "same", Title: "Two"}}},
				},
			},
			wantErr: true,
			errPart: "duplicated",
		},
		{
			name: "rejects more than ten rows in one section",
			request: domainSend.ListRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Description: "Too many",
				Sections:    []domainSend.ListSection{{Title: "Big", Rows: makeRows(11)}},
			},
			wantErr: true,
			errPart: "maximum 10 rows per section",
		},
		{
			name: "rejects blank row title",
			request: domainSend.ListRequest{
				BaseRequest: domainSend.BaseRequest{Phone: testPhone},
				Description: "Blank row",
				Sections: []domainSend.ListSection{
					{Title: "S", Rows: []domainSend.ListRow{{RowID: "x", Title: "  "}}},
				},
			},
			wantErr: true,
			errPart: "title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validations.ValidateSendList(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errPart != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errPart))
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

// makeRows builds n rows with unique ids, used to exercise the size limits.
func makeRows(n int) []domainSend.ListRow {
	rows := make([]domainSend.ListRow, 0, n)
	for i := 0; i < n; i++ {
		id := string(rune('a' + i))
		rows = append(rows, domainSend.ListRow{RowID: id, Title: "Row " + id})
	}
	return rows
}
