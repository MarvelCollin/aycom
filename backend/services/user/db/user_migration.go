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
		{"00004_add_follower_counts", migrateAddFollowerCounts},
		{"00005_add_account_privacy", migrateAddAccountPrivacy},
		{"00006_fix_premium_requests", migrateFixPremiumRequests},
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

	log.Printf("Running premium requests fix...")
	if err := UpdatePremiumRequestsTable(db); err != nil {
		log.Printf("Error fixing premium requests: %v", err)
	}

	return nil
}

func migrateInitSchema(db *gorm.DB) error {

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

	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_follower_followed ON follows(follower_id, followed_id)").Error; err != nil {
		log.Printf("Warning: Failed to create follow index: %v", err)
	}

	return nil
}

func migrateAddAdminFields(db *gorm.DB) error {

	if !db.Migrator().HasColumn(&model.User{}, "is_admin") {
		if err := db.Exec("ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE").Error; err != nil {
			return fmt.Errorf("failed to add is_admin column: %w", err)
		}
		log.Println("Added is_admin column to users table")
	}

	if !db.Migrator().HasColumn(&model.User{}, "is_banned") {
		if err := db.Exec("ALTER TABLE users ADD COLUMN is_banned BOOLEAN DEFAULT FALSE").Error; err != nil {
			return fmt.Errorf("failed to add is_banned column: %w", err)
		}
		log.Println("Added is_banned column to users table")
	}

	if err := db.Exec("UPDATE users SET is_admin = TRUE WHERE username = 'admin'").Error; err != nil {
		log.Printf("Warning: Failed to set admin flag for admin user: %v", err)
	}

	return nil
}

func migrateAddAdminTables(db *gorm.DB) error {

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

func migrateAddFollowerCounts(db *gorm.DB) error {

	if !db.Migrator().HasColumn(&model.User{}, "follower_count") {
		if err := db.Exec("ALTER TABLE users ADD COLUMN follower_count INT DEFAULT 0").Error; err != nil {
			return fmt.Errorf("failed to add follower_count column: %w", err)
		}
		log.Println("Added follower_count column to users table")
	}

	if !db.Migrator().HasColumn(&model.User{}, "following_count") {
		if err := db.Exec("ALTER TABLE users ADD COLUMN following_count INT DEFAULT 0").Error; err != nil {
			return fmt.Errorf("failed to add following_count column: %w", err)
		}
		log.Println("Added following_count column to users table")
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_follows_follower_id ON follows(follower_id)").Error; err != nil {
		log.Printf("Warning: Failed to create index on follower_id: %v", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_follows_followed_id ON follows(followed_id)").Error; err != nil {
		log.Printf("Warning: Failed to create index on followed_id: %v", err)
	}

	log.Println("Recalculating follower and following counts for all users...")

	if err := db.Exec(`
		UPDATE users u SET follower_count = (
			SELECT COUNT(*) FROM follows f WHERE f.followed_id = u.id
		)
	`).Error; err != nil {
		log.Printf("Warning: Failed to update follower counts: %v", err)
	}

	if err := db.Exec(`
		UPDATE users u SET following_count = (
			SELECT COUNT(*) FROM follows f WHERE f.follower_id = u.id
		)
	`).Error; err != nil {
		log.Printf("Warning: Failed to update following counts: %v", err)
	}

	return nil
}

func migrateAddAccountPrivacy(db *gorm.DB) error {
	if !db.Migrator().HasColumn(&model.User{}, "is_private") {
		if err := db.Exec("ALTER TABLE users ADD COLUMN is_private BOOLEAN DEFAULT FALSE").Error; err != nil {
			return fmt.Errorf("failed to add is_private column: %w", err)
		}
		log.Println("Added is_private column to users table")
	}
	return nil
}

func migrateFixPremiumRequests(db *gorm.DB) error {
	return UpdatePremiumRequestsTable(db)
}