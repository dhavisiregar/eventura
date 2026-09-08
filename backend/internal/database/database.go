package database

import (
	"fmt"
	"log"
	"time"

	"eventman/backend/internal/config"
	"eventman/backend/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a MySQL connection with retries (useful when the DB
// container is still starting up under docker-compose) and runs
// AutoMigrate for every model.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	if cfg.DBSSL {
		// Required by managed MySQL-compatible hosts like PlanetScale.
		dsn += "&tls=true"
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
