package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/multihubandredye-prog/W-C-A_GO/config"
	"github.com/purpshell/meowcaller"
	"github.com/sirupsen/logrus"
)

var (
	ErrCallUnavailable   = errors.New("Call is unavailable or no longer ringing")
	ErrCallUnsupported   = errors.New("Only individual audio calls are supported")
	ErrCallConflict      = errors.New("Call is already being answered")
	ErrAutoRejectEnabled = errors.New("Disable automatic call rejection before using this endpoint")
)

type inboundCall interface {
	ID() string
	State() meowcaller.CallPhase
	IsVideo() bool
	OnReady(func())
	OnEnd(func(string))
	Answer() error
	Hangup() error
	Subscribe(*meowcaller.Player)
}
type incomingEntry struct {
	call                           inboundCall
	firstObserved                  time.Time
	deviceKey, callerPhone         string
	identityRequired               bool
	offerKnown, group, video, busy bool
	player                         *meowcaller.Player
	done                           chan struct{}
	once                           sync.Once
}
type incomingRegistry struct {
	mu         sync.Mutex
	calls      map[string]*incomingEntry
	changed    chan struct{}
	policyPath string
}

func newIncomingRegistry() *incomingRegistry {
	return &incomingRegistry{calls: make(map[string]*incomingEntry), changed: make(chan struct{}), policyPath: AutoCallPoliciesPath}
}

// signalLocked wakes requests that arrived before both offer handlers finished.
func (r *incomingRegistry) signalLocked() { close(r.changed); r.changed = make(chan struct{}) }

func (r *incomingRegistry) waitForCall(ctx context.Context, id string, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		r.mu.Lock()
		e := r.calls[id]
		if e != nil && e.offerKnown && e.call != nil {
			r.mu.Unlock()
			return nil
		}
		changed := r.changed
		r.mu.Unlock()
		select {
		case <-changed:
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			r.mu.Lock()
			e = r.calls[id]
			offerKnown, callRegistered := e != nil && e.offerKnown, e != nil && e.call != nil
			r.mu.Unlock()
			logrus.Warnf("Incoming call %s unavailable after waiting: offer_received=%t meowcaller_registered=%t; check device selection and meowcaller logs", id, offerKnown, callRegistered)
			return ErrCallUnavailable
		}
	}
}

func (r *incomingRegistry) slot(id string) *incomingEntry {
	e := r.calls[id]
	if e == nil {
		if len(r.calls) >= 256 {
			return nil
		}
		e = &incomingEntry{done: make(chan struct{}), firstObserved: time.Now()}
		r.calls[id] = e
	}
	return e
}
func (r *incomingRegistry) remove(id string, e *incomingEntry) {
	r.mu.Lock()
	if r.calls[id] == e {
		delete(r.calls, id)
	}
	r.mu.Unlock()
	e.once.Do(func() {
		close(e.done)
		if e.player != nil {
			e.player.Stop()
		}
	})
}
func (r *incomingRegistry) noteOffer(id string, group, video bool) {
	if id == "" {
		return
	}
	r.mu.Lock()
	e := r.slot(id)
	if e == nil {
		r.mu.Unlock()
		return
	}
	e.offerKnown = true
	e.group = group
	e.video = video
	r.signalLocked()
	r.mu.Unlock()
	time.AfterFunc(90*time.Second, func() { r.expire(id, e) })
}
func (r *incomingRegistry) noteOfferIdentity(id, deviceKey, callerPhone string, group, video bool) {
	r.mu.Lock()
	e := r.slot(id)
	if e != nil {
		e.deviceKey = deviceKey
		e.callerPhone = callerPhone
		e.identityRequired = true
	}
	r.mu.Unlock()
	r.noteOffer(id, group, video)
}
func (r *incomingRegistry) expire(id string, e *incomingEntry) {
	r.mu.Lock()
	pending := r.calls[id] == e && !e.busy
	if pending {
		delete(r.calls, id)
	}
	r.mu.Unlock()
	if pending {
		e.once.Do(func() {
			close(e.done)
			if e.player != nil {
				e.player.Stop()
			}
		})
	}
}
func (r *incomingRegistry) register(c inboundCall) {
	if c == nil || c.ID() == "" {
		return
	}
	id := c.ID()
	r.mu.Lock()
	e := r.slot(id)
	if e == nil {
		r.mu.Unlock()
		return
	}
	e.call = c
	c.OnEnd(func(reason string) { logrus.Infof("Incoming call %s ended: %s", id, reason); r.remove(id, e) })
	r.signalLocked()
	r.mu.Unlock()
	logrus.Infof("Incoming call %s registered by meowcaller", id)
	if config.WhatsappAutoAnswerCall && !config.WhatsappAutoRejectCall {
		go r.autoAnswer(id, config.WhatsappAutoAnswerAudio, openAutoAnswerMP3)
	}
	time.AfterFunc(90*time.Second, func() { r.expire(id, e) })
}
func (r *incomingRegistry) clear() {
	r.mu.Lock()
	entries := r.calls
	r.calls = make(map[string]*incomingEntry)
	r.mu.Unlock()
	for _, e := range entries {
		e.once.Do(func() {
			close(e.done)
			if e.player != nil {
				e.player.Stop()
			}
		})
	}
}

// Open only local, regular MP3 files under Android shared storage. Original files are not modified.
func openIncomingMP3(path string) (meowcaller.AudioSource, error) {
	return openMP3Under(path, "/storage/emulated/0")
}

func openAutoAnswerMP3(path string) (meowcaller.AudioSource, error) {
	return openMP3Under(path, "/data/data/net.dinglisch.android.taskerm/files/WCA/audio-call")
}

func openMP3Under(path, allowedRoot string) (meowcaller.AudioSource, error) {
	const maxBytes = 20 * 1000 * 1000
	path = strings.TrimSpace(path)
	if !filepath.IsAbs(path) || !strings.EqualFold(filepath.Ext(path), ".mp3") {
		return nil, errors.New("audio_path must be an absolute path to an MP3 file")
	}
	root, err := filepath.EvalSymlinks(allowedRoot)
	if err != nil {
		return nil, errors.New("Allowed audio directory is unavailable")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, errors.New("Audio file not found")
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return nil, errors.New("Audio file must be inside the allowed audio directory")
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maxBytes {
		return nil, errors.New("MP3 is empty, inaccessible, or larger than 20 MB")
	}
	source, err := meowcaller.MP3File(resolved)
	if err != nil {
		return nil, errors.New("Invalid or unsupported MP3")
	}
	return source, nil
}

// answer reserves a ringing call exactly once; the returned acknowledgement does not
// imply the remote party heard audio. The player is attached after Answer.
func (r *incomingRegistry) answer(id string, src meowcaller.AudioSource) error {
	transferred := false
	defer func() {
		if !transferred {
			_ = src.Close()
		}
	}()
	if config.WhatsappAutoRejectCall {
		return ErrAutoRejectEnabled
	}
	r.mu.Lock()
	e := r.calls[id]
	if e == nil || e.call == nil || !e.offerKnown {
		r.mu.Unlock()
		return ErrCallUnavailable
	}
	if e.group || e.video || e.call.IsVideo() {
		r.mu.Unlock()
		return ErrCallUnsupported
	}
	if e.busy {
		r.mu.Unlock()
		return ErrCallConflict
	}
	if phase := e.call.State(); phase != meowcaller.CallPhaseRinging && phase != meowcaller.CallPhaseConnecting {
		r.mu.Unlock()
		return ErrCallUnavailable
	}
	e.busy = true
	c := e.call
	player := meowcaller.NewPlayer()
	e.player = player
	player.Play(src)
	transferred = true
	r.mu.Unlock()
	player.OnFinish(func() {
		go func() {
			select {
			case <-time.After(500 * time.Millisecond):
				_ = c.Hangup()
				r.remove(id, e)
			case <-e.done:
			}
		}()
	})
	if err := c.Answer(); err != nil {
		r.remove(id, e)
		return fmt.Errorf("Failed to answer call: %w", err)
	}
	// The upstream example subscribes after Answer. Waiting for OnReady can
	// block outbound playback indefinitely when no inbound RTP is received.
	select {
	case <-e.done:
		return ErrCallUnavailable
	default:
	}
	c.Subscribe(player)
	logrus.Infof("Incoming call %s: answer requested and audio player attached", id)
	go func() {
		select {
		case <-e.done:
			return
		case <-time.After(5 * time.Minute):
			logrus.Warnf("Incoming call %s timed out", id)
			_ = c.Hangup()
			r.remove(id, e)
		}
	}()
	return nil
}

// autoAnswer runs separately from whatsmeow's event dispatch: the offer handler
// must be allowed to finish and publish the audio/group classification first.
func (r *incomingRegistry) autoAnswer(id, path string, open func(string) (meowcaller.AudioSource, error)) {
	if err := r.waitForCall(context.Background(), id, 4*time.Second); err != nil {
		return
	}
	r.mu.Lock()
	e := r.calls[id]
	device, phone := "", ""
	required := false
	var firstObserved time.Time
	var done <-chan struct{}
	if e != nil {
		device, phone, required = e.deviceKey, e.callerPhone, e.identityRequired
		firstObserved, done = e.firstObserved, e.done
	}
	r.mu.Unlock()
	if e == nil {
		return
	}
	if required && device == "" {
		logrus.Warn("Automatic call skipped: device identity unavailable")
		return
	}
	if device != "" {
		policy, policyErr := loadDevicePolicy(r.policyPath, device)
		if policyErr != nil {
			logrus.Warnf("Automatic call policy check failed: %v", policyErr)
			return
		}
		allowed, policyErr := policy.Allows(phone, time.Now())
		if policyErr != nil || !allowed {
			logrus.Infof("Automatic call skipped by policy")
			return
		}
		remaining := time.Until(firstObserved.Add(time.Duration(policy.AnswerDelaySeconds) * time.Second))
		if remaining > 0 {
			timer := time.NewTimer(remaining)
			select {
			case <-done:
				timer.Stop()
				return
			case <-timer.C:
			}
			// Recheck filters and schedule after the wait; a caller may have hung up.
			allowed, policyErr = allowsAutomaticCall(r.policyPath, device, phone, time.Now())
			if policyErr != nil || !allowed {
				return
			}
		}
	}
	src, err := open(path)
	if err != nil {
		logrus.Warnf("Automatic audio call %s not answered: %v", id, err)
		return
	}
	if err := r.answer(id, src); err != nil {
		logrus.Warnf("Automatic audio call %s not answered: %v", id, err)
	}
}

func (d *DeviceInstance) AnswerIncomingAudio(ctx context.Context, id, path string) error {
	if config.WhatsappAutoRejectCall {
		return ErrAutoRejectEnabled
	}
	src, err := openIncomingMP3(path)
	if err != nil {
		return err
	}
	d.mu.RLock()
	r := d.incoming
	d.mu.RUnlock()
	if r == nil {
		_ = src.Close()
		return ErrCallUnavailable
	}
	id = strings.TrimSpace(id)
	if err := r.waitForCall(ctx, id, 4*time.Second); err != nil {
		_ = src.Close()
		return err
	}
	return r.answer(id, src)
}
