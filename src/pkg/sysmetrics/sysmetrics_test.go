package sysmetrics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectCountsFolderFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sample.bin"), []byte("12345"), 0600); err != nil {
		t.Fatal(err)
	}
	folderMu.Lock()
	folderCache = nil
	folderCached = folderCached.Add(-folderCacheFor)
	folderMu.Unlock()
	snapshot := Collect(map[string]string{"sample": dir}, dir)
	if snapshot.Folders["sample"].Files != 1 || snapshot.Folders["sample"].Bytes != 5 {
		t.Fatalf("unexpected folder metrics: %+v", snapshot.Folders["sample"])
	}
	if snapshot.UptimeSeconds < 0 || snapshot.GoVersion == "" {
		t.Fatalf("invalid system snapshot: %+v", snapshot)
	}
}
