package rest

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	domainSend "github.com/multihubandredye-prog/W-C-A_GO/domains/send"
	"github.com/multihubandredye-prog/W-C-A_GO/infrastructure/whatsapp"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/utils"
)

// SendCarousel is isolated from existing send endpoints for easy removal.
func (controller *Send) SendCarousel(c *fiber.Ctx) error {
	var request domainSend.CarouselRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Invalid request"})
	}
	if strings.TrimSpace(request.Phone) == "" || strings.TrimSpace(request.ImagePaths) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "phone and image_paths are required"})
	}
	sender, ok := controller.Service.(domainSend.ICarouselSender)
	if !ok {
		return c.Status(fiber.StatusNotImplemented).JSON(utils.ResponseData{Status: 501, Code: "NOT_IMPLEMENTED", Message: "Carousel sender unavailable"})
	}
	utils.SanitizePhone(&request.Phone)
	response, err := sender.SendCarousel(whatsapp.ContextWithDevice(c.UserContext(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)
	return c.JSON(utils.ResponseData{Status: 200, Code: "SUCCESS", Message: response.Status, Results: response})
}
