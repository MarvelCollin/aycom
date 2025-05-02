package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type SchemaVersion struct {
	Version   string    `gorm:"primaryKey"`
	AppliedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
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
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	// Add your user model and session model here
	type User struct {
		ID                     string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
		Email                  string `gorm:"size:255;uniqueIndex;not null"`
		Name                   string `gorm:"size:255;not null"`
		Username               string `gorm:"size:255;uniqueIndex;not null"`
		PasswordHash           string `gorm:"size:255;not null"`
		PasswordSalt           string `gorm:"size:255"`
		VerificationCode       *string
		IsActivated            bool
		Gender                 string
		DateOfBirth            time.Time
		SecurityQuestion       string
		SecurityAnswer         string
		NewsletterSubscription bool
		JoinedAt               time.Time
		CreatedAt              time.Time
		UpdatedAt              time.Time
	}

	type Session struct {
		ID           string    `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
		UserID       string    `gorm:"type:uuid;not null;index"`
		AccessToken  string    `gorm:"type:text;not null"`
		RefreshToken string    `gorm:"type:text;not null"`
		IPAddress    string    `gorm:"size:255"`
		UserAgent    string    `gorm:"size:255"`
		ExpiresAt    time.Time `gorm:"not null"`
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}

	err := db.AutoMigrate(&User{}, &Session{})
	if err != nil {
		return err
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token)")

	return nil
}
