package cmd

import (
	"strings"
	"testing"

	"github.com/multihubandredye-prog/W-C-A_GO/config"
	"github.com/spf13/viper"
)

func TestNormalizeDatabaseURI(t *testing.T) {
	got := normalizeDatabaseURI("file::memory:?cache=shared&amp;_foreign_keys=on")
	want := "file::memory:?cache=shared&_foreign_keys=on"
	if got != want {
		t.Fatalf("normalizeDatabaseURI() = %q, want %q", got, want)
	}
}

func TestPositiveInt64Setting(t *testing.T) {
	const key = "test_positive_limit"
	viper.Set(key, int64(1234))
	t.Cleanup(func() { viper.Set(key, nil) })
	if got := positiveInt64Setting(key, 99); got != 1234 {
		t.Fatalf("positiveInt64Setting() = %d, want 1234", got)
	}
	viper.Set(key, int64(0))
	if got := positiveInt64Setting(key, 99); got != 99 {
		t.Fatalf("invalid limit = %d, want fallback 99", got)
	}
}

func TestChatStorageDisabledUsesMemory(t *testing.T) {
	previous := config.ChatStorageEnabled
	config.ChatStorageEnabled = false
	t.Cleanup(func() { config.ChatStorageEnabled = previous })
	db, err := initChatStorage()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("PRAGMA database_list")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var seq int
	var name, file string
	if !rows.Next() || rows.Scan(&seq, &name, &file) != nil {
		t.Fatal("could not inspect chat storage")
	}
	if strings.Contains(file, "chatstorage.db") {
		t.Fatalf("disabled chat storage opened persistent file %q", file)
	}
}

func TestNormalizeRequestLimitsSupportsBase64(t *testing.T) {
	oldBody := config.AppBodyLimit
	oldImage := config.WhatsappSettingMaxImageSize
	oldFile := config.WhatsappSettingMaxFileSize
	oldVideo := config.WhatsappSettingMaxVideoSize
	oldDownload := config.WhatsappSettingMaxDownloadSize
	t.Cleanup(func() {
		config.AppBodyLimit = oldBody
		config.WhatsappSettingMaxImageSize = oldImage
		config.WhatsappSettingMaxFileSize = oldFile
		config.WhatsappSettingMaxVideoSize = oldVideo
		config.WhatsappSettingMaxDownloadSize = oldDownload
	})
	config.AppBodyLimit = 10
	config.WhatsappSettingMaxImageSize = 20
	config.WhatsappSettingMaxFileSize = 50
	config.WhatsappSettingMaxVideoSize = 100
	config.WhatsappSettingMaxDownloadSize = 60
	normalizeRequestLimits()
	if config.AppBodyLimit < 134 || config.WhatsappSettingMaxDownloadSize != 100 {
		t.Fatalf("limits not normalized: body=%d download=%d", config.AppBodyLimit, config.WhatsappSettingMaxDownloadSize)
	}
}

func TestRedactedSettingsHidesCredentials(t *testing.T) {
	settings := redactedSettings(map[string]any{
		"app_basic_auth":          "user:pass",
		"whatsapp_webhook_secret": "secret",
		"app_port":                "3000",
	})
	if settings["app_basic_auth"] != "[REDACTED]" || settings["whatsapp_webhook_secret"] != "[REDACTED]" {
		t.Fatalf("credentials were not redacted: %#v", settings)
	}
	if settings["app_port"] != "3000" {
		t.Fatalf("non-sensitive setting changed: %#v", settings)
	}
}
