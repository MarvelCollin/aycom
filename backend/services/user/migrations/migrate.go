package migrations

import (
	"fmt"
	"log"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/model"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/seeds"
	"gorm.io/gorm"
)

// SchemaVersion tracks the migration version in the database
type SchemaVersion struct {
	Version   string    `gorm:"primaryKey"`
	AppliedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// Migrate runs migrations for the user service
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
	// Check if seeding is needed (no profiles exist)
	var count int64
	db.Model(&model.UserProfile{}).Count(&count)

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

	// Create tables with updated UserProfile model
	err := db.AutoMigrate(&model.UserProfile{}, &model.Contact{})
	if err != nil {
		return err
	}

	// Ensure indexes are created for frequently queried fields
	db.Exec("CREATE INDEX IF NOT EXISTS idx_user_profiles_user_id ON user_profiles(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_contacts_user_id ON contacts(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_contacts_contact_user_id ON contacts(contact_user_id)")

	return nil
}
