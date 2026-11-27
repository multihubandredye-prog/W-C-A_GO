package whatsapp

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/purpshell/meowcaller"
)

func TestAutoCallDelayValidationAndLegacyPreservation(t *testing.T) {
	p := AutoCallPolicy{Mode: "all", ListsVersion: 1}
	for _, n := range []int{-1, 11} {
		p.AnswerDelaySeconds = n
		if p.Validate() == nil {
			t.Fatalf("accepted %d", n)
		}
	}
	for n := 0; n <= 10; n++ {
		p.AnswerDelaySeconds = n
		if err := p.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "policy.json")
	p.AnswerDelaySeconds = 7
	if err := saveDevicePolicy(path, "device", p); err != nil {
		t.Fatal(err)
	}
	p.AnswerDelaySeconds = 0
	if err := saveDevicePolicyPreservingDelay(path, "device", p, true); err != nil {
		t.Fatal(err)
	}
	saved, err := loadDevicePolicy(path, "device")
	if err != nil || saved.AnswerDelaySeconds != 7 {
		t.Fatalf("legacy client lost delay: %+v %v", saved, err)
	}
	if err = saveDevicePolicy(path, "device", p); err != nil {
		t.Fatal(err)
	}
	saved, err = loadDevicePolicy(path, "device")
	if err != nil || saved.AnswerDelaySeconds != 0 {
		t.Fatal("new client cannot reset delay")
	}
}

func TestAutoCallDelayAndCancellation(t *testing.T) {
	for _, scenario := range []string{"wait", "hangup", "policy-change"} {
		t.Run(scenario, func(t *testing.T) {
			r := newIncomingRegistry()
			r.policyPath = filepath.Join(t.TempDir(), "policy.json")
			policy := AutoCallPolicy{Mode: "all", ListsVersion: 1, AnswerDelaySeconds: 1}
			if err := saveDevicePolicy(r.policyPath, "device", policy); err != nil {
				t.Fatal(err)
			}
			c := &fakeInbound{id: scenario}
			r.noteOfferIdentity(scenario, "device", "5581999999999", false, false)
			r.register(c)
			started := time.Now()
			done := make(chan struct{})
			source := &fakeAudio{}
			go func() {
				r.autoAnswer(scenario, "unused", func(string) (meowcaller.AudioSource, error) { return source, nil })
				close(done)
			}()
			time.Sleep(150 * time.Millisecond)
			if c.answers.Load() != 0 {
				t.Fatal("answered before delay")
			}
			if scenario == "hangup" {
				c.end("hangup")
			}
			if scenario == "policy-change" {
				policy.Mode = "allow"
				policy.Numbers = []string{}
				if err := saveDevicePolicy(r.policyPath, "device", policy); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				t.Fatal("wait did not end")
			}
			if scenario == "wait" {
				if c.answers.Load() != 1 || time.Since(started) < 800*time.Millisecond {
					t.Fatal("delay or answer incorrect")
				}
				c.end("done")
			} else if c.answers.Load() != 0 {
				t.Fatal("answered after cancellation or policy change")
			}
		})
	}
}

func TestAutoCallDelayWaitRejectsInvalidPolicy(t *testing.T) {
	r := newIncomingRegistry()
	r.policyPath = filepath.Join(t.TempDir(), "missing.json")
	c := &fakeInbound{id: "missing-device"}
	r.noteOfferIdentity(c.id, "", "", false, false)
	r.register(c)
	r.autoAnswer(c.id, "unused", func(string) (meowcaller.AudioSource, error) { return nil, errors.New("unexpected open") })
	if c.answers.Load() != 0 {
		t.Fatal("missing identity answered")
	}
}
