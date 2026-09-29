package send

import "mime/multipart"

// CallRequest defines the payload for placing a VoIP audio call. It accepts
// JSON (audio_url / audio_path) and multipart/form-data (audio file upload).
type CallRequest struct {
	BaseRequest
	AudioPath string `json:"audio_path,omitempty" form:"audio_path"` // Local file path OR Base64 audio string
	AudioURL  string `json:"audio_url,omitempty" form:"audio_url"`   // URL to download audio file

	// Audio is an MP3 file uploaded with multipart/form-data (field "audio"),
	// e.g. from the embedded web UI. Mutually exclusive with AudioPath and
	// AudioURL.
	Audio *multipart.FileHeader `json:"-" form:"audio"`
}

// CallResponse returns details of the placed call.
type CallResponse struct {
	CallID string `json:"call_id"`
	Status string `json:"status"`
}
