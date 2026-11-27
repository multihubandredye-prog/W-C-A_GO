package whatsapp

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAutoCallPolicyModes(t *testing.T) {
	now := time.Date(2026, 9, 14, 15, 0, 0, 0, time.UTC)
	p := AutoCallPolicy{Mode: "allow", Numbers: []string{"5511999999999"}}
	for _, tc := range []struct {
		mode, phone string
		want        bool
	}{{"all", "", true}, {"allow", "5511999999999", true}, {"allow", "5511888888888", false}, {"ignore", "5511999999999", false}, {"ignore", "5511888888888", true}, {"ignore", "", false}} {
		p.Mode = tc.mode
		got, err := p.Allows(tc.phone, now)
		if err != nil || got != tc.want {
			t.Errorf("%s %q: %v %v", tc.mode, tc.phone, got, err)
		}
	}
}
func TestAutoCallPolicyWeeklyAndOvernight(t *testing.T) {
	p := AutoCallPolicy{Mode: "all", ScheduleEnabled: true, Timezone: "America/Sao_Paulo", Days: []int{1}, Start: "22:00", End: "02:00"}
	for _, tc := range []struct {
		ts   string
		want bool
	}{{"2026-09-15T00:59:00Z", false}, {"2026-09-15T01:00:00Z", true}, {"2026-09-15T04:59:00Z", true}, {"2026-09-15T05:00:00Z", false}, {"2026-09-16T01:00:00Z", false}} {
		tm, _ := time.Parse(time.RFC3339, tc.ts)
		got, err := p.Allows("", tm)
		if err != nil || got != tc.want {
			t.Errorf("%s got %v err %v", tc.ts, got, err)
		}
	}
}
func TestAutoCallPolicyInvalidFailsClosed(t *testing.T) {
	for _, p := range []AutoCallPolicy{{Mode: "bad"}, {Mode: "allow", Numbers: []string{"123@lid"}}, {Mode: "all", ScheduleEnabled: true, Timezone: "Local", Days: []int{1}, Start: "08:00", End: "18:00"}, {Mode: "all", ScheduleEnabled: true, Timezone: "UTC", Days: []int{1}, Start: "08:00", End: "08:00"}} {
		got, err := p.Allows("", time.Now())
		if got || err == nil {
			t.Fatalf("invalid policy accepted %#v", p)
		}
	}
}
func TestCanonicalManualPhone(t *testing.T) {
	n, e := canonicalManualPhone("+55 (11) 99999-9999")
	if e != nil || n != "5511999999999" {
		t.Fatalf("bad number %q %v", n, e)
	}
	if _, e = canonicalManualPhone("abc@lid"); e == nil {
		t.Fatal("lid accepted")
	}
}

func TestAutoCallPolicyStoreIsolationAndCorruption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private", "policies.json")
	p := AutoCallPolicy{Mode: "allow", Numbers: []string{"5511999999999"}}
	if err := saveDevicePolicy(path, "device-a", p); err != nil {
		t.Fatal(err)
	}
	q, err := loadDevicePolicy(path, "device-a")
	if err != nil || q.Mode != "allow" {
		t.Fatalf("lost policy: %v", err)
	}
	other, err := loadDevicePolicy(path, "device-b")
	if err != nil || other.Mode != "all" {
		t.Fatalf("device isolation: %v", err)
	}
	if err = os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err = loadDevicePolicy(path, "device-a"); err == nil {
		t.Fatal("missing configured policy must fail closed")
	}
	if err = os.WriteFile(path, []byte("bad"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = loadDevicePolicy(path, "device-a"); err == nil {
		t.Fatal("corrupt policy must fail closed")
	}
	if err = saveDevicePolicy(path, "device-a", p); err == nil {
		t.Fatal("should not overwrite corrupt policy")
	}
}
