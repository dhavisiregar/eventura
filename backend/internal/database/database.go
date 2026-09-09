package database

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"time"

	"eventman/backend/internal/config"
	"eventman/backend/internal/models"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// tlsConfigName is registered with the mysql driver when DBSSLCA is set, so the DSN
// can reference it via tls=<name> instead of the driver's built-in "true" mode (which
// verifies against the system trust store — useless for a host signing with its own
// private CA, e.g. Aiven).
const tlsConfigName = "eventura-pinned-ca"

// registerPinnedCA loads a PEM-encoded CA certificate (content, not a file path) and
// registers it as a named TLS config so the DSN can opt into it.
func registerPinnedCA(pemContent string) error {
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM([]byte(pemContent)); !ok {
		return fmt.Errorf("no valid certificates found in DB_SSL_CA")
	}
	return mysqldriver.RegisterTLSConfig(tlsConfigName, &tls.Config{RootCAs: pool})
}

// Connect opens a MySQL connection with retries (useful when the DB
// container is still starting up under docker-compose) and runs
// AutoMigrate for every model.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	if cfg.DBSSL {
		if cfg.DBSSLCA != "" {
			// Host signs with a private CA (e.g. Aiven) — pin it instead of relying on
			// the system trust store, which would otherwise reject the chain.
			if err := registerPinnedCA(cfg.DBSSLCA); err != nil {
				return nil, fmt.Errorf("registering DB_SSL_CA: %w", err)
			}
			dsn += "&tls=" + tlsConfigName
		} else {
			// Host has a publicly-trusted cert (or accepts opportunistic TLS) — the
			// driver's built-in mode verifies against the system trust store.
			dsn += "&tls=true"
		}
	}

	var db *gorm.DB
	var err error

	logLevel := logger.Warn
	if cfg.AppEnv == "development" {
		logLevel = logger.Warn
	}

	for attempt := 1; attempt <= 15; attempt++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
			// Vitess-based hosts (e.g. PlanetScale) don't support DB-level FK
			// constraints; referential integrity here is enforced in application
			// code (see internal/services), so this is safe everywhere.
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err == nil {
			sqlDB, sqlErr := db.DB()
			if sqlErr == nil && sqlDB.Ping() == nil {
				break
			}
			err = sqlErr
		}
		log.Printf("database: waiting for MySQL (attempt %d/15): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(models.AllModels()...); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}

	return db, nil
}
