package testutil

import (
	"fmt"
	"testing"
	"time"

	"eventman/backend/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewTestDB spins up an isolated in-memory SQLite DB (pure-Go driver, no CGO
// required) with the full schema migrated, for fast service/handler-level
// tests without a real MySQL server. Each call gets its own uniquely-named
// in-memory database so parallel/sequential tests never see each other's data.
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", t.Name(), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(models.AllModels()...); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}
