package whatsapp

import (
	"context"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
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

// syncKeysDevice ensures the keysDB contains exactly the same set of devices as the
// primary DB. The original implementation used GetFirstDevice and deleted everything
// else, which silently wiped session keys for all but the first device in multi-device
// setups. This version does a full bidirectional sync:
//   - devices in keysDB not found in primary → deleted (orphans)
//   - devices in primary not found in keysDB → added (missing keys store entries)
func syncKeysDevice(ctx context.Context, primaryDB, keysDB *sqlstore.Container) {
	if keysDB == nil {
		return
	}

	primaryDevices, err := primaryDB.GetAllDevices(ctx)
	if err != nil {
		log.Errorf("[SYNC_KEYS] failed to list devices from primary DB: %v", err)
		return
	}

	// Build a set of IDs present in primary for O(1) lookup.
	primaryIDs := make(map[string]struct{}, len(primaryDevices))
	for _, d := range primaryDevices {
		if d.ID != nil {
			primaryIDs[d.ID.String()] = struct{}{}
		}
	}

	// Remove from keysDB any device not present in primary.
	keysDevices, err := keysDB.GetAllDevices(ctx)
	if err != nil {
		log.Errorf("[SYNC_KEYS] failed to list devices from keys DB: %v", err)
		return
	}
	keysIDs := make(map[string]struct{}, len(keysDevices))
	for _, d := range keysDevices {
		if d.ID == nil {
			continue
		}
		keysIDs[d.ID.String()] = struct{}{}
		if _, ok := primaryIDs[d.ID.String()]; !ok {
			if delErr := keysDB.DeleteDevice(ctx, d); delErr != nil {
				log.Warnf("[SYNC_KEYS] failed to delete orphan device %s from keys DB: %v", d.ID, delErr)
			} else {
				log.Infof("[SYNC_KEYS] removed orphan device %s from keys DB", d.ID)
			}
		}
	}

	// Add to keysDB any device present in primary but missing from keysDB.
	for _, d := range primaryDevices {
		if d.ID == nil {
			continue
		}
		if _, ok := keysIDs[d.ID.String()]; !ok {
			if putErr := keysDB.PutDevice(ctx, d); putErr != nil {
				log.Warnf("[SYNC_KEYS] failed to add device %s to keys DB: %v", d.ID, putErr)
			} else {
				log.Infof("[SYNC_KEYS] added missing device %s to keys DB", d.ID)
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

	// Resolve fingerprint — try loading persisted one from DB, else assign a random one
	var fp DeviceFingerprint
	nonADJID := ""
	if device.ID != nil {
		nonADJID = device.ID.ToNonAD().String()
	}
	if chatStorageRepo != nil && nonADJID != "" {
		records, _ := chatStorageRepo.ListDeviceRecords()
		for _, rec := range records {
			if rec.JID == nonADJID && rec.Fingerprint != "" {
				if parsed, ok := ParseFingerprint(rec.Fingerprint); ok {
					fp = parsed
					break
				}
			}
		}
	}
	if fp.Os == "" {
		fp = RandomFingerprint()
	}

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

	// Create and configure the client under the fingerprint mutex to prevent races
	devicePropsMu.Lock()
	applyFingerprintLocked(fp)
	baseLogger := waLog.Stdout("Client", config.WhatsappLogLevel, true)
	client := whatsmeow.NewClient(device, newFilteredLogger(baseLogger))
	devicePropsMu.Unlock()
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true
	client.SetForceActiveDeliveryReceipts(true)

	deviceRepo := newDeviceChatStorage(instanceID, chatStorageRepo)
	instance := NewDeviceInstance(instanceID, client, deviceRepo)
	instance.SetFingerprint(fp.String())

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
