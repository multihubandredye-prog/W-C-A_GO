package config

import (
	"time"

	"go.mau.fi/whatsmeow/proto/waCompanionReg"
)

var (
	AppVersion             = "1.3.46"
	AppPort                = "3000"
	AppHost                = "0.0.0.0"
	AppDebug               = false
	AppOs                  = "WCA"
	AppPlatform            = waCompanionReg.DeviceProps_PlatformType(1)
	AppBasicAuthCredential []string
	AppBasePath                     = ""
	AppBodyLimit           int64    = 135000000 // Supports a 100MB payload encoded as base64
	AppTrustedProxies      []string             // Trusted proxy IP ranges (e.g., "0.0.0.0/0" for all, or specific CIDRs)

	McpPort = "8080"
	McpHost = "localhost"

	PathQrCode   = "statics/qrcode"
	PathMedia    = "statics/media"
	PathStorages = "storages"

	DBURI     = "file:storages/whatsapp.db"
	DBKeysURI = ""

	WhatsappAutoReplyMessage          string
	WhatsappAutoMarkRead              = false // Auto-mark incoming messages as read
	WhatsappAutoDownloadMedia         = true  // Auto-download media from incoming messages
	AutoDeleteMedia                   = false // Toggle automatic media deletion
	AutoDeleteMediaDuration           = 300   // Duration in seconds to keep media files before deletion
	WhatsappWebhook                   []string
	WhatsappWebhookSecret             = "secret"
	WhatsappWebhookInsecureSkipVerify = false // Skip TLS certificate verification for webhooks (insecure)
	WhatsappWebhookAckDelivered       = true
	WhatsappWebhookAckRead            = true
	WhatsappWebhookAckReadSelf        = true
	WhatsappWebhookEvents             []string         // Whitelist of events to forward to webhook (empty = all events)
	WhatsappWebhookIncludeOutgoing             = true  // Preserve existing outgoing webhook behavior
	WhatsappWebhookDelayMS                     = 1500  // Delay between webhook deliveries per URL in milliseconds (0 = parallel delivery)
	WhatsappWebhookMaxAttempts                 = 1     // Max webhook delivery attempts per URL (1 = no retry; legacy upstream behavior was 5)
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
	WhatsappPresencePulseEnabled               = true          // Periodically pulse presence available, then unavailable
	WhatsappPresencePulseInterval              = 24 * time.Hour
	WhatsappPresencePulseDuration              = 5 * time.Minute

	ChatStorageEnabled           = true
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
)
