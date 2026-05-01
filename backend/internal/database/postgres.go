package database

import (
	"fmt"
	"log"

	"bank-sampah-backend/internal/config"
	"bank-sampah-backend/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to PostgreSQL and runs auto-migrations
func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	logLevel := logger.Info
	if cfg.Host == "" {
		logLevel = logger.Silent
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true, // Performance: skip wrapping single queries in tx
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	log.Println("✅ Connected to PostgreSQL successfully")

	// Auto-migrate all models
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	err := db.AutoMigrate(
		&model.Admin{},
		&model.School{},
		&model.SIDocument{},
		&model.SIItem{},
		&model.AuditLog{},
		&model.CallbackQueue{},
		&model.UsedNonce{},
	)
	if err != nil {
		return err
	}

	log.Println("✅ Database migrations completed")
	return nil
}
