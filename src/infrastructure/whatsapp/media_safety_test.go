package whatsapp

import (
	"context"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"testing"
)

func TestCleanupDoesNotSelectSQLiteFiles(t *testing.T) {
	for _, name := range []string{"whatsapp.db", "whatsapp.db-wal", "whatsapp.db-shm", "whatsapp.db-journal", "database.sqlite", "database.sqlite-wal", "device.json", "data.txt"} {
		if isDisposableMedia(name) {
			t.Errorf("unsafe cleanup target: %s", name)
		}
	}
	for _, name := range []string{"image.jpg", "image.PNG", "video.mp4", "audio.ogg"} {
		if !isDisposableMedia(name) {
			t.Errorf("media not selected: %s", name)
		}
	}
}

func TestMatchingStoredDevicesDoesNotSelectOtherAccounts(t *testing.T) {
	a := types.JID{User: "123", Server: types.DefaultUserServer, Device: 1}
	b := types.JID{User: "123", Server: types.DefaultUserServer, Device: 2}
	c := types.JID{User: "456", Server: types.DefaultUserServer, Device: 1}
	devices := []*store.Device{{ID: &a}, {ID: &b}, {ID: &c}}
	exact := matchingStoredDevices(devices, a)
	if len(exact) != 1 || exact[0].ID == nil || *exact[0].ID != a {
		t.Fatal("full JID must match only its session")
	}
	normalized := a.ToNonAD()
	if len(matchingStoredDevices(devices, normalized)) != 2 {
		t.Fatal("ambiguous non-AD JID must remain ambiguous")
	}
	if len(matchingStoredDevices(devices[2:], normalized)) != 0 {
		t.Fatal("must not match another account")
	}
}

func TestPurgeDeviceUsesJIDAndPreservesOtherSessions(t *testing.T) {
	ctx := context.Background()
	primary := newSyncKeysTestContainer(t, "purge_primary")
	keys := newSyncKeysTestContainer(t, "purge_keys")
	first := putSyncKeysTestDevice(t, primary, "551111")
	second := putSyncKeysTestDevice(t, primary, "552222")
	putSyncKeysTestDevice(t, keys, "551111")
	putSyncKeysTestDevice(t, keys, "552222")
	manager := NewDeviceManager(primary, keys, nil)
	instance := NewDeviceInstance("uuid-primeiro", nil, nil)
	instance.jid = first.String()
	manager.AddDevice(instance)
	if err := manager.PurgeDevice(ctx, instance.ID()); err != nil {
		t.Fatal(err)
	}
	if _, ok := manager.GetDevice(instance.ID()); ok {
		t.Fatal("instance still in registry")
	}
	for _, container := range []*sqlstore.Container{primary, keys} {
		devices, err := container.GetAllDevices(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(devices) != 1 || devices[0].ID == nil || *devices[0].ID != second {
			t.Fatalf("unexpected sessions after purge: %+v", devices)
		}
	}
	reloaded := NewDeviceManager(primary, keys, nil)
	if err := reloaded.LoadExistingDevices(ctx); err != nil {
		t.Fatal(err)
	}
	if len(reloaded.ListDevices()) != 1 || reloaded.ListDevices()[0].JID() != second.ToNonAD().String() {
		t.Fatal("removed session reappeared or other session was lost")
	}
}

func TestCleanupProtectsConfiguredDatabaseWithMediaExtension(t *testing.T) {
	for _, suffix := range []string{"", "-wal", "-shm", "-journal"} {
		if !isConfiguredDatabase("storages/custom.jpg"+suffix, "file:storages/custom.jpg?_foreign_keys=on") {
			t.Fatal("database was not protected", suffix)
		}
	}
	if isConfiguredDatabase("storages/other.jpg", "file:storages/custom.jpg") {
		t.Fatal("unrelated image incorrectly protected")
	}
}
