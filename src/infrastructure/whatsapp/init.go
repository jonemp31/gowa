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

func syncKeysDevice(ctx context.Context, db, keysDB *sqlstore.Container) {
	if keysDB == nil {
		return
	}

	dev, err := db.GetFirstDevice(ctx)
	if err != nil {
		log.Errorf("Failed to get all devices: %v", err)
	} else {
		found := false
		if devs, err := keysDB.GetAllDevices(ctx); err != nil {
			log.Errorf("Failed to get all devices: %v", err)
		} else {
			for _, d := range devs {
				if d.ID == dev.ID {
					found = true
					break
				} else {
					keysDB.DeleteDevice(ctx, d)
				}
			}

			if !found {
				keysDB.PutDevice(ctx, dev)
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
		device.Identities = innerStore
		device.Sessions = innerStore
		device.PreKeys = innerStore
		device.SenderKeys = innerStore
		device.MsgSecrets = innerStore
		device.PrivacyTokens = innerStore
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

	client.AddEventHandler(func(rawEvt interface{}) {
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
