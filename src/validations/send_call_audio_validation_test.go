package validations

import (
	"bytes"
	"context"
	"mime/multipart"
	"testing"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// uploadedFileHeader builds a *multipart.FileHeader as Fiber's BodyParser
// would produce it for a multipart/form-data request with an "audio" field.
func uploadedFileHeader(t *testing.T) *multipart.FileHeader {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("audio", "alerta.mp3")
	require.NoError(t, err)
	_, err = part.Write([]byte("ID3 fake mp3 payload"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	reader := multipart.NewReader(bytes.NewReader(buf.Bytes()), writer.Boundary())
	form, err := reader.ReadForm(1 << 20)
	require.NoError(t, err)
	require.Len(t, form.File["audio"], 1)
	return form.File["audio"][0]
}

// TestValidateSendCallAudioSources covers the mutual exclusion of the three
// audio sources: uploaded file (multipart), audio_path and audio_url.
func TestValidateSendCallAudioSources(t *testing.T) {
	baseRequest := func() domainSend.CallRequest {
		return domainSend.CallRequest{
			BaseRequest: domainSend.BaseRequest{Phone: "5588999999999", Duration: intPtr(30)},
		}
	}

	t.Run("only the uploaded file is accepted", func(t *testing.T) {
		request := baseRequest()
		request.Audio = uploadedFileHeader(t)
		assert.NoError(t, ValidateSendCall(context.Background(), request))
	})

	t.Run("file and url together are refused", func(t *testing.T) {
		request := baseRequest()
		request.Audio = uploadedFileHeader(t)
		request.AudioURL = "https://example.com/a.mp3"
		assert.Equal(t,
			pkgError.ValidationError("only one audio source is allowed: either the uploaded file (audio), audio_path or audio_url"),
			ValidateSendCall(context.Background(), request))
	})

	t.Run("file and path together are refused", func(t *testing.T) {
		request := baseRequest()
		request.Audio = uploadedFileHeader(t)
		request.AudioPath = "data:audio/mp3;base64,//uQx"
		assert.Equal(t,
			pkgError.ValidationError("only one audio source is allowed: either the uploaded file (audio), audio_path or audio_url"),
			ValidateSendCall(context.Background(), request))
	})

	t.Run("url alone keeps working", func(t *testing.T) {
		request := baseRequest()
		request.AudioURL = "https://example.com/a.mp3"
		assert.NoError(t, ValidateSendCall(context.Background(), request))
	})

	t.Run("path alone keeps working", func(t *testing.T) {
		request := baseRequest()
		request.AudioPath = "data:audio/mp3;base64,//uQx"
		assert.NoError(t, ValidateSendCall(context.Background(), request))
	})
}
