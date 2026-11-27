package whatsapp

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const AutoCallAudioDir = "/data/data/net.dinglisch.android.taskerm/files/WCA/audio-call"
const AutoCallAudioFile = AutoCallAudioDir + "/auto-call.mp3"
const maxAutoCallAudioBytes int64 = 20_000_000

// copyAutoCallAudio copies bytes to a temporary file in the private directory
// and atomically replaces the destination. Source is never modified.
func copyAutoCallAudio(source, sharedRoot, privateRoot string) (string, error) {
	source = strings.TrimSpace(source)
	if !filepath.IsAbs(source) || !strings.EqualFold(filepath.Ext(source), ".mp3") {
		return "", errors.New("audio_path must be an absolute MP3 path")
	}
	resolved, err := filepath.EvalSymlinks(source)
	if err != nil {
		return "", errors.New("audio file not found or inaccessible")
	}
	valid := false
	for _, root := range []string{sharedRoot, privateRoot} {
		rootResolved, e := filepath.EvalSymlinks(root)
		if e != nil {
			continue
		}
		rel, e := filepath.Rel(rootResolved, resolved)
		if e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && !filepath.IsAbs(rel) {
			valid = true
			break
		}
	}
	if !valid {
		return "", errors.New("audio file must be in shared storage or the private audio directory")
	}
	f, err := os.Open(resolved)
	if err != nil {
		return "", errors.New("audio file is not readable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maxAutoCallAudioBytes {
		return "", errors.New("MP3 must be a non-empty regular file at most 20 MB")
	}
	if err = os.MkdirAll(privateRoot, 0700); err != nil {
		return "", errors.New("private audio directory is unavailable")
	}
	target := filepath.Join(privateRoot, "auto-call.mp3")
	// Only skip copying if the destination itself is a regular file, not a
	// symlink pointing back out to shared storage.
	if stat, e := os.Lstat(target); e == nil && stat.Mode().IsRegular() {
		if resolvedTarget, e := filepath.EvalSymlinks(target); e == nil && resolvedTarget == resolved {
			return target, nil
		}
	}
	tmp, err := os.CreateTemp(privateRoot, ".auto-call-*.tmp")
	if err != nil {
		return "", errors.New("cannot create private audio file")
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(0600); err == nil {
		var n int64
		n, err = io.Copy(tmp, io.LimitReader(f, maxAutoCallAudioBytes+1))
		if err == nil && (n < 1 || n > maxAutoCallAudioBytes || n != info.Size()) {
			err = errors.New("audio changed or exceeds size limit")
		}
	}
	if err == nil {
		err = tmp.Sync()
	}
	if e := tmp.Close(); err == nil {
		err = e
	}
	if err != nil {
		return "", errors.New("could not copy audio file")
	}
	if err = os.Rename(name, target); err != nil {
		return "", errors.New("could not save private audio file")
	}
	return target, nil
}
func CopyAutoCallAudio(source string) (string, error) {
	return copyAutoCallAudio(source, "/storage/emulated/0", AutoCallAudioDir)
}
