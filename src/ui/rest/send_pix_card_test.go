package rest

import (
	"context"
	"github.com/gofiber/fiber/v2"
	domainSend "github.com/multihubandredye-prog/W-C-A_GO/domains/send"
	"net/http/httptest"
	"strings"
	"testing"
)

type pixCardService struct {
	domainSend.ISendUsecase
	called  int
	request domainSend.ButtonsRequest
}

func (s *pixCardService) SendButtons(_ context.Context, r domainSend.ButtonsRequest) (domainSend.GenericResponse, error) {
	s.called++
	s.request = r
	return domainSend.GenericResponse{MessageID: "test-id", Status: "sent"}, nil
}
func TestPixCardEndpoint(t *testing.T) {
	service := &pixCardService{}
	app := fiber.New()
	InitRestSend(app, service)
	cases := []struct {
		body          string
		status, calls int
	}{
		{`{"phone":"5511999999999@s.whatsapp.net","image_url":"data:image/png;base64,aGVsbG8=","pix_key":"5511888888888"}`, 200, 1},
		{`{"phone":"5511999999999@s.whatsapp.net","image_url":"https://example.com/card.png","pix_key":"example@example.com","body":"Pague por Pix"}`, 200, 2},
		{`{"phone":"5511999999999","pix_key":"key"}`, 400, 2},
		{`{"phone":"5511999999999","image_url":"data:image/png;base64,aGVsbG8=","pix_key":"  "}`, 400, 2},
		{`{"phone":`, 400, 2},
	}
	for i, tc := range cases {
		req := httptest.NewRequest("POST", "/send/pix-card", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != tc.status || service.called != tc.calls {
			t.Fatalf("case %d: status %d calls %d", i, res.StatusCode, service.called)
		}
		if i == 0 {
			r := service.request
			if r.Phone != "5511999999999@s.whatsapp.net" || r.ImageURL != "data:image/png;base64,aGVsbG8=" || r.Body != "Chave Pix" || len(r.Buttons) != 1 || r.Buttons[0].Type != "copy" || r.Buttons[0].CopyCode != "5511888888888" || r.Buttons[0].Title != "Copiar chave Pix" {
				t.Fatalf("incorrect forwarding: %+v", r)
			}
		}
		if i == 1 && (service.request.Body != "Pague por Pix" || service.request.Buttons[0].CopyCode != "example@example.com") {
			t.Fatal("custom values lost")
		}
	}
}
