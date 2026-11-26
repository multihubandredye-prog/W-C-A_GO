package whatsapp

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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

// Only files written by storeExtractedMedia (timestamp-UUID.extension) are
// eligible. Never sweep arbitrary files supplied by the user.
var downloadedMediaName = regexp.MustCompile(`^[0-9]+-[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\.[^/\\]+$`)
var cleanupOnce sync.Once

func cleanupDownloadedMedia(now time.Time, dirs []string, age time.Duration, databaseURIs ...string) {
	for _, dir := range dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				logrus.Warnf("Failed to read cleanup directory %s: %v", dir, err)
			}
			continue
		}
		for _, file := range files {
			if !downloadedMediaName.MatchString(file.Name()) || file.Type()&os.ModeSymlink != 0 {
				continue
			}
			info, err := file.Info()
			if err != nil || !info.Mode().IsRegular() || now.Sub(info.ModTime()) < age {
				continue
			}
			path := filepath.Join(dir, file.Name())
			if isConfiguredDatabase(path, databaseURIs...) {
				continue
			}
			if err := os.Remove(path); err != nil {
				logrus.Warnf("Failed to remove downloaded media %s: %v", file.Name(), err)
			}
		}
	}
}

// RunMediaCleanup starts one process-wide periodic worker. A file is removed
// only once old enough; polling may add up to five seconds beyond the duration.
func RunMediaCleanup() {
	if !config.AutoDeleteMedia || !config.WhatsappAutoDownloadMedia {
		return
	}
	cleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				if config.AutoDeleteMedia && config.WhatsappAutoDownloadMedia {
					seconds := config.AutoDeleteMediaDuration
					if seconds < 0 {
						seconds = 0
					}
					cleanupDownloadedMedia(time.Now(), []string{config.PathMedia, config.PathStorages}, time.Duration(seconds)*time.Second, config.DBURI, config.DBKeysURI, config.ChatStorageURI)
				}
				<-ticker.C
			}
		}()
	})
}
