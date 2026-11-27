package usecase

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/purpshell/meowcaller"
)

// The Android build rewrites the exec call; keep the import valid after rewriting.
var _ = exec.ErrNotFound

// openCallAudio preserves MP3 decoding without FFmpeg; other valid audio is
// converted once to a private temporary MP3 before a call is placed.
func openCallAudio(ctx context.Context, input, tempDir string) (meowcaller.AudioSource, func(), error) {
	if source, err := meowcaller.MP3File(input); err == nil {
		return source, func() {}, nil
	}
	output, err := os.CreateTemp(tempDir, ".wca-call-audio-*.mp3")
	if err != nil {
		return nil, nil, errors.New("Unable to prepare temporary audio")
	}
	name := output.Name()
	output.Close()
	cleanup := func() { _ = os.Remove(name) }
	convCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	// The Android release workflow rewrites this call to the bundled FFmpeg wrapper.
	cmd := exec.CommandContext(convCtx, "ffmpeg", "-nostdin", "-v", "error", "-protocol_whitelist", "file,pipe", "-i", input, "-map", "0:a:0", "-vn", "-map_metadata", "-1", "-ac", "1", "-ar", "16000", "-c:a", "libmp3lame", "-b:a", "64k", "-fs", "20000001", "-f", "mp3", "-y", name)
	if err = cmd.Run(); err != nil {
		cleanup()
		return nil, nil, errors.New("Audio conversion failed; verify the format and FFmpeg availability")
	}
	info, err := os.Stat(name)
	if err != nil || info.Size() <= 0 || info.Size() > 20000000 {
		cleanup()
		return nil, nil, errors.New("Converted audio is empty or exceeds 20 MB")
	}
	source, err := meowcaller.MP3File(name)
	if err != nil {
		cleanup()
		return nil, nil, errors.New("Converted audio is invalid")
	}
	return source, cleanup, nil
}
