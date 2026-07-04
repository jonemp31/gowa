package config

import (
	"time"

	"go.mau.fi/whatsmeow/proto/waCompanionReg"
)

var (
	AppVersion             = "v8.9.5"
	AppPort                = "3000"
	AppHost                = "0.0.0.0"
	AppDebug               = false
	AppOs                  = "GOWA"
	AppPlatform            = waCompanionReg.DeviceProps_PlatformType(1)
	AppBasicAuthCredential []string
	AppBasePath            = ""
	AppTrustedProxies      []string // Trusted proxy IP ranges (e.g., "0.0.0.0/0" for all, or specific CIDRs)

	McpPort = "8080"
	McpHost = "localhost"

	PathQrCode    = "statics/qrcode"
	PathSendItems = "statics/senditems"
	PathMedia     = "statics/media"
	PathStorages  = "storages"

	DBURI     = "file:storages/whatsapp.db"
	DBKeysURI = ""

	WhatsappAutoReplyMessage          string
	WhatsappAutoMarkRead              = false // Auto-mark incoming messages as read
	WhatsappAutoDownloadMedia         = true  // Auto-download media from incoming messages
	MediaRetentionDays                = 0     // Days to keep downloaded media files (0 = keep forever)
	WhatsappWebhook                   []string
	WhatsappWebhookSecret             = "secret"
	WhatsappWebhookInsecureSkipVerify = false          // Skip TLS certificate verification for webhooks (insecure)
	WhatsappWebhookEvents             []string         // Whitelist of events to forward to webhook (empty = all events)
	WhatsappApiName                            = ""    // Human-readable identifier injected into every webhook payload as "api_name"
	WhatsappAutoRejectCall                     = false // Auto-reject incoming calls
	WhatsappLogLevel                           = "ERROR"
	WhatsappSettingMaxImageSize       int64    = 20000000  // 20MB
	WhatsappSettingMaxFileSize        int64    = 50000000  // 50MB
	WhatsappSettingMaxVideoSize       int64    = 100000000 // 100MB
	WhatsappSettingMaxDownloadSize    int64    = 500000000 // 500MB
	WhatsappTypeUser                           = "@s.whatsapp.net"
	WhatsappTypeGroup                          = "@g.us"
	WhatsappTypeLid                            = "@lid"
	WhatsappAccountValidation                  = true
	WhatsappPresenceOnConnect                  = "unavailable" // Presence to send on connect: "available", "unavailable", or "none"
	WhatsappVersion                            = ""            // WhatsApp Web version override (e.g. "2.3000.1034187832"). Empty = use whatsmeow default.
	WhatsappProxies                   []string                 // Pool of proxy URLs for per-device connections (e.g. socks5://user:pass@host:port)
	WhatsappPresencePulseEnabled               = true          // Periodically pulse presence available, then unavailable
	WhatsappPresencePulseInterval              = 24 * time.Hour
	WhatsappPresencePulseDuration              = 5 * time.Minute

	ChatStorageURI               = "file:storages/chatstorage.db"
	ChatStorageEnableForeignKeys = true
	ChatStorageEnableWAL         = true

	ChatwootEnabled   = false
	ChatwootURL       = ""
	ChatwootAPIToken  = ""
	ChatwootAccountID = 0
	ChatwootInboxID   = 0
	ChatwootDeviceID  = "" // Device ID for outbound messages (required for multi-device)

	// Chatwoot History Sync settings
	ChatwootImportMessages          = false // Enable message history import to Chatwoot
	ChatwootDaysLimitImportMessages = 3     // Days of history to import (default: 3)

	// Scaling configuration — all values are tunable via environment variables.
	// Defaults preserve the existing behaviour so no breaking change occurs when
	// the variables are not set.
	//
	//   WEBHOOK_SEMAPHORE_SIZE      — max goroutines actively delivering webhook payloads  (default: 500)
	//   WEBHOOK_DISPATCH_SIZE       — max goroutines queued waiting for a delivery slot     (default: 2000)
	//   RECONNECT_MAX_WORKERS       — parallel reconnect attempts per checker cycle         (default: 10)
	//   BOOT_MAX_WORKERS            — parallel connect attempts during boot                 (default: 5)
	//   BOOT_DEVICE_DELAY_MS        — milliseconds between consecutive boot launches        (default: 1200)
	//   RECONNECT_COOLDOWN_SECONDS  — minimum gap before re-attempting a device reconnect   (default: 120)
	//   RECONNECT_TIMEOUT_SECONDS   — per-device connect() deadline                        (default: 15)
	//   APP_REQUEST_TIMEOUT_SECONDS — global HTTP handler deadline                          (default: 45)
	WebhookSemaphoreSize      = 500
	WebhookDispatchSize       = 2000
	ReconnectMaxWorkers       = 10
	BootMaxWorkers            = 5
	BootDeviceDelayMs         = 1200
	ReconnectCooldownSeconds  = 120
	ReconnectTimeoutSeconds   = 15
	AppRequestTimeoutSeconds  = 45
)
