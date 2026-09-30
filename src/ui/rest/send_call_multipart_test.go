package rest

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSendService captures the SendCall request without touching WhatsApp.
// The embedded nil interface panics if any other method is called, which is
// fine: these tests only exercise POST /send/call.
type fakeSendService struct {
	domainSend.ISendUsecase
	captured domainSend.CallRequest
}

func (f *fakeSendService) SendCall(ctx context.Context, request domainSend.CallRequest) (domainSend.CallResponse, error) {
	f.captured = request
	return domainSend.CallResponse{CallID: "test-call", Status: "initiated"}, nil
}

func newSendCallTestApp(t *testing.T) (*fiber.App, *fakeSendService) {
	t.Helper()
	service := &fakeSendService{}
	controller := Send{Service: service}
	app := fiber.New()
	app.Post("/send/call", controller.SendCall)
	return app, service
}

func multipartCallBody(t *testing.T, audio []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writer.WriteField("phone", "558184752564@s.whatsapp.net"))
	require.NoError(t, writer.WriteField("duration", "22"))
	if audio != nil {
		part, err := writer.CreateFormFile("audio", "alerta.mp3")
		require.NoError(t, err)
		_, err = part.Write(audio)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	return body, writer.FormDataContentType()
}

// TestSendCallHandlerMultipartUpload is a regression test: Fiber's BodyParser
// only binds multipart form values, never file parts, so the handler must
// fetch the uploaded audio with c.FormFile (same as SendAudio). Before that
// fix the call was placed with Audio == nil — it rang, but played no audio.
func TestSendCallHandlerMultipartUpload(t *testing.T) {
	app, service := newSendCallTestApp(t)
	uploaded := []byte("ID3 fake mp3 payload for binding test")

	body, contentType := multipartCallBody(t, uploaded)
	req := httptest.NewRequest("POST", "/send/call", body)
	req.Header.Set("Content-Type", contentType)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	require.NotNil(t, service.captured.Audio, "uploaded audio file must reach the usecase")
	assert.Equal(t, "alerta.mp3", service.captured.Audio.Filename)
	file, err := service.captured.Audio.Open()
	require.NoError(t, err)
	got, err := io.ReadAll(file)
	require.NoError(t, err)
	assert.Equal(t, uploaded, got)
	assert.Equal(t, "558184752564@s.whatsapp.net", service.captured.Phone)
	require.NotNil(t, service.captured.Duration)
	assert.Equal(t, 22, *service.captured.Duration)
}

// TestSendCallHandlerJSON ensures the classic JSON contract (Tasker etc.)
// keeps working and no phantom file is injected.
func TestSendCallHandlerJSON(t *testing.T) {
	app, service := newSendCallTestApp(t)

	req := httptest.NewRequest("POST", "/send/call",
		bytes.NewBufferString(`{"phone":"558184752564","audio_path":"data:audio/mp3;base64,//uQx","duration":22}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	assert.Nil(t, service.captured.Audio)
	assert.Equal(t, "data:audio/mp3;base64,//uQx", service.captured.AudioPath)
	// SanitizePhone appends the user suffix to bare numbers
	assert.Equal(t, "558184752564@s.whatsapp.net", service.captured.Phone)
}
