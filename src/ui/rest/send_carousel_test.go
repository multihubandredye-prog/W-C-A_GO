package rest

import (
	"bytes"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/multihubandredye-prog/W-C-A_GO/config"
	domainSend "github.com/multihubandredye-prog/W-C-A_GO/domains/send"
	"mime/multipart"
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

func TestCarouselMultipart(t *testing.T) {
	s := &carouselMock{}
	app := fiber.New(fiber.Config{BodyLimit: 60 * 1024 * 1024})
	InitRestSend(app, s)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("phone", "5511999999999")
	_ = writer.WriteField("body", "Teste")
	_ = writer.WriteField("card_titles", `["A","B"]`)
	for _, name := range []string{"a.jpg", "b.png"} {
		file, err := writer.CreateFormFile("images", name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = file.Write([]byte("image bytes"))
	}
	_ = writer.Close()
	req := httptest.NewRequest("POST", "/send/carousel", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 || s.calls != 1 {
		t.Fatalf("status=%d calls=%d", res.StatusCode, s.calls)
	}
	if len(s.request.Uploads) != 2 || s.request.Uploads[0].Name != "a.jpg" || string(s.request.Uploads[1].Data) != "image bytes" || len(s.request.CardTitles) != 2 {
		t.Fatalf("bad request: %+v", s.request)
	}
}
func TestCarouselMultipartRejectsTooMany(t *testing.T) {
	s := &carouselMock{}
	app := fiber.New()
	InitRestSend(app, s)
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("phone", "5511999999999")
	for i := 0; i < 11; i++ {
		f, _ := w.CreateFormFile("images", "a.jpg")
		_, _ = f.Write([]byte("x"))
	}
	_ = w.Close()
	req := httptest.NewRequest("POST", "/send/carousel", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 400 || s.calls != 0 {
		t.Fatalf("status=%d calls=%d", res.StatusCode, s.calls)
	}
}

func TestCarouselUploadUsesConfiguredImageLimit(t *testing.T) {
	old := config.WhatsappSettingMaxImageSize
	defer func() { config.WhatsappSettingMaxImageSize = old }()
	send := func(limit int64) (int, int) {
		config.WhatsappSettingMaxImageSize = limit
		s := &carouselMock{}
		app := fiber.New(fiber.Config{BodyLimit: 20 * 1024 * 1024})
		InitRestSend(app, s)
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		_ = w.WriteField("phone", "5511999999999")
		for i := 0; i < 2; i++ {
			f, _ := w.CreateFormFile("images", "a.png")
			_, _ = f.Write(bytes.Repeat([]byte("x"), 6*1024*1024))
		}
		_ = w.Close()
		req := httptest.NewRequest("POST", "/send/carousel", &b)
		req.Header.Set("Content-Type", w.FormDataContentType())
		res, err := app.Test(req, 30000)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		return res.StatusCode, s.calls
	}
	if status, calls := send(5000000); status != 400 || calls != 0 {
		t.Fatalf("limit 5MB: %d %d", status, calls)
	}
	if status, calls := send(50000000); status != 200 || calls != 1 {
		t.Fatalf("limit 50MB: %d %d", status, calls)
	}
}
