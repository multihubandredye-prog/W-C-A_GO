package whatsapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const AutoCallPoliciesPath = "/data/data/net.dinglisch.android.taskerm/files/WCA/audio-call/auto-call-policies.json"

var autoCallPoliciesMu sync.Mutex

type policyFile struct {
	Version int                       `json:"version"`
	Devices map[string]AutoCallPolicy `json:"devices"`
}

// A missing file preserves existing behavior; a malformed file fails closed.
func readPolicyFile(path string) (policyFile, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		// After configuration, a missing file must not silently answer everyone.
		_, markerErr := os.Stat(path + ".enabled")
		if markerErr == nil {
			return policyFile{}, errors.New("policy file missing after configuration")
		}
		if !errors.Is(markerErr, os.ErrNotExist) {
			return policyFile{}, markerErr
		}
		return policyFile{Version: 1, Devices: map[string]AutoCallPolicy{}}, nil
	}
	if err != nil {
		return policyFile{}, err
	}
	if len(data) > 4*1024*1024 {
		return policyFile{}, errors.New("policy file too large")
	}
	var f policyFile
	if err = json.Unmarshal(data, &f); err != nil {
		return policyFile{}, err
	}
	if f.Version != 1 || f.Devices == nil {
		return policyFile{}, errors.New("invalid policy file")
	}
	for _, p := range f.Devices {
		if err = p.Validate(); err != nil {
			return policyFile{}, err
		}
	}
	return f, nil
}
func loadDevicePolicy(path, device string) (AutoCallPolicy, error) {
	if device == "" {
		return AutoCallPolicy{}, errors.New("device identity unavailable")
	}
	autoCallPoliciesMu.Lock()
	defer autoCallPoliciesMu.Unlock()
	f, err := readPolicyFile(path)
	if err != nil {
		return AutoCallPolicy{}, err
	}
	p, ok := f.Devices[device]
	if !ok {
		return AutoCallPolicy{Mode: "all", ListsVersion: 1, AllowedNumbers: []string{}, IgnoredNumbers: []string{}, UnclassifiedNumbers: []string{}}, nil
	}
	// Older files only contain one list. Its meaning is known in a filtered
	// mode; in all mode preserve it separately rather than guessing.
	if p.ListsVersion == 0 {
		p.ListsVersion = 1
		p.AllowedNumbers, p.IgnoredNumbers, p.UnclassifiedNumbers = []string{}, []string{}, []string{}
		switch p.Mode {
		case "allow":
			p.AllowedNumbers = append(p.AllowedNumbers, p.Numbers...)
		case "ignore":
			p.IgnoredNumbers = append(p.IgnoredNumbers, p.Numbers...)
		default:
			p.UnclassifiedNumbers = append(p.UnclassifiedNumbers, p.Numbers...)
			p.Numbers = []string{}
		}
	}
	return p, nil
}
func saveDevicePolicy(path, device string, p AutoCallPolicy) error {
	if device == "" {
		return errors.New("device identity unavailable")
	}
	if err := p.Validate(); err != nil {
		return err
	}
	autoCallPoliciesMu.Lock()
	defer autoCallPoliciesMu.Unlock()
	f, err := readPolicyFile(path)
	if err != nil {
		return err
	}
	if existing, ok := f.Devices[device]; ok && existing.ListsVersion == 1 && p.ListsVersion == 0 {
		return errors.New("outdated policy client cannot overwrite separate lists")
	}
	f.Devices[device] = p
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".auto-call-policy-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(data)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if e := tmp.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	marker, err := os.OpenFile(path+".enabled", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if err = marker.Sync(); err != nil {
		_ = marker.Close()
		return err
	}
	if err = marker.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
func (d *DeviceInstance) policyDeviceKey() (string, error) {
	cli := d.GetClient()
	if cli == nil || cli.Store == nil || cli.Store.ID == nil {
		return "", errors.New("device is not paired")
	}
	return cli.Store.ID.ToNonAD().String(), nil
}
func (d *DeviceInstance) GetAutoCallPolicy() (AutoCallPolicy, error) {
	key, err := d.policyDeviceKey()
	if err != nil {
		return AutoCallPolicy{}, err
	}
	return loadDevicePolicy(AutoCallPoliciesPath, key)
}
func (d *DeviceInstance) SetAutoCallPolicy(p AutoCallPolicy) error {
	key, err := d.policyDeviceKey()
	if err != nil {
		return err
	}
	return saveDevicePolicy(AutoCallPoliciesPath, key, p)
}
func allowsAutomaticCall(path, device, phone string, now time.Time) (bool, error) {
	p, err := loadDevicePolicy(path, device)
	if err != nil {
		return false, fmt.Errorf("policy unavailable: %w", err)
	}
	return p.Allows(phone, now)
}
