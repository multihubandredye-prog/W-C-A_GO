package whatsapp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/multihubandredye-prog/W-C-A_GO/config"
	domainChatStorage "github.com/multihubandredye-prog/W-C-A_GO/domains/chatstorage"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Type definitions
type ExtractedMedia struct {
	MediaPath string `json:"media_path"`
	MimeType  string `json:"mime_type"`
	Caption   string `json:"caption"`
}

// Global variables
var (
	globalStateMu sync.RWMutex
	cli           *whatsmeow.Client
	db            *sqlstore.Container // Add global database reference for cleanup
	keysDB        *sqlstore.Container
	deviceManager *DeviceManager
	log           waLog.Logger
	startupTime   = time.Now().Unix()
)

// syncKeysDevice mirrors every device from the primary store into the keys
// store. The keys store doubles as the session cache for outgoing sends (see
// applyKeyCacheStore) and its schema references whatsmeow_device(jid), so
// writing a session for a device without a row there fails with "FOREIGN KEY
// constraint failed". Mirroring all linked devices (not just the first one)
// lets any connected number send messages. The previous implementation also
// compared device IDs by pointer, which never matched across scans, so every
// sync wiped the keys store and re-inserted a single device.
func syncKeysDevice(ctx context.Context, db, keysDB *sqlstore.Container) {
	if keysDB == nil {
		return
	}

	primaryDevices, err := db.GetAllDevices(ctx)
	if err != nil {
		log.Errorf("Failed to get all devices: %v", err)
		return
	}
	primaryJIDs := make(map[types.JID]struct{}, len(primaryDevices))
	for _, dev := range primaryDevices {
		if dev.ID != nil {
			primaryJIDs[*dev.ID] = struct{}{}
		}
	}

	keysDevices, err := keysDB.GetAllDevices(ctx)
	if err != nil {
		log.Errorf("Failed to get all devices: %v", err)
		return
	}
	for _, d := range keysDevices {
		if d.ID != nil {
			if _, ok := primaryJIDs[*d.ID]; !ok {
				keysDB.DeleteDevice(ctx, d)
			}
		}
	}

	for _, dev := range primaryDevices {
		if dev.ID == nil {
			continue
		}
		if keysDev, err := keysDB.GetDevice(ctx, *dev.ID); err != nil {
			log.Errorf("Failed to get device %s in keys store: %v", dev.ID, err)
		} else if keysDev == nil {
			if err := keysDB.PutDevice(ctx, dev); err != nil {
				log.Errorf("Failed to mirror device %s into keys store: %v", dev.ID, err)
			}
		}
	}
}

// InitWaCLI initializes the WhatsApp client
func InitWaCLI(ctx context.Context, storeContainer, keysStoreContainer *sqlstore.Container, chatStorageRepo domainChatStorage.IChatStorageRepository) *whatsmeow.Client {
	device, err := storeContainer.GetFirstDevice(ctx)
	if err != nil {
		log.Errorf("Failed to get device: %v", err)
		panic(err)
	}

	if device == nil {
		log.Errorf("No device found")
		panic("No device found")
	}

	// Configure device properties
	osName := fmt.Sprintf("%s %s", config.AppOs, config.AppVersion)
	store.DeviceProps.PlatformType = &config.AppPlatform
	store.DeviceProps.Os = &osName

	// Keep references for global state update after client creation
	primaryDB := storeContainer
	keysContainer := keysStoreContainer

	// Configure a separated database for accelerating encryption caching
	if keysContainer != nil && device.ID != nil {
		innerStore := sqlstore.NewSQLStore(keysStoreContainer, *device.ID)

		syncKeysDevice(ctx, primaryDB, keysContainer)
		applyKeyCacheStore(device, innerStore)
	}

	instanceID := ""
	if device.ID != nil {
		instanceID = device.ID.String()
	}

	// Create and configure the client with filtered logging to avoid noisy reconnection EOF errors
	baseLogger := waLog.Stdout("Client", config.WhatsappLogLevel, true)
	client := whatsmeow.NewClient(device, newFilteredLogger(baseLogger))
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true

	deviceRepo := newDeviceChatStorage(instanceID, chatStorageRepo)
	instance := NewDeviceInstance(instanceID, client, deviceRepo)

	client.AddEventHandler(func(rawEvt any) {
		handler(ctx, instance, rawEvt)
	})

	// Register device instance in the manager for multi-device awareness
	// Use EnsureDefault to avoid creating duplicates when a device with matching JID already exists
	if device.ID != nil {
		instanceID = device.ID.String()
	}
	dm := InitializeDeviceManager(storeContainer, keysStoreContainer, deviceRepo)
	if dm != nil && instanceID != "" {
		dm.EnsureDefault(instance)
		instance.SetOnLoggedOut(func(deviceID string) {
			dm.RemoveDevice(deviceID)
		})
	}

	globalStateMu.Lock()
	cli = client
	db = primaryDB
	keysDB = keysContainer
	globalStateMu.Unlock()

	return client
}
