package send

type LinkRequest struct {
	BaseRequest
	Caption     string `json:"caption"`
	Link        string `json:"link"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	ImageBase64 string `json:"image_base64,omitempty"`
}
