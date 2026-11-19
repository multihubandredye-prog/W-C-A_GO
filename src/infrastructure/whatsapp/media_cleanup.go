package whatsapp

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/multihubandredye-prog/W-C-A_GO/config"
	"github.com/sirupsen/logrus"
)

// isDisposableMedia limits cleanup to known media extensions. Database files and
// SQLite sidecars must never be removed, even when stored in the media folders.
func isDisposableMedia(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tif", ".tiff", ".mp4", ".m4v", ".mov", ".webm", ".avi", ".mp3", ".ogg", ".opus", ".wav", ".m4a", ".aac", ".pdf":
		return true
	default:
		return false
	}
}

// isConfiguredDatabase protects configured SQLite databases even if their
// filenames use a media extension, along with their SQLite companion files.
func isConfiguredDatabase(path string, uris ...string) bool {
	candidate, err := filepath.Abs(path)
	if err != nil {
		return true
	}
	for _, raw := range uris {
		raw = strings.Trim(strings.TrimSpace(raw), `"'`)
		if !strings.HasPrefix(raw, "file:") {
			continue
		}
		name := strings.SplitN(strings.TrimPrefix(raw, "file:"), "?", 2)[0]
		if name == "" || name == ":memory:" {
			continue
		}
		if decoded, err := url.PathUnescape(name); err == nil {
			name = decoded
		}
		db, err := filepath.Abs(name)
		if err != nil {
			return true
		}
		if candidate == db || candidate == db+"-wal" || candidate == db+"-shm" || candidate == db+"-journal" {
			return true
		}
	}
	return false
}

// RunMediaCleanup removes only known media files after the configured age.
func RunMediaCleanup() {
	go func() {
		// Wait for a short moment to ensure the current operation finishes
		time.Sleep(5 * time.Second)

		// Use relative paths from config
		dirs := []string{
			config.PathMedia,
			config.PathStorages,
		}

		// Calculate the threshold for deletion based on configuration (in seconds)
		threshold := time.Duration(config.AutoDeleteMediaDuration) * time.Second

		for _, dir := range dirs {
			files, err := os.ReadDir(dir)
			if err != nil {
				logrus.Warnf("Failed to read directory for cleanup: %s: %v", dir, err)
				continue
			}

			for _, file := range files {
				if file.IsDir() {
					continue
				}

				info, err := file.Info()
				if err != nil {
					logrus.Warnf("Failed to get file info: %s: %v", file.Name(), err)
					continue
				}

				// Only delete files that are older than the threshold
				if time.Since(info.ModTime()) < threshold {
					continue
				}

				name := file.Name()
				if !isDisposableMedia(name) || isConfiguredDatabase(filepath.Join(dir, name), config.DBURI, config.DBKeysURI, config.ChatStorageURI) {
					continue
				}

				// Remove the file
				err = os.Remove(filepath.Join(dir, name))
				if err != nil {
					logrus.Warnf("Failed to remove file during cleanup: %s: %v", name, err)
				} else {
					logrus.Debugf("Cleaned up file: %s", name)
				}
			}
		}
	}()
}
