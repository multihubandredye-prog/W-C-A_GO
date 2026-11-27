package rest

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/multihubandredye-prog/W-C-A_GO/infrastructure/whatsapp"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/utils"
	"strings"
)

type answerCallRequest struct {
	CallID    string `json:"call_id"`
	AudioPath string `json:"audio_path"`
}

func (controller *Send) AnswerIncomingCall(c *fiber.Ctx) error {
	var req answerCallRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.CallID) == "" || strings.TrimSpace(req.AudioPath) == "" {
		return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "BAD_REQUEST", Message: "Informe call_id e audio_path (MP3 local)"})
	}
	instance := getDeviceFromCtx(c)
	if instance == nil {
		return c.Status(400).JSON(utils.ResponseData{Status: 400, Code: "DEVICE_ID_REQUIRED", Message: "Dispositivo não selecionado"})
	}
	err := instance.AnswerIncomingAudio(req.CallID, req.AudioPath)
	if err != nil {
		status, code := 400, "BAD_REQUEST"
		if errors.Is(err, whatsapp.ErrCallUnavailable) || errors.Is(err, whatsapp.ErrCallUnsupported) || errors.Is(err, whatsapp.ErrCallConflict) || errors.Is(err, whatsapp.ErrAutoRejectEnabled) {
			status, code = 409, "CALL_UNAVAILABLE"
		}
		return c.Status(status).JSON(utils.ResponseData{Status: status, Code: code, Message: err.Error()})
	}
	return c.JSON(utils.ResponseData{Status: 200, Code: "SUCCESS", Message: "Atendimento iniciado; reprodução depende da conexão de mídia", Results: map[string]string{"call_id": strings.TrimSpace(req.CallID), "status": "answer_requested"}})
}
