package send

// MaxListRowsPerSection is WhatsApp's practical limit of rows inside one section.
const MaxListRowsPerSection = 10

// MaxListRows is WhatsApp's practical limit of rows across all sections.
const MaxListRows = 30

// MaxListRowTitleLength is the maximum length (in runes) of a row title.
const MaxListRowTitleLength = 24

// MaxListRowDescriptionLength is the maximum length (in runes) of a row description.
const MaxListRowDescriptionLength = 72

// ListRow is a single selectable item inside a ListSection.
type ListRow struct {
	// RowID is returned in the webhook when the user picks this row.
	// Defaults to Title when omitted.
	RowID string `json:"row_id,omitempty" form:"row_id"`
	// Title is the visible label of the row. Required.
	Title string `json:"title" form:"title"`
	// Description is optional secondary text shown under the title.
	Description string `json:"description,omitempty" form:"description"`
}

// ListSection groups related rows under an optional heading.
type ListSection struct {
	// Title is the section heading shown in the picker.
	Title string `json:"title,omitempty" form:"title"`
	// Rows holds the selectable items of this section. Required.
	Rows []ListRow `json:"rows" form:"rows"`
}

// ListRequest is the payload accepted by POST /send/list.
type ListRequest struct {
	BaseRequest
	// ButtonText is the label of the button that opens the list. Defaults to "Select".
	ButtonText string `json:"button_text,omitempty" form:"button_text"`
	// Description is the main message body. Required.
	Description string `json:"description" form:"description"`
	// Title is an optional header above the body.
	Title string `json:"title,omitempty" form:"title"`
	// Footer is an optional small text at the bottom.
	Footer string `json:"footer,omitempty" form:"footer"`
	// Sections holds one or more groups of rows. Required.
	Sections []ListSection `json:"sections" form:"sections"`

	// Paginate splits a catalogue larger than the WhatsApp limit across
	// several messages, appending a navigation row that sends the next page.
	Paginate bool `json:"paginate,omitempty" form:"paginate"`
	// PageSize is how many catalogue rows each page carries, excluding the
	// navigation row. Defaults to DefaultPageSize.
	PageSize int `json:"page_size,omitempty" form:"page_size"`
	// PaginationLabel is the text of the navigation row. Defaults to "Ver mais".
	PaginationLabel string `json:"pagination_label,omitempty" form:"pagination_label"`
	// ForwardPagination reports navigation taps to the webhook, flagged with
	// IsPagination, instead of consuming them silently.
	ForwardPagination bool `json:"forward_pagination,omitempty" form:"forward_pagination"`
}

// DefaultPageSize is the catalogue rows per page when PageSize is omitted.
// One row of the ten allowed is reserved for navigation.
const DefaultPageSize = 9

// PaginationRowPrefix marks navigation rows. A row id starting with this
// prefix is a page request, never a customer choice.
const PaginationRowPrefix = "__wca_page_"

// DefaultPaginationLabel is the navigation row text when none is supplied.
const DefaultPaginationLabel = "Ver mais"
