package usecase

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/disintegration/imaging"
	"github.com/multihubandredye-prog/W-C-A_GO/config"
	domainSend "github.com/multihubandredye-prog/W-C-A_GO/domains/send"
	"github.com/multihubandredye-prog/W-C-A_GO/infrastructure/whatsapp"
	pkgError "github.com/multihubandredye-prog/W-C-A_GO/pkg/error"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/utils"
	"github.com/multihubandredye-prog/W-C-A_GO/validations"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

const (
	carouselStorageRoot = "/storage/emulated/0"
	carouselMaxPixels   = 25000000
	carouselMaxCards    = 10
)

type carouselImage struct {
	data          []byte
	mimetype      string
	thumbnail     []byte
	width, height uint32
}

func imageDecodeConfig(data []byte) (image.Config, string, error) {
	return image.DecodeConfig(bytes.NewReader(data))
}

func carouselPaths(input string) ([]string, error) {
	paths := strings.Split(input, ",")
	if len(paths) < 2 || len(paths) > carouselMaxCards {
		return nil, pkgError.ValidationError("image_paths: informe entre 2 e 10 caminhos separados por vírgula")
	}
	for i := range paths {
		paths[i] = strings.TrimSpace(paths[i])
		if paths[i] == "" || !filepath.IsAbs(paths[i]) {
			return nil, pkgError.ValidationError(fmt.Sprintf("image_paths[%d]: caminho absoluto obrigatório", i+1))
		}
	}
	return paths, nil
}

// Permit only files under the shared storage directory, including after symlink resolution.
func loadCarouselImage(root, path string) (carouselImage, error) {
	var result carouselImage
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return result, fmt.Errorf("armazenamento compartilhado indisponível")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return result, fmt.Errorf("arquivo não encontrado ou inacessível")
	}
	rel, err := filepath.Rel(rootResolved, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return result, fmt.Errorf("imagem fora do armazenamento compartilhado permitido")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return result, fmt.Errorf("arquivo inacessível")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return result, fmt.Errorf("não é um arquivo regular")
	}
	if info.Size() <= 0 || info.Size() > config.WhatsappSettingMaxImageSize {
		return result, fmt.Errorf("imagem vazia ou maior que o limite de imagem configurado")
	}
	data, err := io.ReadAll(io.LimitReader(file, config.WhatsappSettingMaxImageSize+1))
	if err != nil || len(data) == 0 || int64(len(data)) > config.WhatsappSettingMaxImageSize {
		return result, fmt.Errorf("falha ao ler imagem ou arquivo acima do limite configurado")
	}
	return decodeCarouselImage(data)
}

func decodeCarouselImage(data []byte) (carouselImage, error) {
	var result carouselImage
	if len(data) == 0 || int64(len(data)) > config.WhatsappSettingMaxImageSize {
		return result, fmt.Errorf("imagem vazia ou maior que o limite de imagem configurado")
	}
	config, format, err := imageDecodeConfig(data)
	if err != nil || (format != "jpeg" && format != "png") {
		return result, fmt.Errorf("imagem inválida: use JPEG ou PNG")
	}
	if config.Width <= 0 || config.Height <= 0 || config.Width > carouselMaxPixels || config.Height > carouselMaxPixels || int64(config.Width)*int64(config.Height) > carouselMaxPixels {
		return result, fmt.Errorf("dimensões inválidas ou maiores que 25 megapixels")
	}
	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return result, fmt.Errorf("não foi possível decodificar a imagem")
	}
	thumbnail := imaging.Fit(img, 256, 256, imaging.Lanczos)
	var preview bytes.Buffer
	if err := jpeg.Encode(&preview, thumbnail, &jpeg.Options{Quality: 70}); err != nil {
		return result, fmt.Errorf("falha ao criar miniatura")
	}
	mimetype := "image/jpeg"
	if format == "png" {
		mimetype = "image/png"
	}
	return carouselImage{data: data, mimetype: mimetype, thumbnail: preview.Bytes(), width: uint32(config.Width), height: uint32(config.Height)}, nil
}

func buildCarouselMessage(body string, captions []string, buttonText string, images []*waE2E.ImageMessage, duration *int, forwarded bool) (*waE2E.Message, error) {
	if len(images) < 2 || len(images) > carouselMaxCards || len(captions) != len(images) {
		return nil, fmt.Errorf("quantidade de cartões inválida")
	}
	cards := make([]*waE2E.InteractiveMessage, 0, len(images))
	for i, image := range images {
		if image == nil {
			return nil, fmt.Errorf("imagem ausente no cartão %d", i+1)
		}
		buttons, err := buildNativeFlowButtons([]domainSend.Button{{Type: domainSend.ButtonTypeReply, Title: buttonText, ID: fmt.Sprintf("carousel_%d", i+1)}})
		if err != nil {
			return nil, err
		}
		cards = append(cards, &waE2E.InteractiveMessage{
			Header:             &waE2E.InteractiveMessage_Header{HasMediaAttachment: proto.Bool(true), Media: &waE2E.InteractiveMessage_Header_ImageMessage{ImageMessage: image}},
			Body:               &waE2E.InteractiveMessage_Body{Text: proto.String(captions[i])},
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{Buttons: buttons, MessageVersion: proto.Int32(1)}},
		})
	}
	contextInfo := &waE2E.ContextInfo{}
	if forwarded {
		contextInfo.IsForwarded = proto.Bool(true)
	}
	if duration != nil && *duration > 0 {
		contextInfo.Expiration = proto.Uint32(uint32(*duration))
	}
	kind := waE2E.InteractiveMessage_CarouselMessage_HSCROLL_CARDS
	return &waE2E.Message{InteractiveMessage: &waE2E.InteractiveMessage{
		Body: &waE2E.InteractiveMessage_Body{Text: proto.String(body)}, ContextInfo: contextInfo,
		InteractiveMessage: &waE2E.InteractiveMessage_CarouselMessage_{CarouselMessage: &waE2E.InteractiveMessage_CarouselMessage{Cards: cards, MessageVersion: proto.Int32(1), CarouselCardType: &kind}},
	}}, nil
}

// carouselCaptions preserves explicit captions; by default the displayed text
// is the local file name including its extension, never the full path.
func carouselCaptions(paths, titles []string) ([]string, error) {
	if len(titles) != 0 && len(titles) != len(paths) {
		return nil, pkgError.ValidationError("card_titles deve ter a mesma quantidade de image_paths")
	}
	captions := make([]string, len(paths))
	for i, path := range paths {
		caption := filepath.Base(path)
		if len(titles) != 0 {
			caption = strings.TrimSpace(titles[i])
		}
		if caption == "" || caption == "." || caption == string(filepath.Separator) || utf8.RuneCountInString(caption) > 160 {
			return nil, pkgError.ValidationError(fmt.Sprintf("card_titles[%d]: o nome ou texto do cartão deve ter entre 1 e 160 caracteres", i+1))
		}
		captions[i] = caption
	}
	return captions, nil
}

func (service serviceSend) SendCarousel(ctx context.Context, request domainSend.CarouselRequest) (response domainSend.GenericResponse, err error) {
	var paths []string
	if len(request.Uploads) > 0 {
		if strings.TrimSpace(request.ImagePaths) != "" || len(request.Uploads) < 2 || len(request.Uploads) > carouselMaxCards {
			return response, pkgError.ValidationError("informe entre 2 e 10 imagens, sem combinar arquivos e caminhos")
		}
		for _, upload := range request.Uploads {
			paths = append(paths, filepath.Base(strings.ReplaceAll(upload.Name, "\\", "/")))
		}
	} else {
		paths, err = carouselPaths(request.ImagePaths)
		if err != nil {
			return response, err
		}
	}
	body := strings.TrimSpace(request.Body)
	if body == "" {
		body = "Confira as imagens:"
	}
	buttonText := strings.TrimSpace(request.ButtonText)
	if buttonText == "" {
		buttonText = "Selecionar"
	}
	if utf8.RuneCountInString(body) > 1024 || utf8.RuneCountInString(buttonText) > domainSend.MaxButtonTitleLength {
		return response, pkgError.ValidationError("body ou button_text excede o limite permitido")
	}
	captions, err := carouselCaptions(paths, request.CardTitles)
	if err != nil {
		return response, err
	}
	if err = validations.ValidateSendButtons(ctx, domainSend.ButtonsRequest{BaseRequest: request.BaseRequest, Body: body, Buttons: []domainSend.Button{{Type: domainSend.ButtonTypeReply, Title: buttonText}}}); err != nil {
		return response, err
	}
	client := whatsapp.ClientFromContext(ctx)
	if client == nil {
		return response, pkgError.ErrWaCLI
	}
	recipient, err := utils.ValidateJidWithLogin(client, request.Phone)
	if err != nil {
		return response, err
	}
	// Validate every local file before uploading anything.
	localImages := make([]carouselImage, len(paths))
	for i, path := range paths {
		if len(request.Uploads) > 0 {
			localImages[i], err = decodeCarouselImage(request.Uploads[i].Data)
		} else {
			localImages[i], err = loadCarouselImage(carouselStorageRoot, path)
		}
		if err != nil {
			return response, pkgError.ValidationError(fmt.Sprintf("image_paths[%d]: %v", i+1, err))
		}
	}
	images := make([]*waE2E.ImageMessage, len(paths))
	for i, img := range localImages {
		uploaded, uploadErr := client.Upload(ctx, img.data, whatsmeow.MediaImage)
		if uploadErr != nil {
			return response, pkgError.InternalServerError(fmt.Sprintf("falha ao enviar imagem %d: %v", i+1, uploadErr))
		}
		images[i] = &waE2E.ImageMessage{URL: proto.String(uploaded.URL), DirectPath: proto.String(uploaded.DirectPath), MediaKey: uploaded.MediaKey, Mimetype: proto.String(img.mimetype), FileEncSHA256: uploaded.FileEncSHA256, FileSHA256: uploaded.FileSHA256, FileLength: proto.Uint64(uint64(len(img.data))), Width: proto.Uint32(img.width), Height: proto.Uint32(img.height), JPEGThumbnail: img.thumbnail}
	}
	msg, err := buildCarouselMessage(body, captions, buttonText, images, request.Duration, request.IsForwarded)
	if err != nil {
		return response, err
	}
	nodes := buttonsBizNode()
	result, err := service.wrapSendMessage(ctx, client, recipient, msg, body, whatsmeow.SendRequestExtra{ID: client.GenerateMessageID(), AdditionalNodes: &nodes})
	if err != nil {
		return response, err
	}
	response.MessageID = result.ID
	response.Status = fmt.Sprintf("Send carousel success %s (server timestamp: %s)", request.Phone, result.Timestamp.String())
	return response, nil
}
