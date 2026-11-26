package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCarouselPaths(t *testing.T) {
	for _, tc := range []struct {
		value string
		ok    bool
		count int
	}{{"/a, /b", true, 2}, {"/a", false, 0}, {"/a,,/b", false, 0}, {"/a,relative", false, 0}, {strings.Repeat("/a,", 9) + "/z", true, 10}, {strings.Repeat("/a,", 10) + "/z", false, 0}} {
		paths, err := carouselPaths(tc.value)
		if (err == nil) != tc.ok || (tc.ok && len(paths) != tc.count) {
			t.Fatalf("%q paths %v error %v", tc.value, paths, err)
		}
	}
}
func makeCarouselImage(t *testing.T, path, format string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 24, 16))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if format == "png" {
		err = png.Encode(f, img)
	} else {
		err = jpeg.Encode(f, img, nil)
	}
	if err != nil {
		t.Fatal(err)
	}
}
func TestCarouselLocalImages(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	for _, ext := range []string{"jpg", "png"} {
		path := filepath.Join(root, "imagem."+ext)
		makeCarouselImage(t, path, ext)
		got, err := loadCarouselImage(root, path)
		if err != nil {
			t.Fatal(err)
		}
		if got.width != 24 || got.height != 16 || len(got.data) == 0 || len(got.thumbnail) == 0 {
			t.Fatal("invalid image")
		}
		if _, _, err := image.Decode(bytes.NewReader(got.thumbnail)); err != nil {
			t.Fatal(err)
		}
	}
	secret := filepath.Join(outside, "out.jpg")
	makeCarouselImage(t, secret, "jpg")
	link := filepath.Join(root, "link.jpg")
	if err := os.Symlink(secret, link); err != nil {
		t.Fatal(err)
	}
	if _, err := loadCarouselImage(root, link); err == nil {
		t.Fatal("symlink escape accepted")
	}
	bad := filepath.Join(root, "bad.jpg")
	if err := os.WriteFile(bad, []byte("not an image"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadCarouselImage(root, bad); err == nil {
		t.Fatal("invalid accepted")
	}
	large := filepath.Join(root, "large.jpg")
	if err := os.WriteFile(large, make([]byte, carouselMaxImageBytes+1), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadCarouselImage(root, large); err == nil {
		t.Fatal("oversized accepted")
	}
}
func TestCarouselMessage(t *testing.T) {
	var images []*waE2E.ImageMessage
	var titles []string
	for i := 1; i <= 10; i++ {
		images = append(images, &waE2E.ImageMessage{})
		titles = append(titles, fmt.Sprintf("Imagem %d", i))
	}
	message, err := buildCarouselMessage("Confira", titles, "Selecionar", images, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	carousel := message.GetInteractiveMessage().GetCarouselMessage()
	if carousel == nil || len(carousel.GetCards()) != 10 || carousel.GetCarouselCardType() != waE2E.InteractiveMessage_CarouselMessage_HSCROLL_CARDS {
		t.Fatal("invalid carousel")
	}
	for i, card := range carousel.GetCards() {
		if !card.GetHeader().GetHasMediaAttachment() || card.GetHeader().GetImageMessage() == nil || card.GetBody().GetText() != titles[i] {
			t.Fatalf("invalid card %d", i)
		}
		buttons := card.GetNativeFlowMessage().GetButtons()
		if len(buttons) != 1 || buttons[0].GetName() != "quick_reply" {
			t.Fatal("invalid button")
		}
		var p map[string]string
		if json.Unmarshal([]byte(buttons[0].GetButtonParamsJSON()), &p) != nil || p["id"] != fmt.Sprintf("carousel_%d", i+1) || p["display_text"] != "Selecionar" {
			t.Fatalf("invalid params %v", p)
		}
	}
	if _, err := buildCarouselMessage("x", titles[:1], "x", images[:1], nil, false); err == nil {
		t.Fatal("single card accepted")
	}
}
