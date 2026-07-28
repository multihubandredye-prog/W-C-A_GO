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
}
