package rest

import (
	domainNewsletter "github.com/multihubandredye-prog/W-C-A_GO/domains/newsletter"
	"github.com/multihubandredye-prog/W-C-A_GO/infrastructure/whatsapp"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

type Newsletter struct {
	Service domainNewsletter.INewsletterUsecase
}

func InitRestNewsletter(app fiber.Router, service domainNewsletter.INewsletterUsecase) Newsletter {
	rest := Newsletter{Service: service}
	app.Post("/newsletter/unfollow", rest.Unfollow)
	return rest
}

func (controller *Newsletter) Unfollow(c *fiber.Ctx) error {
	var request domainNewsletter.UnfollowRequest
	err := c.BodyParser(&request)
	utils.PanicIfNeeded(err)

	err = controller.Service.Unfollow(whatsapp.ContextWithDevice(c.UserContext(), getDeviceFromCtx(c)), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Success unfollow newsletter",
	})
}
