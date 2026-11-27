package whatsapp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSeparateListsAndLegacyMigration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "p.json")
	legacy := AutoCallPolicy{Mode: "ignore", Numbers: []string{"5511111111"}}
	data, _ := json.Marshal(policyFile{Version: 1, Devices: map[string]AutoCallPolicy{"a": legacy}})
	if e := os.WriteFile(path, data, 0600); e != nil {
		t.Fatal(e)
	}
	p, e := loadDevicePolicy(path, "a")
	if e != nil || p.ListsVersion != 1 || len(p.IgnoredNumbers) != 1 || len(p.AllowedNumbers) != 0 {
		t.Fatal(p, e)
	}
	p.AllowedNumbers = []string{"5522222222"}
	if e = saveDevicePolicy(path, "a", p); e != nil {
		t.Fatal(e)
	}
	q, e := loadDevicePolicy(path, "a")
	if e != nil || len(q.AllowedNumbers) != 1 || len(q.IgnoredNumbers) != 1 {
		t.Fatal(q, e)
	}
	if yes, _ := q.Allows("5511111111", time.Now()); yes {
		t.Fatal("ignored answered")
	}
	q.Mode = "allow"
	q.Numbers = q.AllowedNumbers
	if yes, _ := q.Allows("5522222222", time.Now()); !yes {
		t.Fatal("allowed blocked")
	}
	if yes, _ := q.Allows("5511111111", time.Now()); yes {
		t.Fatal("non allowed answered")
	}
	if e = saveDevicePolicy(path, "a", legacy); e == nil {
		t.Fatal("old client overwrote lists")
	}
	q.Numbers = []string{"5511111111"}
	if e = q.Validate(); e == nil {
		t.Fatal("mismatched active numbers accepted")
	}
}
func TestLegacyAllPreservesUnclassified(t *testing.T) {
	path := filepath.Join(t.TempDir(), "p.json")
	legacy := AutoCallPolicy{Mode: "all", Numbers: []string{"5511111111"}}
	data, _ := json.Marshal(policyFile{Version: 1, Devices: map[string]AutoCallPolicy{"a": legacy}})
	os.WriteFile(path, data, 0600)
	p, e := loadDevicePolicy(path, "a")
	if e != nil || len(p.UnclassifiedNumbers) != 1 || len(p.Numbers) != 0 {
		t.Fatal(p, e)
	}
	if e = saveDevicePolicy(path, "a", p); e != nil {
		t.Fatal(e)
	}
	p, e = loadDevicePolicy(path, "a")
	if e != nil || len(p.UnclassifiedNumbers) != 1 {
		t.Fatal(p, e)
	}
}
