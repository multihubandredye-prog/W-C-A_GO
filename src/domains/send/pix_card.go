package send

// PixCardRequest describes an image card with one copy-to-clipboard button.
// It shares a key only; it does not create or confirm a Pix payment.
type PixCardRequest struct {
	BaseRequest
	ImageURL string `json:"image_url" form:"image_url"`
	PixKey   string `json:"pix_key" form:"pix_key"`
	Body     string `json:"body,omitempty" form:"body"`
}
