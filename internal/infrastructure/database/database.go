package database

import (
	"fmt"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/smart-onestop/platform/internal/domain"
	"github.com/smart-onestop/platform/internal/infrastructure/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes database connection, connection pool, and auto-migrations
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	if cfg.DatabaseDriver == "postgres" {
		dialector = postgres.Open(cfg.DatabaseURL)
		log.Printf("[DB] Connecting to PostgreSQL...")
	} else {
		dialector = sqlite.Open(cfg.DatabaseURL)
		log.Printf("[DB] Connecting to SQLite (%s)...", cfg.DatabaseURL)
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sql.DB: %w", err)
	}

	// Enterprise Connection Pooling
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)

	// Run AutoMigrations
	err = db.AutoMigrate(
		&domain.AdministrativeStructure{},
		&domain.User{},
		&domain.Citizen{},
		&domain.ServiceType{},
		&domain.Case{},
		&domain.CaseAttachment{},
		&domain.HearingSlot{},
		&domain.CaseAuditLog{},
		&domain.EscalationPolicy{},
	)
	if err != nil {
		return nil, fmt.Errorf("database auto-migration failed: %w", err)
	}

	log.Printf("[DB] Database schema migrated successfully.")
	return db, nil
}
