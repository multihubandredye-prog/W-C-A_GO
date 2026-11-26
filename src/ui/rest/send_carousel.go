package rest

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/multihubandredye-prog/W-C-A_GO/config"
	domainSend "github.com/multihubandredye-prog/W-C-A_GO/domains/send"
	"github.com/multihubandredye-prog/W-C-A_GO/infrastructure/whatsapp"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/utils"
)

// SendCarousel is isolated from existing send endpoints for easy removal.
func (controller *Send) SendCarousel(c *fiber.Ctx) error {
	var request domainSend.CarouselRequest
	if strings.HasPrefix(c.Get("Content-Type"), "multipart/form-data") {
		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Formulário inválido"})
		}
		files := form.File["images"]
		if len(files) < 2 || len(files) > 10 {
			return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Selecione de 2 a 10 imagens"})
		}
		request.Phone = c.FormValue("phone")
		request.Body = c.FormValue("body")
		request.ButtonText = c.FormValue("button_text")
		if titles := c.FormValue("card_titles"); titles != "" {
			if err := json.Unmarshal([]byte(titles), &request.CardTitles); err != nil {
				return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Títulos inválidos"})
			}
		}
		for _, file := range files {
			if file.Size <= 0 || file.Size > config.WhatsappSettingMaxImageSize {
				return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Cada imagem deve respeitar o limite configurado na API"})
			}
			f, err := file.Open()
			if err != nil {
				return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Não foi possível ler a imagem"})
			}
			data, readErr := io.ReadAll(io.LimitReader(f, config.WhatsappSettingMaxImageSize+1))
			f.Close()
			if readErr != nil || len(data) == 0 || int64(len(data)) > config.WhatsappSettingMaxImageSize {
				return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Não foi possível ler a imagem ou ela excede o limite configurado"})
			}
			request.Uploads = append(request.Uploads, domainSend.CarouselUpload{Name: file.Filename, Data: data})
		}
	} else if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Invalid request"})
	}
	if strings.TrimSpace(request.Phone) == "" || (strings.TrimSpace(request.ImagePaths) == "" && len(request.Uploads) == 0) {
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
