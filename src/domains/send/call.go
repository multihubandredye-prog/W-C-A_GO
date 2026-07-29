package send

// CallRequest defines the JSON payload for placing a VoIP audio call.
type CallRequest struct {
	BaseRequest
	AudioPath string `json:"audio_path,omitempty" form:"audio_path"` // Local file path OR Base64 audio string
	AudioURL  string `json:"audio_url,omitempty" form:"audio_url"`   // URL to download audio file
}

// CallResponse returns details of the placed call.
type CallResponse struct {
	CallID string `json:"call_id"`
	Status string `json:"status"`
}
