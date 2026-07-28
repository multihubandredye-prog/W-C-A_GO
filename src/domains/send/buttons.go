package send

// Button types supported by WhatsApp NativeFlow interactive messages.
const (
	// ButtonTypeReply is a quick reply button. Clicking it sends the button ID back.
	ButtonTypeReply = "reply"
	// ButtonTypeURL opens an external URL. Requires the URL field.
	ButtonTypeURL = "cta_url"
	// ButtonTypeCall starts a phone call. Requires the PhoneNumber field.
	ButtonTypeCall = "cta_call"
	// ButtonTypeCopy copies a code to the clipboard. Requires the CopyCode field.
	ButtonTypeCopy = "copy"
)

// MaxButtons is the maximum number of buttons WhatsApp renders in a single
// interactive message. Messages with more buttons are rejected by validation.
const MaxButtons = 3

// MaxButtonTitleLength is the maximum length (in runes) of a button label.
// Longer titles are truncated rather than rejected, mirroring WhatsApp clients.
const MaxButtonTitleLength = 20

// Button represents a single interactive button inside a ButtonsRequest.
type Button struct {
	// Type is one of: reply, cta_url, cta_call, copy. Defaults to reply.
	Type string `json:"type" form:"type"`
	// Title is the visible label, limited to 20 characters.
	Title string `json:"title" form:"title"`
	// ID is the identifier returned when a reply button is pressed.
	// Defaults to Title when omitted.
	ID string `json:"id,omitempty" form:"id"`
	// URL is the destination for cta_url buttons.
	URL string `json:"url,omitempty" form:"url"`
	// PhoneNumber is the number dialled by cta_call buttons.
	PhoneNumber string `json:"phone_number,omitempty" form:"phone_number"`
	// CopyCode is the text copied by copy buttons.
	CopyCode string `json:"copy_code,omitempty" form:"copy_code"`
}

// ButtonsRequest is the payload accepted by POST /send/buttons.
type ButtonsRequest struct {
	BaseRequest
	// Body is the main message text. Required.
	Body string `json:"body" form:"body"`
	// Title is an optional header rendered above the body.
	Title string `json:"title,omitempty" form:"title"`
	// Footer is an optional small text below the buttons.
	Footer string `json:"footer,omitempty" form:"footer"`
	// ImageURL is an optional header image (http/https URL or data:image base64).
	ImageURL string `json:"image_url,omitempty" form:"image_url"`
	// Buttons holds 1 to 3 buttons. Required.
	Buttons []Button `json:"buttons" form:"buttons"`
}
