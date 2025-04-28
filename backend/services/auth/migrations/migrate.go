package migrations

import (
	"fmt"
	"log"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/model"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/seeds"
	"gorm.io/gorm"
)

// SchemaVersion tracks the migration version in the database
type SchemaVersion struct {
	Version   string    `gorm:"primaryKey"`
	AppliedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// Migrate runs migrations for the auth service
func Migrate(db *gorm.DB) error {
	// Create schema_versions table if it doesn't exist
	err := db.AutoMigrate(&SchemaVersion{})
	if err != nil {
		return fmt.Errorf("failed to create schema_versions table: %w", err)
	}

	// Get the current schema version
	var latest SchemaVersion
	db.Order("version desc").First(&latest)

	log.Printf("Current schema version: %s", latest.Version)

	// Execute migrations in order
	migrations := []struct {
		version string
		migrate func(*gorm.DB) error
	}{
		{"00001_init_schema", migrateInitSchema},
		// Add more migrations here in the future
	}

	// Apply migrations that haven't been applied yet
	for _, migration := range migrations {
		if latest.Version < migration.version {
			log.Printf("Applying migration: %s", migration.version)

			// Start a transaction
			tx := db.Begin()

			err := migration.migrate(tx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to apply migration %s: %w", migration.version, err)
			}

			// Record migration
			tx.Create(&SchemaVersion{Version: migration.version})

			// Commit transaction
			tx.Commit()

			log.Printf("Migration applied: %s", migration.version)
		}
	}

	return nil
}

// Seeds the database
func Seed(db *gorm.DB) error {
	// Check if seeding is needed (no users exist)
	var count int64
	db.Model(&model.User{}).Count(&count)

	if count == 0 {
		log.Println("Database is empty, seeding...")
		return seeds.SeedAll(db)
	}

	log.Println("Database already contains data, skipping seeding")
	return nil
}

// MigrateAndSeed performs both migration and seeding if needed
func MigrateAndSeed(db *gorm.DB) error {
	err := Migrate(db)
	if err != nil {
		return err
	}

	return Seed(db)
}

// migrateInitSchema creates the initial schema
func migrateInitSchema(db *gorm.DB) error {
	// Enable UUID extension
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	// Create tables with updated User model
	err := db.AutoMigrate(&model.User{}, &model.Session{})
	if err != nil {
		return err
	}

	// Ensure indexes are created for frequently queried fields
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id)")

	// Index for sessions table
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token)")

	return nil
}
