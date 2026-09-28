package whatsapp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
)

// captureLogrus redirects the global logrus output for the duration of fn and
// returns everything that was written.
func captureLogrus(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	original := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(original)
	fn()
	return buf.String()
}

// TestMeowcallerLogBridgeRoutesLevels verifies the zerolog -> logrus bridge:
// meowcaller's JSON lines must surface as [meowcaller] entries with the right
// severity, otherwise call failures stay invisible.
func TestMeowcallerLogBridgeRoutesLevels(t *testing.T) {
	bridge := meowcallerLogBridge{}

	lines := []struct {
		json    string
		wantSub string
	}{
		{`{"level":"info","message":"first RTP sent to relay"}`, "first RTP sent to relay"},
		{`{"level":"warn","message":"relay retry"}`, "relay retry"},
		{`{"level":"error","message":"allocate failed","error":"timeout"}`, "allocate failed: timeout"},
	}

	for _, line := range lines {
		var out string
		out = captureLogrus(t, func() {
			if n, err := bridge.Write([]byte(line.json)); err != nil || n != len(line.json) {
				t.Errorf("write %q: n=%d err=%v", line.json, n, err)
			}
		})
		if !strings.Contains(out, "[meowcaller] "+line.wantSub) {
			t.Errorf("log for %q missing %q, got: %q", line.json, line.wantSub, out)
		}
	}

	// Garbage (non-JSON) input must not error and must still be forwarded.
	out := captureLogrus(t, func() {
		if _, err := bridge.Write([]byte("raw line\n")); err != nil {
			t.Errorf("raw write: %v", err)
		}
	})
	if !strings.Contains(out, "raw line") {
		t.Errorf("raw line not forwarded, got: %q", out)
	}
}

// TestNewMeowCallerInstallsBeforeConnect builds a caller for a fresh (not yet
// connected) whatsmeow client: the low-level call interception must install
// cleanly. If this ever fails with "raw call adapter is unavailable", a
// whatsmeow upgrade changed the internals the hook relies on and answered
// calls would stay on "Conectando..." forever.
func TestNewMeowCallerInstallsBeforeConnect(t *testing.T) {
	if NewMeowCaller(nil) != nil {
		t.Fatal("NewMeowCaller(nil) must return nil")
	}

	wa := whatsmeow.NewClient(&store.Device{}, nil)
	logs := captureLogrus(t, func() {
		if NewMeowCaller(wa) == nil {
			t.Fatal("NewMeowCaller must return a client for a non-nil whatsmeow client")
		}
	})

	if strings.Contains(logs, "raw call adapter is unavailable") {
		t.Fatalf("call interception failed to install on a fresh client: %s", logs)
	}
}

// TestDeviceInstanceExposesCaller locks the wiring: every DeviceInstance
// carries a meowcaller client bound to its whatsmeow client, so outbound
// calls reuse the pre-connect interception instead of creating a broken
// late one.
func TestDeviceInstanceExposesCaller(t *testing.T) {
	wa := whatsmeow.NewClient(&store.Device{}, nil)
	instance := NewDeviceInstance("test-device", wa, nil)

	if instance.GetCaller() == nil {
		t.Fatal("DeviceInstance must expose a meowcaller client")
	}

	// Swapping the underlying client must rebind the caller too.
	wa2 := whatsmeow.NewClient(&store.Device{}, nil)
	instance.SetClient(wa2)
	if instance.GetCaller() == nil {
		t.Fatal("SetClient must rebind the meowcaller client")
	}
}
