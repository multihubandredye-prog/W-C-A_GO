package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/sqlite"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// InitWaDB initializes the WhatsApp database connection
func InitWaDB(ctx context.Context, DBURI string) *sqlstore.Container {
	log = waLog.Stdout("Main", config.WhatsappLogLevel, true)
	dbLog := waLog.Stdout("Database", config.WhatsappLogLevel, true)

	storeContainer, err := initDatabase(ctx, dbLog, DBURI)
	if err != nil {
		log.Errorf("Database initialization error: %v", err)
		panic(pkgError.InternalServerError(fmt.Sprintf("Database initialization error: %v", err)))
	}

	return storeContainer
}

// initDatabase creates and returns a database store container based on the configured URI
func initDatabase(ctx context.Context, dbLog waLog.Logger, DBURI string) (*sqlstore.Container, error) {
	// Strip surrounding quotes that may come from .env file parsing
	DBURI = strings.Trim(DBURI, `"'`)

	if strings.HasPrefix(DBURI, "file:") {
		DBURI = sqlite.FormatChatStorageURI(DBURI, true, true)
		return sqlstore.New(ctx, sqlite.DriverName, DBURI, dbLog)
	} else if strings.HasPrefix(DBURI, "postgres:") {
		return sqlstore.New(ctx, "postgres", DBURI, dbLog)
	}

	return nil, fmt.Errorf("unknown database type: %s. Currently only sqlite3(file:) and postgres are supported", DBURI)
}

// DatabaseIntegrityResult describes the outcome of a SQLite integrity check.
type DatabaseIntegrityResult struct {
	OK      bool   `json:"ok"`
	Detail  string `json:"detail"`
	URI     string `json:"uri"`
}

// CheckDatabaseIntegrity opens a fresh connection to the configured WhatsApp
// store database and runs PRAGMA quick_check. It reports whether the SQLite
// file is healthy so callers can detect a corrupted database (for example the
// "database disk image is malformed" error) before it surfaces on endpoints.
func CheckDatabaseIntegrity(ctx context.Context) DatabaseIntegrityResult {
	result := DatabaseIntegrityResult{URI: config.DBURI}

	db, err := sql.Open(sqlite.DriverName, sqlite.FormatChatStorageURI(strings.Trim(config.DBURI, `"'`), true, true))
	if err != nil {
		result.Detail = fmt.Sprintf("failed to open database: %v", err)
		return result
	}
	defer db.Close()

	var status string
	if err := db.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&status); err != nil {
		result.Detail = fmt.Sprintf("failed to run quick_check: %v", err)
		return result
	}

	result.OK = status == "ok"
	result.Detail = status
	return result
}
