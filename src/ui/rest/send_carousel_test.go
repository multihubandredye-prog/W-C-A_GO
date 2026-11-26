package rest

import (
	"context"
	"github.com/gofiber/fiber/v2"
	domainSend "github.com/multihubandredye-prog/W-C-A_GO/domains/send"
	"net/http/httptest"
	"strings"
	"testing"
)

type carouselMock struct {
	domainSend.ISendUsecase
	calls   int
	request domainSend.CarouselRequest
}

func (s *carouselMock) SendCarousel(_ context.Context, r domainSend.CarouselRequest) (domainSend.GenericResponse, error) {
	s.calls++
	s.request = r
	return domainSend.GenericResponse{MessageID: "test", Status: "accepted"}, nil
}
func TestCarouselEndpoint(t *testing.T) {
	s := &carouselMock{}
	app := fiber.New()
	InitRestSend(app, s)
	cases := []struct {
		body          string
		status, calls int
	}{{`{"phone":"5511999999999","body":"Teste","image_paths":"/storage/emulated/0/a.jpg,/storage/emulated/0/b.jpg","button_text":"Escolher"}`, 200, 1}, {`{"phone":"5511999999999"}`, 400, 1}, {`{"image_paths":"/a,/b"}`, 400, 1}, {`{"phone":`, 400, 1}}
	for _, tc := range cases {
		req := httptest.NewRequest("POST", "/send/carousel", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != tc.status || s.calls != tc.calls {
			t.Fatalf("status=%d calls=%d", res.StatusCode, s.calls)
		}
	}
	if s.request.Phone != "5511999999999@s.whatsapp.net" || s.request.ButtonText != "Escolher" || s.request.ImagePaths != "/storage/emulated/0/a.jpg,/storage/emulated/0/b.jpg" {
		t.Fatalf("unexpected request %+v", s.request)
	}
}
