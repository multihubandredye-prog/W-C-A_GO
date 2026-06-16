package whatsapp

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/sirupsen/logrus"
)

// RunMediaCleanup scans directories and removes files not matching exclusion criteria automatically.
func RunMediaCleanup() {
	go func() {
		// Wait for 30 seconds as requested
		time.Sleep(30 * time.Second)

		// Use relative paths from config
		dirs := []string{
			config.PathMedia,
			config.PathStorages,
		}

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

				name := file.Name()
				ext := strings.ToLower(filepath.Ext(name))

				// Exclusion criteria: keep .json and .db files
				if ext == ".json" || ext == ".db" {
					continue
				}

				// Remove the file
				err := os.Remove(filepath.Join(dir, name))
				if err != nil {
					logrus.Warnf("Failed to remove file during cleanup: %s: %v", name, err)
				} else {
					logrus.Debugf("Cleaned up file: %s", name)
				}
			}
		}
	}()
}
