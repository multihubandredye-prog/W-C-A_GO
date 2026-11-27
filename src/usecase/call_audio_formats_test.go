package usecase

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCallAudioConversion(t *testing.T) {
	ff, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not available")
	}
	dir := t.TempDir()
	for _, tc := range []struct{ name, codec string }{{"sample.mp3", "libmp3lame"}, {"sample.opus", "libopus"}, {"sample.wav", "pcm_s16le"}, {"sample.m4a", "aac"}, {"sample.ogg", "libvorbis"}, {"sample.flac", "flac"}} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.name)
			cmd := exec.Command(ff, "-nostdin", "-v", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=1", "-c:a", tc.codec, "-y", path)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("fixture: %v: %s", err, out)
			}
			source, clean, err := openCallAudio(context.Background(), path, dir)
			if err != nil {
				t.Fatal(err)
			}
			clean() // An open decoder remains usable after the temporary pathname is removed.
			frame, frameErr := source.ReadFrame()
			if frameErr != nil || len(frame) == 0 {
				t.Fatalf("audio frame: %v", frameErr)
			}
			_ = source.Close()
			clean()
			matches, _ := filepath.Glob(filepath.Join(dir, ".wca-call-audio-*"))
			if len(matches) != 0 {
				t.Fatal("temporary output not removed")
			}
		})
	}
	invalid := filepath.Join(dir, "invalid.opus")
	if err := os.WriteFile(invalid, []byte("not audio"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := openCallAudio(context.Background(), invalid, dir); err == nil {
		t.Fatal("invalid file accepted")
	}
	matches, _ := filepath.Glob(filepath.Join(dir, ".wca-call-audio-*"))
	if len(matches) != 0 {
		t.Fatal("temporary output remains after failure")
	}
}
