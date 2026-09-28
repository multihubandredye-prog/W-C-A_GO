package usecase

import (
	"testing"

	"github.com/purpshell/meowcaller"
)

// TestEndlessSilenceReader locks the warm-up source contract: infinite zero
// bytes (raw s16le silence) and a no-op Close.
func TestEndlessSilenceReader(t *testing.T) {
	var silence endlessSilence

	buf := make([]byte, 64)
	n, err := silence.Read(buf)
	if err != nil {
		t.Fatalf("read: unexpected error: %v", err)
	}
	if n != len(buf) {
		t.Fatalf("read: got %d bytes, want %d", n, len(buf))
	}
	for i, b := range buf {
		if b != 0 {
			t.Fatalf("read: byte %d is %d, want 0", i, b)
		}
	}

	if err := silence.Close(); err != nil {
		t.Fatalf("close: unexpected error: %v", err)
	}

	// The reader must keep yielding silence after Close (the engine decides
	// when to stop pulling frames, not the source).
	n, err = silence.Read(buf)
	if err != nil || n != len(buf) {
		t.Fatalf("read after close: n=%d err=%v", n, err)
	}
}

// TestEndlessSilenceFeedsPCMStream verifies the silence is actually delivered
// through meowcaller's PCM source: every frame is non-empty and all-zero.
func TestEndlessSilenceFeedsPCMStream(t *testing.T) {
	source := meowcaller.PCMStream(endlessSilence{})
	defer source.Close()

	for i := 0; i < 3; i++ {
		frame, err := source.ReadFrame()
		if err != nil {
			t.Fatalf("frame %d: unexpected error: %v", i, err)
		}
		if len(frame) == 0 {
			t.Fatalf("frame %d: empty frame", i)
		}
		for j, sample := range frame {
			if sample != 0 {
				t.Fatalf("frame %d sample %d: %v, want silence (0)", i, j, sample)
			}
		}
	}
}
