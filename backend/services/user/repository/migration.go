package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"

	// Import migrations package to register Goose migrations via init()
	_ "github.com/Acad600-Tpa/WEB-MV-242/services/user/migrations"
)

// MigrationConfig holds configuration for migrations
type MigrationConfig struct {
	MigrationsDir string
	DBDialect     string
}

// RunMigrations runs database migrations using Goose
func RunMigrations(gormDB *gorm.DB, config MigrationConfig) error {
	// Get the underlying *sql.DB from the gorm.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get *sql.DB from gorm.DB: %w", err)
	}

	return runGooseMigrations(sqlDB, config)
}

// runGooseMigrations runs the actual Goose migrations
func runGooseMigrations(db *sql.DB, config MigrationConfig) error {
	// Set dialect
	if err := goose.SetDialect(config.DBDialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Run migrations
	if err := goose.Up(db, config.MigrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

// GetMigrationStatus returns the status of all migrations
func GetMigrationStatus(gormDB *gorm.DB, config MigrationConfig) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get *sql.DB from gorm.DB: %w", err)
	}

	if err := goose.SetDialect(config.DBDialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	return goose.Status(sqlDB, config.MigrationsDir)
}
