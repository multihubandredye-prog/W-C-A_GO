package send

import "context"

// CarouselRequest is a separate, experimental image carousel endpoint.
// ImagePaths is a comma-separated list of absolute Android shared-storage paths.
type CarouselRequest struct {
	BaseRequest
	Body       string   `json:"body"`
	ImagePaths string   `json:"image_paths"`
	CardTitles []string `json:"card_titles,omitempty"`
	ButtonText string   `json:"button_text,omitempty"`
}

// ICarouselSender keeps the optional feature independent of ISendUsecase.
type ICarouselSender interface {
	SendCarousel(ctx context.Context, request CarouselRequest) (GenericResponse, error)
}
