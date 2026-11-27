package rest

import (
	"github.com/gofiber/fiber/v2"
	"testing"
)

func TestCallRoutes(t *testing.T) {
	app := fiber.New()
	InitRestSend(app, nil)
	var outgoing, removed bool
	for _, route := range app.GetRoutes() {
		if route.Method == fiber.MethodPost && route.Path == "/send/call" {
			outgoing = true
		}
		if route.Method == fiber.MethodPost && route.Path == "/call/answer" {
			removed = true
		}
	}
	if !outgoing {
		t.Error("outgoing call route was removed")
	}
	if removed {
		t.Error("manual answer route is still registered")
	}
}
