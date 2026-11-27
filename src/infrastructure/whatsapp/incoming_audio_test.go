package whatsapp

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multihubandredye-prog/W-C-A_GO/config"
	"github.com/purpshell/meowcaller"
)

type fakeAudio struct{ closed atomic.Int32 }

func (*fakeAudio) ReadFrame() ([]float32, error) { return nil, io.EOF }
func (s *fakeAudio) Close() error                { s.closed.Add(1); return nil }

type fakeInbound struct {
	id            string
	phase         *meowcaller.CallPhase
	video         bool
	answerErr     error
	answers       atomic.Int32
	subscriptions atomic.Int32
	ready         func()
	end           func(string)
}

func (c *fakeInbound) ID() string { return c.id }
func (c *fakeInbound) State() meowcaller.CallPhase {
	if c.phase != nil {
		return *c.phase
	}
	return meowcaller.CallPhaseRinging
}
func (c *fakeInbound) IsVideo() bool                { return c.video }
func (c *fakeInbound) OnReady(f func())             { c.ready = f }
func (c *fakeInbound) OnEnd(f func(string))         { c.end = f }
func (c *fakeInbound) Answer() error                { c.answers.Add(1); return c.answerErr }
func (*fakeInbound) Hangup() error                  { return nil }
func (c *fakeInbound) Subscribe(*meowcaller.Player) { c.subscriptions.Add(1) }
func TestIncomingAnswerLifecycle(t *testing.T) {
	config.WhatsappAutoRejectCall = false
	r := newIncomingRegistry()
	c := &fakeInbound{id: "a"}
	r.register(c)
	a := &fakeAudio{}
	if !errors.Is(r.answer("a", a), ErrCallUnavailable) || a.closed.Load() != 1 {
		t.Fatal("offer must be known and source closed")
	}
	r.noteOffer("a", false, false)
	s := &fakeAudio{}
	if err := r.answer("a", s); err != nil {
		t.Fatal(err)
	}
	if c.subscriptions.Load() != 1 {
		t.Fatal("player must be attached without waiting for inbound RTP")
	}
	another := &fakeAudio{}
	if !errors.Is(r.answer("a", another), ErrCallConflict) || another.closed.Load() != 1 {
		t.Fatal("duplicate answer must fail")
	}
	c.end("remote")
	if s.closed.Load() != 1 {
		t.Fatal("source not closed on end")
	}
	if !errors.Is(r.answer("a", &fakeAudio{}), ErrCallUnavailable) {
		t.Fatal("ended call still registered")
	}
}
func TestIncomingRestrictionsAndIsolation(t *testing.T) {
	config.WhatsappAutoRejectCall = false
	for _, tc := range []struct {
		name                    string
		group, video, callVideo bool
	}{{"group", true, false, false}, {"video", false, true, false}, {"call video", false, false, true}} {
		t.Run(tc.name, func(t *testing.T) {
			r := newIncomingRegistry()
			c := &fakeInbound{id: "x", video: tc.callVideo}
			r.noteOffer("x", tc.group, tc.video)
			r.register(c)
			s := &fakeAudio{}
			if !errors.Is(r.answer("x", s), ErrCallUnsupported) || s.closed.Load() != 1 || c.answers.Load() != 0 {
				t.Fatal("unsupported call was answered")
			}
		})
	}
	a, b := newIncomingRegistry(), newIncomingRegistry()
	a.noteOffer("id", false, false)
	a.register(&fakeInbound{id: "id"})
	if !errors.Is(b.answer("id", &fakeAudio{}), ErrCallUnavailable) {
		t.Fatal("device isolation failed")
	}
	config.WhatsappAutoRejectCall = true
	defer func() { config.WhatsappAutoRejectCall = false }()
	if !errors.Is(a.answer("id", &fakeAudio{}), ErrAutoRejectEnabled) {
		t.Fatal("auto reject not enforced")
	}
}
func TestConcurrentIncomingAnswers(t *testing.T) {
	config.WhatsappAutoRejectCall = false
	r := newIncomingRegistry()
	c := &fakeInbound{id: "race"}
	r.noteOffer("race", false, false)
	r.register(c)
	var wg sync.WaitGroup
	var success atomic.Int32
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if r.answer("race", &fakeAudio{}) == nil {
				success.Add(1)
			}
		}()
	}
	wg.Wait()
	if success.Load() != 1 || c.answers.Load() != 1 {
		t.Fatal("more than one answer")
	}
	c.end("done")
}
func TestIncomingAnswerFailureClosesAudio(t *testing.T) {
	r := newIncomingRegistry()
	c := &fakeInbound{id: "bad", answerErr: errors.New("no")}
	r.noteOffer("bad", false, false)
	r.register(c)
	s := &fakeAudio{}
	if r.answer("bad", s) == nil || s.closed.Load() != 1 {
		t.Fatal("failed answer did not close source")
	}
}

func TestWaitForIncomingCallHandlers(t *testing.T) {
	r := newIncomingRegistry()
	r.noteOffer("pending", false, false)
	result := make(chan error, 1)
	go func() { result <- r.waitForCall(context.Background(), "pending", time.Second) }()
	time.Sleep(20 * time.Millisecond)
	r.register(&fakeInbound{id: "pending"})
	if err := <-result; err != nil {
		t.Fatalf("wait should succeed after registration: %v", err)
	}
	if !errors.Is(r.waitForCall(context.Background(), "missing", 10*time.Millisecond), ErrCallUnavailable) {
		t.Fatal("missing call must time out")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(r.waitForCall(ctx, "missing", time.Second), context.Canceled) {
		t.Fatal("canceled request must stop waiting")
	}
}

func TestAutoAnswerAfterOfferAndRegistration(t *testing.T) {
	r := newIncomingRegistry()
	c := &fakeInbound{id: "auto"}
	r.register(c)
	done := make(chan struct{})
	src := &fakeAudio{}
	go func() {
		r.autoAnswer("auto", "unused", func(string) (meowcaller.AudioSource, error) { return src, nil })
		close(done)
	}()
	r.noteOffer("auto", false, false)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("automatic answer did not finish")
	}
	if c.answers.Load() != 1 || c.subscriptions.Load() != 1 {
		t.Fatal("automatic answer or player not started")
	}
	c.end("done")
	if src.closed.Load() != 1 {
		t.Fatal("automatic player did not close")
	}
}
func TestAutoRejectsUnsupportedCallAndInvalidSource(t *testing.T) {
	for _, tc := range []struct {
		name         string
		group, video bool
		badAudio     bool
	}{
		{"group", true, false, false}, {"video", false, true, false}, {"invalid audio", false, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newIncomingRegistry()
			c := &fakeInbound{id: tc.name}
			r.noteOffer(tc.name, tc.group, tc.video)
			r.register(c)
			src := &fakeAudio{}
			r.autoAnswer(tc.name, "path", func(string) (meowcaller.AudioSource, error) {
				if tc.badAudio {
					return nil, errors.New("bad MP3")
				}
				return src, nil
			})
			if c.answers.Load() != 0 {
				t.Fatal("unsupported or invalid audio was answered")
			}
			if !tc.badAudio && src.closed.Load() != 1 {
				t.Fatal("unsupported audio source not closed")
			}
		})
	}
}
func TestOpenMP3UnderPreventsEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	f := filepath.Join(outside, "outside.mp3")
	if err := os.WriteFile(f, []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.mp3")
	if err := os.Symlink(f, link); err != nil {
		t.Fatal(err)
	}
	if _, err := openMP3Under(link, root); err == nil || !strings.Contains(err.Error(), "allowed audio directory") {
		t.Fatal("symlink escape must fail")
	}
	if _, err := openMP3Under("https://example.com/file.mp3", root); err == nil {
		t.Fatal("URL must fail")
	}
	f = filepath.Join(root, "bad.mp3")
	if err := os.WriteFile(f, []byte("not an mp3"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := openMP3Under(f, root); err == nil {
		t.Fatal("invalid mp3 accepted")
	}
}

func TestAnswerAcceptsConnectingButNotAlreadyActive(t *testing.T) {
	for _, tc := range []struct {
		phase      meowcaller.CallPhase
		wantAnswer bool
	}{
		{meowcaller.CallPhaseConnecting, true}, {meowcaller.CallPhaseActive, false}, {meowcaller.CallPhaseEnded, false},
	} {
		r := newIncomingRegistry()
		phase := tc.phase
		c := &fakeInbound{id: "phase", phase: &phase}
		r.noteOffer("phase", false, false)
		r.register(c)
		err := r.answer("phase", &fakeAudio{})
		if (err == nil) != tc.wantAnswer {
			t.Fatalf("phase %v: unexpected answer result %v", phase, err)
		}
		r.clear()
	}
}

func TestAutoAnswerAppliesPerDevicePolicy(t *testing.T) {
	config.WhatsappAutoAnswerCall = false
	config.WhatsappAutoRejectCall = false
	path := filepath.Join(t.TempDir(), "policies.json")
	if err := saveDevicePolicy(path, "device-a", AutoCallPolicy{Mode: "allow", Numbers: []string{"5511999999999"}}); err != nil {
		t.Fatal(err)
	}
	if err := saveDevicePolicy(path, "device-b", AutoCallPolicy{Mode: "ignore", Numbers: []string{"5511999999999"}}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		device, phone string
		want          int32
	}{{"device-a", "5511999999999", 1}, {"device-a", "5511888888888", 0}, {"device-b", "5511999999999", 0}, {"device-b", "5511888888888", 1}, {"device-b", "", 0}, {"", "5511999999999", 0}} {
		r := newIncomingRegistry()
		r.policyPath = path
		c := &fakeInbound{id: "test"}
		r.noteOfferIdentity("test", tc.device, tc.phone, false, false)
		r.register(c)
		r.autoAnswer("test", "test", func(string) (meowcaller.AudioSource, error) { return &fakeAudio{}, nil })
		if got := c.answers.Load(); got != tc.want {
			t.Fatalf("device %s phone %s got %d want %d", tc.device, tc.phone, got, tc.want)
		}
		r.clear()
	}
}
