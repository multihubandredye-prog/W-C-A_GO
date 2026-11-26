package whatsapp

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanupDownloadedMediaOnlyAfterRetention(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	old := "1234567890-12345678-1234-1234-1234-123456789abc.jpg"
	recent := "1234567891-12345678-1234-1234-1234-123456789abc.png"
	db := "1234567892-12345678-1234-1234-1234-123456789abc.jpg"
	for _, name := range []string{old, recent, db, "user-photo.jpg"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
		timestamp := now.Add(-time.Minute)
		if name == recent {
			timestamp = now.Add(-10 * time.Second)
		}
		if err := os.Chtimes(path, timestamp, timestamp); err != nil {
			t.Fatal(err)
		}
	}
	cleanupDownloadedMedia(now, []string{dir}, 30*time.Second, "file:"+filepath.Join(dir, db))
	if _, err := os.Stat(filepath.Join(dir, old)); !os.IsNotExist(err) {
		t.Fatal("old downloaded media not deleted")
	}
	for _, name := range []string{recent, db, "user-photo.jpg"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatal(name, err)
		}
	}
	cleanupDownloadedMedia(now.Add(25*time.Second), []string{dir}, 30*time.Second, "file:"+filepath.Join(dir, db))
	if _, err := os.Stat(filepath.Join(dir, recent)); !os.IsNotExist(err) {
		t.Fatal("recent file was not deleted after expiry")
	}
}
