package whatsapp

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyAutoCallAudio(t *testing.T) {
	root := t.TempDir()
	shared := filepath.Join(root, "shared")
	private := filepath.Join(root, "private")
	if e := os.MkdirAll(shared, 0700); e != nil {
		t.Fatal(e)
	}
	src := filepath.Join(shared, "um nome ' novo.mp3")
	content := []byte{0, 1, 255, 42, 10, 13}
	if e := os.WriteFile(src, content, 0600); e != nil {
		t.Fatal(e)
	}
	target, e := copyAutoCallAudio(src, shared, private)
	if e != nil {
		t.Fatal(e)
	}
	got, e := os.ReadFile(target)
	if e != nil || !bytes.Equal(got, content) {
		t.Fatal("bytes changed", e)
	}
	if orig, e := os.ReadFile(src); e != nil || !bytes.Equal(orig, content) {
		t.Fatal("source changed", e)
	}
	if _, e = copyAutoCallAudio(target, shared, private); e != nil {
		t.Fatal(e)
	}
	for _, bad := range []string{filepath.Join(shared, "missing.mp3"), filepath.Join(root, "outside.mp3"), src + ".txt"} {
		if _, e = copyAutoCallAudio(bad, shared, private); e == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
	got, _ = os.ReadFile(target)
	if !bytes.Equal(got, content) {
		t.Fatal("target changed after failure")
	}
}
func TestCopyAutoCallAudioRejectsEscapeAndOversize(t *testing.T) {
	root := t.TempDir()
	shared := filepath.Join(root, "shared")
	private := filepath.Join(root, "private")
	os.MkdirAll(shared, 0700)
	outside := filepath.Join(root, "outside.mp3")
	os.WriteFile(outside, []byte("outside"), 0600)
	link := filepath.Join(shared, "link.mp3")
	os.Symlink(outside, link)
	if _, e := copyAutoCallAudio(link, shared, private); e == nil {
		t.Fatal("symlink escape allowed")
	}
	large := filepath.Join(shared, "large.mp3")
	f, e := os.Create(large)
	if e != nil {
		t.Fatal(e)
	}
	if e = f.Truncate(maxAutoCallAudioBytes + 1); e != nil {
		t.Fatal(e)
	}
	f.Close()
	if _, e = copyAutoCallAudio(large, shared, private); e == nil {
		t.Fatal("oversized file accepted")
	}
}
func TestCopyReplacesSymlinkDestination(t *testing.T) {
	root := t.TempDir()
	shared := filepath.Join(root, "shared")
	private := filepath.Join(root, "private")
	os.MkdirAll(shared, 0700)
	os.MkdirAll(private, 0700)
	src := filepath.Join(shared, "source.mp3")
	os.WriteFile(src, []byte("binary"), 0600)
	dest := filepath.Join(private, "auto-call.mp3")
	os.Symlink(src, dest)
	if _, e := copyAutoCallAudio(src, shared, private); e != nil {
		t.Fatal(e)
	}
	info, e := os.Lstat(dest)
	if e != nil || !info.Mode().IsRegular() {
		t.Fatal("destination symlink not replaced", e)
	}
}
