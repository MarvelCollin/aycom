package db

import (
	"fmt"
	"log"
	"time"

	"aycom/backend/services/user/model"

	"gorm.io/gorm"
)

type SchemaVersion struct {
	Version   string `gorm:"primaryKey"`
	AppliedAt time.Time
}

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&SchemaVersion{})
	if err != nil {
		return fmt.Errorf("failed to create schema_versions table: %w", err)
	}

	var latest SchemaVersion
	db.Order("version desc").First(&latest)

	log.Printf("Current schema version: %s", latest.Version)

	migrations := []struct {
		version string
		migrate func(*gorm.DB) error
	}{
		{"00001_init_schema", migrateInitSchema},
	}

	for _, migration := range migrations {
		if latest.Version < migration.version {
			log.Printf("Applying migration: %s", migration.version)
			tx := db.Begin()
			err := migration.migrate(tx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to apply migration %s: %w", migration.version, err)
			}
			tx.Create(&SchemaVersion{Version: migration.version})
			tx.Commit()
			log.Printf("Migration applied: %s", migration.version)
		}
	}

	return nil
}

func migrateInitSchema(db *gorm.DB) error {
	err := db.AutoMigrate(&model.User{}, &model.Session{})
	if err != nil {
		return fmt.Errorf("failed to migrate User or Session models: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)").Error; err != nil {
		log.Printf("Warning: Failed to create username index: %v", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error; err != nil {
		log.Printf("Warning: Failed to create email index: %v", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)").Error; err != nil {
		log.Printf("Warning: Failed to create session user_id index: %v", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token)").Error; err != nil {
		log.Printf("Warning: Failed to create session refresh_token index: %v", err)
	}

	return nil
}
