package whatsapp

import (
	"context"
	"errors"
	"io"
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
	video         bool
	answerErr     error
	answers       atomic.Int32
	subscriptions atomic.Int32
	ready         func()
	end           func(string)
}

func (c *fakeInbound) ID() string                   { return c.id }
func (c *fakeInbound) State() meowcaller.CallPhase  { return meowcaller.CallPhaseRinging }
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
	if c.subscriptions.Load() != 0 {
		t.Fatal("must wait for ready")
	}
	c.ready()
	if c.subscriptions.Load() != 1 {
		t.Fatal("ready must subscribe")
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
