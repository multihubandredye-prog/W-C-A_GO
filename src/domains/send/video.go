package send

import "mime/multipart"

type VideoRequest struct {
	BaseRequest
	Caption     string                `json:"caption" form:"caption"`
	Video       *multipart.FileHeader `json:"-" form:"video"`
	VideoURL    *string               `json:"video_url" form:"video_url"`
	VideoPath   *string               `json:"video_path" form:"video_path"` // Renamed for consistency
	ViewOnce    bool                  `json:"view_once" form:"view_once"`
	Compress    bool                  `json:"compress"`
	GifPlayback bool                  `json:"gif_playback" form:"gif_playback"`
}
