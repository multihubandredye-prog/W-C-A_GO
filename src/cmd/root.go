package cmd

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/store/sqlstore"

	_ "github.com/lib/pq"
	"github.com/multihubandredye-prog/W-C-A_GO/config"
	domainApp "github.com/multihubandredye-prog/W-C-A_GO/domains/app"
	domainChat "github.com/multihubandredye-prog/W-C-A_GO/domains/chat"
	domainChatStorage "github.com/multihubandredye-prog/W-C-A_GO/domains/chatstorage"
	domainDevice "github.com/multihubandredye-prog/W-C-A_GO/domains/device"
	domainGroup "github.com/multihubandredye-prog/W-C-A_GO/domains/group"
	domainMessage "github.com/multihubandredye-prog/W-C-A_GO/domains/message"
	domainNewsletter "github.com/multihubandredye-prog/W-C-A_GO/domains/newsletter"
	domainSend "github.com/multihubandredye-prog/W-C-A_GO/domains/send"
	domainUser "github.com/multihubandredye-prog/W-C-A_GO/domains/user"
	"github.com/multihubandredye-prog/W-C-A_GO/infrastructure/chatstorage"
	"github.com/multihubandredye-prog/W-C-A_GO/infrastructure/whatsapp"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/sqlite"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/utils"
	"github.com/multihubandredye-prog/W-C-A_GO/usecase"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mau.fi/whatsmeow"
)

var (
	EmbedIndex embed.FS
	EmbedViews embed.FS

	// Whatsapp
	whatsappCli *whatsmeow.Client

	// Chat Storage
	chatStorageDB   *sql.DB
	chatStorageRepo domainChatStorage.IChatStorageRepository

	// Usecase
	appUsecase        domainApp.IAppUsecase
	chatUsecase       domainChat.IChatUsecase
	sendUsecase       domainSend.ISendUsecase
	userUsecase       domainUser.IUserUsecase
	messageUsecase    domainMessage.IMessageUsecase
	groupUsecase      domainGroup.IGroupUsecase
	newsletterUsecase domainNewsletter.INewsletterUsecase
	deviceUsecase     domainDevice.IDeviceUsecase
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Short: "Send free whatsapp API",
	Long: `This application is from clone https://github.com/multihubandredye-prog/W-C-A_GO, 
you can send whatsapp over http api but your whatsapp account have to be multi device version`,
}

func init() {
	// Load environment variables first
	utils.LoadConfig(".")

	time.Local = time.UTC

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Initialize flags first, before any subcommands are added
	initFlags()

	// Then initialize other components
	cobra.OnInitialize(initEnvConfig, initApp)
}

// normalizeDatabaseURI repairs HTML-escaped query separators that may be
// introduced when a Tasker/WebView configuration is copied through HTML.
func normalizeDatabaseURI(value string) string {
	return strings.ReplaceAll(value, "&amp;", "&")
}

func positiveInt64Setting(key string, current int64) int64 {
	if !viper.IsSet(key) {
		return current
	}
	value := viper.GetInt64(key)
	if value <= 0 {
		logrus.Warnf("Ignoring invalid %s=%d; keeping %d", strings.ToUpper(key), value, current)
		return current
	}
	return value
}

func normalizeRequestLimits() {
	largest := config.WhatsappSettingMaxImageSize
	if config.WhatsappSettingMaxFileSize > largest {
		largest = config.WhatsappSettingMaxFileSize
	}
	if config.WhatsappSettingMaxVideoSize > largest {
		largest = config.WhatsappSettingMaxVideoSize
	}
	// Base64 expands binary data to roughly four thirds of its size. The
	// additional MiB covers JSON fields or multipart boundaries.
	requiredBody := largest + largest/3 + 1048576
	if config.AppBodyLimit < requiredBody {
		logrus.Warnf("Raising APP_BODY_LIMIT from %d to %d so configured media limits remain usable", config.AppBodyLimit, requiredBody)
		config.AppBodyLimit = requiredBody
	}
	if config.WhatsappSettingMaxDownloadSize < largest {
		logrus.Warnf("Raising WHATSAPP_MAX_DOWNLOAD_SIZE from %d to %d so configured media limits remain usable", config.WhatsappSettingMaxDownloadSize, largest)
		config.WhatsappSettingMaxDownloadSize = largest
	}
}

func redactedSettings(settings map[string]any) map[string]any {
	redacted := make(map[string]any, len(settings))
	for key, value := range settings {
		lower := strings.ToLower(key)
		sensitive := strings.Contains(lower, "secret") || strings.Contains(lower, "token") ||
			strings.Contains(lower, "password") || strings.Contains(lower, "credential") ||
			strings.Contains(lower, "basic_auth") || lower == "db_keys_uri"
		if sensitive && fmt.Sprint(value) != "" {
			redacted[key] = "[REDACTED]"
		} else {
			redacted[key] = value
		}
	}
	return redacted
}

// initEnvConfig loads configuration from environment variables
func initEnvConfig() {
	fmt.Println(redactedSettings(viper.AllSettings()))
	// Application settings
	if envPort := viper.GetString("app_port"); envPort != "" {
		config.AppPort = envPort
	}
	if envHost := viper.GetString("app_host"); envHost != "" {
		config.AppHost = envHost
	}
	if envDebug := viper.GetBool("app_debug"); envDebug {
		config.AppDebug = envDebug
	}
	if envOs := viper.GetString("app_os"); envOs != "" {
		config.AppOs = envOs
	}
	if envBasicAuth := viper.GetString("app_basic_auth"); envBasicAuth != "" {
		credential := strings.Split(envBasicAuth, ",")
		config.AppBasicAuthCredential = credential
	}
	if envBasePath := viper.GetString("app_base_path"); envBasePath != "" {
		config.AppBasePath = envBasePath
	}
	config.AppBodyLimit = positiveInt64Setting("app_body_limit", config.AppBodyLimit)
	if envTrustedProxies := viper.GetString("app_trusted_proxies"); envTrustedProxies != "" {
		proxies := strings.Split(envTrustedProxies, ",")
		config.AppTrustedProxies = proxies
	}

	// Database settings
	if envDBURI := viper.GetString("db_uri"); envDBURI != "" {
		config.DBURI = normalizeDatabaseURI(envDBURI)
	}
	if envDBKEYSURI := viper.GetString("db_keys_uri"); envDBKEYSURI != "" {
		config.DBKeysURI = normalizeDatabaseURI(envDBKEYSURI)
	}

	if viper.IsSet("wca_experimental_primary_signal_store") {
		config.ExperimentalPrimarySignalStore = viper.GetBool("wca_experimental_primary_signal_store")
	}

	// WhatsApp settings
	if envAutoReply := viper.GetString("whatsapp_auto_reply"); envAutoReply != "" {
		config.WhatsappAutoReplyMessage = envAutoReply
	}
	if viper.IsSet("whatsapp_auto_mark_read") {
		config.WhatsappAutoMarkRead = viper.GetBool("whatsapp_auto_mark_read")
	}
	if viper.IsSet("whatsapp_auto_download_media") {
		config.WhatsappAutoDownloadMedia = viper.GetBool("whatsapp_auto_download_media")
	}
	config.WhatsappSettingMaxImageSize = positiveInt64Setting("whatsapp_max_image_size", config.WhatsappSettingMaxImageSize)
	config.WhatsappSettingMaxFileSize = positiveInt64Setting("whatsapp_max_file_size", config.WhatsappSettingMaxFileSize)
	config.WhatsappSettingMaxVideoSize = positiveInt64Setting("whatsapp_max_video_size", config.WhatsappSettingMaxVideoSize)
	config.WhatsappSettingMaxDownloadSize = positiveInt64Setting("whatsapp_max_download_size", config.WhatsappSettingMaxDownloadSize)
	if viper.IsSet("auto_delete_media") {
		config.AutoDeleteMedia = viper.GetBool("auto_delete_media")
	}
	if viper.IsSet("auto_delete_media_duration") {
		config.AutoDeleteMediaDuration = viper.GetInt("auto_delete_media_duration")
	}
	if envWebhook := viper.GetString("whatsapp_webhook"); envWebhook != "" {
		webhook := strings.Split(envWebhook, ",")
		config.WhatsappWebhook = webhook
	}
	if envWebhookSecret := viper.GetString("whatsapp_webhook_secret"); envWebhookSecret != "" {
		config.WhatsappWebhookSecret = envWebhookSecret
	}
	if viper.IsSet("whatsapp_webhook_insecure_skip_verify") {
		config.WhatsappWebhookInsecureSkipVerify = viper.GetBool("whatsapp_webhook_insecure_skip_verify")
	}
	if viper.IsSet("whatsapp_webhook_delay_ms") {
		config.WhatsappWebhookDelayMS = viper.GetInt("whatsapp_webhook_delay_ms")
	}
	if viper.IsSet("whatsapp_webhook_max_attempts") {
		config.WhatsappWebhookMaxAttempts = viper.GetInt("whatsapp_webhook_max_attempts")
		if config.WhatsappWebhookMaxAttempts < 1 {
			config.WhatsappWebhookMaxAttempts = 1
		}
	}
	if envWebhookEvents := viper.GetString("whatsapp_webhook_events"); envWebhookEvents != "" {
		events := strings.Split(envWebhookEvents, ",")
		config.WhatsappWebhookEvents = events
	}
	if viper.IsSet("whatsapp_webhook_ack_delivered") {
		config.WhatsappWebhookAckDelivered = viper.GetBool("whatsapp_webhook_ack_delivered")
	}
	if viper.IsSet("whatsapp_webhook_ack_read") {
		config.WhatsappWebhookAckRead = viper.GetBool("whatsapp_webhook_ack_read")
	}
	if viper.IsSet("whatsapp_webhook_ack_read_self") {
		config.WhatsappWebhookAckReadSelf = viper.GetBool("whatsapp_webhook_ack_read_self")
	}
	if viper.IsSet("whatsapp_webhook_include_outgoing") {
		config.WhatsappWebhookIncludeOutgoing = viper.GetBool("whatsapp_webhook_include_outgoing")
	}
	if viper.IsSet("whatsapp_account_validation") {
		config.WhatsappAccountValidation = viper.GetBool("whatsapp_account_validation")
	}
	if viper.IsSet("whatsapp_auto_reject_call") {
		config.WhatsappAutoRejectCall = viper.GetBool("whatsapp_auto_reject_call")
	}
	if viper.IsSet("whatsapp_auto_answer_call") {
		config.WhatsappAutoAnswerCall = viper.GetBool("whatsapp_auto_answer_call")
	}
	if viper.IsSet("whatsapp_auto_answer_audio") {
		config.WhatsappAutoAnswerAudio = strings.TrimSpace(viper.GetString("whatsapp_auto_answer_audio"))
	}
	if config.WhatsappAutoAnswerCall && config.WhatsappAutoRejectCall {
		logrus.Warn("Automatic call answering is disabled because automatic call rejection is enabled")
	}
	if envPresenceOnConnect := viper.GetString("whatsapp_presence_on_connect"); envPresenceOnConnect != "" {
		config.WhatsappPresenceOnConnect = envPresenceOnConnect
	}
	if viper.IsSet("whatsapp_presence_pulse_enabled") {
		config.WhatsappPresencePulseEnabled = viper.GetBool("whatsapp_presence_pulse_enabled")
	}
	if viper.IsSet("whatsapp_presence_pulse_interval") {
		if interval := viper.GetDuration("whatsapp_presence_pulse_interval"); interval > 0 {
			config.WhatsappPresencePulseInterval = interval
		}
	}
	if viper.IsSet("whatsapp_presence_pulse_duration") {
		if duration := viper.GetDuration("whatsapp_presence_pulse_duration"); duration > 0 {
			config.WhatsappPresencePulseDuration = duration
		}
	}

	if viper.IsSet("whatsapp_chat_storage") {
		config.ChatStorageEnabled = viper.GetBool("whatsapp_chat_storage")
	}

	// Chatwoot settings
	if viper.IsSet("chatwoot_enabled") {
		config.ChatwootEnabled = viper.GetBool("chatwoot_enabled")
	}
	if envChatwootURL := viper.GetString("chatwoot_url"); envChatwootURL != "" {
		config.ChatwootURL = envChatwootURL
	}
	if envChatwootAPIToken := viper.GetString("chatwoot_api_token"); envChatwootAPIToken != "" {
		config.ChatwootAPIToken = envChatwootAPIToken
	}
	if viper.IsSet("chatwoot_account_id") {
		config.ChatwootAccountID = viper.GetInt("chatwoot_account_id")
	}
	if viper.IsSet("chatwoot_inbox_id") {
		config.ChatwootInboxID = viper.GetInt("chatwoot_inbox_id")
	}
	if envChatwootDeviceID := viper.GetString("chatwoot_device_id"); envChatwootDeviceID != "" {
		config.ChatwootDeviceID = envChatwootDeviceID
	}
	// Chatwoot History Sync settings
	if viper.IsSet("chatwoot_import_messages") {
		config.ChatwootImportMessages = viper.GetBool("chatwoot_import_messages")
	}
	if viper.IsSet("chatwoot_days_limit_import_messages") {
		config.ChatwootDaysLimitImportMessages = viper.GetInt("chatwoot_days_limit_import_messages")
	}
	normalizeRequestLimits()
}

func initFlags() {
	// Application flags
	rootCmd.PersistentFlags().StringVarP(
		&config.AppPort,
		"port", "p",
		config.AppPort,
		"change port number with --port <number> | example: --port=8080",
	)

	rootCmd.PersistentFlags().StringVarP(
		&config.AppHost,
		"host", "H",
		config.AppHost,
		`host to bind the server --host <string> | example: --host="127.0.0.1"`,
	)

	rootCmd.PersistentFlags().BoolVarP(
		&config.AppDebug,
		"debug", "d",
		config.AppDebug,
		"hide or displaying log with --debug <true/false> | example: --debug=true",
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.AppOs,
		"os", "",
		config.AppOs,
		`os name --os <string> | example: --os="Chrome"`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.AppBasicAuthCredential,
		"basic-auth", "b",
		config.AppBasicAuthCredential,
		"basic auth credential | -b=yourUsername:yourPassword",
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.AppBasePath,
		"base-path", "",
		config.AppBasePath,
		`base path for subpath deployment --base-path <string> | example: --base-path="/gowa"`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.AppTrustedProxies,
		"trusted-proxies", "",
		config.AppTrustedProxies,
		`trusted proxy IP ranges for reverse proxy deployments --trusted-proxies <string> | example: --trusted-proxies="0.0.0.0/0" or --trusted-proxies="10.0.0.0/8,172.16.0.0/12"`,
	)

	// Database flags
	rootCmd.PersistentFlags().StringVarP(
		&config.DBURI,
		"db-uri", "",
		config.DBURI,
		`the database uri to store the connection data database uri (by default, we'll use sqlite3 under storages/whatsapp.db). database uri --db-uri <string> | example: --db-uri="file:storages/whatsapp.db?_foreign_keys=on or postgres://user:password@localhost:5432/whatsapp"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.DBKeysURI,
		"db-keys-uri", "",
		config.DBKeysURI,
		`the database uri to store the optional keys cache (by default, we'll use the same database uri). avoid in-memory storage in production. database uri --db-keys-uri <string> | example: --db-keys-uri="file:storages/whatsapp-keys.db?_foreign_keys=on"`,
	)

	// WhatsApp flags
	rootCmd.PersistentFlags().StringVarP(
		&config.WhatsappAutoReplyMessage,
		"autoreply", "",
		config.WhatsappAutoReplyMessage,
		`auto reply when received message --autoreply <string> | example: --autoreply="Don't reply this message"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAutoMarkRead,
		"auto-mark-read", "",
		config.WhatsappAutoMarkRead,
		`auto mark incoming messages as read --auto-mark-read <true/false> | example: --auto-mark-read=true`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAutoDownloadMedia,
		"auto-download-media", "",
		config.WhatsappAutoDownloadMedia,
		`auto download media from incoming messages --auto-download-media <true/false> | example: --auto-download-media=false`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.AutoDeleteMedia,
		"auto-delete-media", "",
		config.AutoDeleteMedia,
		`auto delete media --auto-delete-media <true/false> | example: --auto-delete-media=true`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.WhatsappWebhook,
		"webhook", "w",
		config.WhatsappWebhook,
		`forward event to webhook --webhook <string> | example: --webhook="https://yourcallback.com/callback"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.WhatsappWebhookSecret,
		"webhook-secret", "",
		config.WhatsappWebhookSecret,
		`secure webhook request --webhook-secret <string> | example: --webhook-secret="super-secret-key"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappWebhookInsecureSkipVerify,
		"webhook-insecure-skip-verify", "",
		config.WhatsappWebhookInsecureSkipVerify,
		`skip TLS certificate verification for webhooks (INSECURE - use only for development/self-signed certs) --webhook-insecure-skip-verify <true/false> | example: --webhook-insecure-skip-verify=true`,
	)
	rootCmd.PersistentFlags().IntVarP(
		&config.WhatsappWebhookDelayMS,
		"webhook-delay-ms", "",
		config.WhatsappWebhookDelayMS,
		`delay between webhook deliveries per URL in milliseconds (0 = parallel delivery) --webhook-delay-ms <int> | example: --webhook-delay-ms=1500`,
	)
	rootCmd.PersistentFlags().IntVarP(
		&config.WhatsappWebhookMaxAttempts,
		"webhook-max-attempts", "",
		config.WhatsappWebhookMaxAttempts,
		`max webhook delivery attempts per URL (1 = no retry) --webhook-max-attempts <int> | example: --webhook-max-attempts=1`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.WhatsappWebhookEvents,
		"webhook-events", "",
		config.WhatsappWebhookEvents,
		`whitelist of events to forward to webhook (empty = all events) --webhook-events <string> | example: --webhook-events="message,message.ack,group.participants"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAccountValidation,
		"account-validation", "",
		config.WhatsappAccountValidation,
		`enable or disable account validation --account-validation <true/false> | example: --account-validation=true`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAutoRejectCall,
		"auto-reject-call", "",
		config.WhatsappAutoRejectCall,
		`auto reject incoming calls --auto-reject-call <true/false> | example: --auto-reject-call=true`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.WhatsappPresenceOnConnect,
		"presence-on-connect", "",
		config.WhatsappPresenceOnConnect,
		`presence to send on connect: "available", "unavailable", or "none" --presence-on-connect <string> | example: --presence-on-connect="unavailable"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappPresencePulseEnabled,
		"presence-pulse-enabled", "",
		config.WhatsappPresencePulseEnabled,
		`enable daily presence pulse --presence-pulse-enabled <true/false> | example: --presence-pulse-enabled=true`,
	)
	rootCmd.PersistentFlags().DurationVarP(
		&config.WhatsappPresencePulseInterval,
		"presence-pulse-interval", "",
		config.WhatsappPresencePulseInterval,
		`presence pulse interval --presence-pulse-interval <duration> | example: --presence-pulse-interval=24h`,
	)
	rootCmd.PersistentFlags().DurationVarP(
		&config.WhatsappPresencePulseDuration,
		"presence-pulse-duration", "",
		config.WhatsappPresencePulseDuration,
		`duration to stay available during a presence pulse --presence-pulse-duration <duration> | example: --presence-pulse-duration=5m`,
	)

	// Chatwoot flags
	rootCmd.PersistentFlags().BoolVarP(
		&config.ChatwootEnabled,
		"chatwoot-enabled", "",
		config.ChatwootEnabled,
		`enable Chatwoot integration --chatwoot-enabled <true/false> | example: --chatwoot-enabled=true`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.ChatwootDeviceID,
		"chatwoot-device-id", "",
		config.ChatwootDeviceID,
		`device ID for Chatwoot outbound messages --chatwoot-device-id <string> | example: --chatwoot-device-id="my-device"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.ChatwootImportMessages,
		"chatwoot-import-messages", "",
		config.ChatwootImportMessages,
		`enable message history import to Chatwoot --chatwoot-import-messages <true/false> | example: --chatwoot-import-messages=true`,
	)
	rootCmd.PersistentFlags().IntVarP(
		&config.ChatwootDaysLimitImportMessages,
		"chatwoot-days-limit-import-messages", "",
		config.ChatwootDaysLimitImportMessages,
		`days of message history to import to Chatwoot --chatwoot-days-limit-import-messages <int> | example: --chatwoot-days-limit-import-messages=7`,
	)
}

func initChatStorage() (*sql.DB, error) {
	chatStorageURI := normalizeDatabaseURI(config.ChatStorageURI)
	enableWAL := config.ChatStorageEnableWAL
	if !config.ChatStorageEnabled {
		chatStorageURI = "file:wca-chatstorage-disabled?mode=memory&cache=shared"
		enableWAL = false
		logrus.Warn("WHATSAPP_CHAT_STORAGE=false: chats, messages and device registry are temporary and will be lost on restart")
	}
	connStr := sqlite.FormatChatStorageURI(chatStorageURI, enableWAL, config.ChatStorageEnableForeignKeys)

	db, err := sql.Open(sqlite.DriverName, connStr)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func initApp() {
	if config.AppDebug {
		config.WhatsappLogLevel = "DEBUG"
		logrus.SetLevel(logrus.DebugLevel)
	}

	//preparing folder if not exist
	err := utils.CreateFolder(config.PathQrCode, config.PathMedia, config.PathStorages, config.PathMedia)
	if err != nil {
		logrus.Errorln(err)
	}

	ctx := context.Background()

	chatStorageDB, err = initChatStorage()
	if err != nil {
		// Terminate the application if chat storage fails to initialize to avoid nil pointer panics later.
		logrus.Fatalf("failed to initialize chat storage: %v", err)
	}

	chatStorageRepo = chatstorage.NewStorageRepository(chatStorageDB)
	chatStorageRepo.InitializeSchema()

	whatsappDB := whatsapp.InitWaDB(ctx, config.DBURI)
	var keysDB *sqlstore.Container
	if config.ExperimentalPrimarySignalStore {
		logrus.Warn("Primary Signal store enabled; existing secondary Signal state is not migrated")
	} else if config.DBKeysURI != "" {
		keysDB = whatsapp.InitWaDB(ctx, config.DBKeysURI)
	}

	whatsappCli = whatsapp.InitWaCLI(ctx, whatsappDB, keysDB, chatStorageRepo)

	// Initialize device manager and usecase for multi-device support
	dm := whatsapp.GetDeviceManager()
	if dm != nil {
		if err := dm.LoadExistingDevices(ctx); err != nil {
			logrus.WithError(err).Error("failed to load WhatsApp devices; automatic recovery may be incomplete")
		}
	}

	// Usecase
	appUsecase = usecase.NewAppService(chatStorageRepo, dm)
	chatUsecase = usecase.NewChatService(chatStorageRepo)
	sendUsecase = usecase.NewSendService(appUsecase, chatStorageRepo)
	userUsecase = usecase.NewUserService(chatStorageRepo)
	messageUsecase = usecase.NewMessageService(chatStorageRepo)
	groupUsecase = usecase.NewGroupService()
	newsletterUsecase = usecase.NewNewsletterService()
	deviceUsecase = usecase.NewDeviceService(dm)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(embedIndex embed.FS, embedViews embed.FS) {
	EmbedIndex = embedIndex
	EmbedViews = embedViews
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
