package whatsapp

import (
	"context"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/sqlite"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"path/filepath"
	"testing"
)

// Simulates device creation and a later switch of stores; no WhatsApp account is used.
func TestAuditPairingStoreSwitchDoesNotMigrateKeys(t *testing.T) {
	ctx := context.Background()
	primary := newSyncKeysTestContainer(t, "audit_primary")
	defer primary.Close()
	secondaryPath := filepath.Join(t.TempDir(), "keys.db")
	uri := "file:" + secondaryPath + "?_foreign_keys=on"
	openSecondary := func() *sqlstore.Container {
		c, err := sqlstore.New(ctx, sqlite.DriverName, uri, waLog.Noop)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	secondary := openSecondary()
	jid := putSyncKeysTestDevice(t, primary, "555000002222")
	device, err := primary.GetDevice(ctx, jid)
	if err != nil {
		t.Fatal(err)
	}
	key, err := device.PreKeys.GenOnePreKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err = device.Sessions.PutSession(ctx, "audit-peer", []byte("test-session")); err != nil {
		t.Fatal(err)
	}
	syncKeysDevice(ctx, primary, secondary)
	applyKeyCacheStore(device, sqlstore.NewSQLStore(secondary, jid))
	missing, err := device.PreKeys.GetPreKey(ctx, key.KeyID)
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Fatal("unexpectedly migrated primary prekey")
	}
	session, err := device.Sessions.GetSession(ctx, "audit-peer")
	if err != nil {
		t.Fatal(err)
	}
	if len(session) != 0 {
		t.Fatal("unexpectedly migrated primary session")
	}
	// A new key written to persistent secondary should survive reopening it.
	saved, err := device.PreKeys.GenOnePreKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := secondary.Close(); err != nil {
		t.Fatal(err)
	}
	secondary = openSecondary()
	defer secondary.Close()
	restored, err := sqlstore.NewSQLStore(secondary, jid).GetPreKey(ctx, saved.KeyID)
	if err != nil || restored == nil {
		t.Fatalf("persistent prekey not restored: %v", err)
	}
}

// Test the architecture of the experimental mode with a new isolated database:
// both the freshly saved device and the reloaded device use the primary store.
func TestAuditPrimaryStoreSurvivesRestart(t *testing.T) {
	ctx := context.Background()
	uri := "file:" + filepath.Join(t.TempDir(), "primary.db") + "?_foreign_keys=on"
	open := func() *sqlstore.Container {
		c, err := sqlstore.New(ctx, sqlite.DriverName, uri, waLog.Noop)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	primary := open()
	jid := putSyncKeysTestDevice(t, primary, "555000003333")
	dev, err := primary.GetDevice(ctx, jid)
	if err != nil {
		t.Fatal(err)
	}
	key, err := dev.PreKeys.GenOnePreKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err = dev.Sessions.PutSession(ctx, "test-peer", []byte("session")); err != nil {
		t.Fatal(err)
	}
	if err = dev.SenderKeys.PutSenderKey(ctx, "test-group", "test-peer", []byte("group-key")); err != nil {
		t.Fatal(err)
	}
	if err = primary.Close(); err != nil {
		t.Fatal(err)
	}
	primary = open()
	defer primary.Close()
	dev, err = primary.GetDevice(ctx, jid)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := dev.PreKeys.GetPreKey(ctx, key.KeyID)
	if err != nil || restored == nil {
		t.Fatalf("prekey missing: %v", err)
	}
	session, err := dev.Sessions.GetSession(ctx, "test-peer")
	if err != nil || string(session) != "session" {
		t.Fatalf("session missing: %v", err)
	}
	group, err := dev.SenderKeys.GetSenderKey(ctx, "test-group", "test-peer")
	if err != nil || string(group) != "group-key" {
		t.Fatalf("group key missing: %v", err)
	}
}
