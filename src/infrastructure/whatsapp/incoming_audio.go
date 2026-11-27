package whatsapp

import (
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
	ErrCallUnavailable   = errors.New("chamada não está disponível ou não está mais tocando")
	ErrCallUnsupported   = errors.New("somente chamadas de áudio individuais são suportadas")
	ErrCallConflict      = errors.New("chamada já está sendo atendida")
	ErrAutoRejectEnabled = errors.New("desative a recusa automática de chamadas antes de usar esta rota")
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
	offerKnown, group, video, busy bool
	player                         *meowcaller.Player
	done                           chan struct{}
	once                           sync.Once
}
type incomingRegistry struct {
	mu    sync.Mutex
	calls map[string]*incomingEntry
}

func newIncomingRegistry() *incomingRegistry {
	return &incomingRegistry{calls: make(map[string]*incomingEntry)}
}
func (r *incomingRegistry) slot(id string) *incomingEntry {
	e := r.calls[id]
	if e == nil {
		if len(r.calls) >= 256 {
			return nil
		}
		e = &incomingEntry{done: make(chan struct{})}
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
	r.mu.Unlock()
	time.AfterFunc(90*time.Second, func() { r.expire(id, e) })
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
	r.mu.Unlock()
	c.OnEnd(func(reason string) { logrus.Infof("Incoming call %s ended: %s", id, reason); r.remove(id, e) })
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
	const maxBytes = 20 * 1000 * 1000
	path = strings.TrimSpace(path)
	if !filepath.IsAbs(path) || !strings.EqualFold(filepath.Ext(path), ".mp3") {
		return nil, errors.New("audio_path deve ser um caminho absoluto de arquivo MP3")
	}
	root, err := filepath.EvalSymlinks("/storage/emulated/0")
	if err != nil {
		return nil, errors.New("armazenamento compartilhado indisponível")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, errors.New("arquivo de áudio não encontrado")
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return nil, errors.New("áudio deve estar no armazenamento compartilhado")
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maxBytes {
		return nil, errors.New("MP3 vazio, inacessível ou maior que 20 MB")
	}
	source, err := meowcaller.MP3File(resolved)
	if err != nil {
		return nil, errors.New("MP3 inválido ou incompatível")
	}
	return source, nil
}

// answer reserves a ringing call exactly once; the returned acknowledgement does not
// imply the remote party heard audio. Media playback begins only when OnReady fires.
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
	if e.call.State() != meowcaller.CallPhaseRinging {
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
	c.OnReady(func() {
		select {
		case <-e.done:
			return
		default:
		}
		c.Subscribe(player)
		logrus.Infof("Incoming call %s ready; playing notice", id)
	})
	if err := c.Answer(); err != nil {
		r.remove(id, e)
		return fmt.Errorf("falha ao atender: %w", err)
	}
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

func (d *DeviceInstance) AnswerIncomingAudio(id, path string) error {
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
	return r.answer(strings.TrimSpace(id), src)
}
