package send

// CallRequest defines the JSON payload for placing a VoIP audio call.
type CallRequest struct {
	BaseRequest
	AudioPath string `json:"audio_path,omitempty" form:"audio_path"`
}

// CallResponse returns details of the placed call.
type CallResponse struct {
	CallID string `json:"call_id"`
	Status string `json:"status"`
}
