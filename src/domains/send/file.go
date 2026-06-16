package send

import "mime/multipart"

type FileRequest struct {
	BaseRequest
	File     *multipart.FileHeader `json:"-" form:"file"`
	FileURL  *string               `json:"file_url" form:"file_url"`
	FilePath *string               `json:"file_path" form:"file_path"` // Base64 data
	FileName string                `json:"file_name" form:"file_name"` // Custom filename
	Caption  string                `json:"caption" form:"caption"`     // Caption
}
