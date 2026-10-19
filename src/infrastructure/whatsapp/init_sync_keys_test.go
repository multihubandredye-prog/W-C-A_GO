package whatsapp

import (
	"context"
	"testing"

	"github.com/multihubandredye-prog/W-C-A_GO/pkg/sqlite"
	"go.mau.fi/whatsmeow/proto/waAdv"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func newSyncKeysTestContainer(t *testing.T, name string) *sqlstore.Container {
	t.Helper()
	c, err := sqlstore.New(context.Background(), sqlite.DriverName, "file:"+name+"?mode=memory&cache=shared&_foreign_keys=on", waLog.Stdout("sync-keys-test", "ERROR", true))
	if err != nil {
		t.Fatalf("failed to create container %s: %v", name, err)
	}
	return c
}

func putSyncKeysTestDevice(t *testing.T, c *sqlstore.Container, user string) types.JID {
	t.Helper()
	jid := types.JID{User: user, Server: types.DefaultUserServer}
	dev := c.NewDevice()
	dev.ID = &jid
	// PutDevice dereferences Account (filled during real pairing); tests only
	// need the whatsmeow_device row to exist for the foreign key behavior.
	dev.Account = &waAdv.ADVSignedDeviceIdentity{
		Details:             []byte{},
		AccountSignature:    make([]byte, 64), // schema CHECK: length = 64
		AccountSignatureKey: make([]byte, 32), // schema CHECK: length = 32
		DeviceSignature:     make([]byte, 64), // schema CHECK: length = 64
	}
	if err := c.PutDevice(context.Background(), dev); err != nil {
		t.Fatalf("failed to put device %s: %v", user, err)
	}
	return jid
}

func TestSyncKeysDevice_MirrorsAllDevices(t *testing.T) {
	ctx := context.Background()
	primary := newSyncKeysTestContainer(t, "synckeys_primary")
	keys := newSyncKeysTestContainer(t, "synckeys_keys")

	jid1 := putSyncKeysTestDevice(t, primary, "555000000001")
	jid2 := putSyncKeysTestDevice(t, primary, "555000000002")

	// Keys store starts with a stale device that is gone from the primary
	// store, and without the second linked device entirely.
	putSyncKeysTestDevice(t, keys, "555000000099")

	syncKeysDevice(ctx, primary, keys)

	devs, err := keys.GetAllDevices(ctx)
	if err != nil {
		t.Fatalf("failed to list keys devices: %v", err)
	}
	got := make(map[types.JID]bool, len(devs))
	for _, d := range devs {
		if d.ID != nil {
			got[*d.ID] = true
		}
	}
	if len(got) != 2 || !got[jid1] || !got[jid2] {
		t.Fatalf("expected keys store to mirror exactly the 2 primary devices, got %v", got)
	}

	// Regression for the reported send failure: writing a session for the
	// SECOND linked number must not violate the whatsmeow_device foreign key.
	sender := sqlstore.NewSQLStore(keys, jid2)
	if err := sender.PutSession(ctx, "555999999999.0:1@s.whatsapp.net", []byte("session-bytes")); err != nil {
		t.Fatalf("foreign key regression: failed to store session for second device: %v", err)
	}
}

func TestSyncKeysDevice_EmptyPrimaryClearsKeys(t *testing.T) {
	ctx := context.Background()
	primary := newSyncKeysTestContainer(t, "synckeys_empty_primary")
	keys := newSyncKeysTestContainer(t, "synckeys_empty_keys")

	putSyncKeysTestDevice(t, keys, "555000000099")

	syncKeysDevice(ctx, primary, keys)

	devs, err := keys.GetAllDevices(ctx)
	if err != nil {
		t.Fatalf("failed to list keys devices: %v", err)
	}
	if len(devs) != 0 {
		t.Fatalf("expected keys store to be empty when primary has no devices, got %d devices", len(devs))
	}
}
