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
		{"00002_add_admin_fields", migrateAddAdminFields},
		{"00003_add_admin_tables", migrateAddAdminTables},
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
	// Migrate all required models including Follow
	err := db.AutoMigrate(&model.User{}, &model.Session{}, &model.Follow{})
	if err != nil {
		return fmt.Errorf("failed to migrate models: %w", err)
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
	// Add index for follower-followed pairs
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_follower_followed ON follows(follower_id, followed_id)").Error; err != nil {
		log.Printf("Warning: Failed to create follow index: %v", err)
	}

	return nil
}

func migrateAddAdminFields(db *gorm.DB) error {
	// Add IsAdmin column if it doesn't exist
	if !db.Migrator().HasColumn(&model.User{}, "is_admin") {
		if err := db.Exec("ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE").Error; err != nil {
			return fmt.Errorf("failed to add is_admin column: %w", err)
		}
		log.Println("Added is_admin column to users table")
	}

	// Add IsBanned column if it doesn't exist
	if !db.Migrator().HasColumn(&model.User{}, "is_banned") {
		if err := db.Exec("ALTER TABLE users ADD COLUMN is_banned BOOLEAN DEFAULT FALSE").Error; err != nil {
			return fmt.Errorf("failed to add is_banned column: %w", err)
		}
		log.Println("Added is_banned column to users table")
	}

	// Set admin flag for admin user
	if err := db.Exec("UPDATE users SET is_admin = TRUE WHERE username = 'admin'").Error; err != nil {
		log.Printf("Warning: Failed to set admin flag for admin user: %v", err)
	}

	return nil
}

func migrateAddAdminTables(db *gorm.DB) error {
	// Migrate admin-related models
	err := db.AutoMigrate(
		&model.CommunityRequest{},
		&model.PremiumRequest{},
		&model.ReportRequest{},
		&model.ThreadCategory{},
		&model.CommunityCategory{},
		&model.Newsletter{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate admin models: %w", err)
	}

	// Create indices for admin tables
	tables := []struct {
		name   string
		column string
	}{
		{"community_requests", "user_id"},
		{"community_requests", "status"},
		{"premium_requests", "user_id"},
		{"premium_requests", "status"},
		{"report_requests", "reporter_id"},
		{"report_requests", "reported_user_id"},
		{"report_requests", "status"},
		{"newsletters", "sent_by"},
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_%s ON %s(%s)",
			table.name, table.column, table.name, table.column)).Error; err != nil {
			log.Printf("Warning: Failed to create index %s on %s: %v", table.column, table.name, err)
		}
	}

	return nil
}
