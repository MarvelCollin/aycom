package db

import (
	"log"

	"aycom/backend/services/thread/model"

	"gorm.io/gorm"
)

// RunMigrations runs all the database migrations for thread service models
func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Ensure valid connection
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		return err
	}

	// Fix the likes table structure - drop the table first
	log.Println("Checking if likes table needs fixes...")
	if db.Migrator().HasTable(&model.Like{}) {
		// Check if there's a not-null constraint on reply_id
		var count int64
		db.Raw("SELECT count(*) FROM information_schema.table_constraints tc JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name WHERE tc.table_name = 'likes' AND ccu.column_name = 'reply_id' AND tc.constraint_type = 'PRIMARY KEY'").Count(&count)

		if count > 0 {
			log.Println("Fixing likes table structure...")
			// Drop old table
			if err := db.Migrator().DropTable(&model.Like{}); err != nil {
				log.Printf("Failed to drop likes table: %v", err)
				return err
			}

			// Create it with new structure
			if err := db.Migrator().CreateTable(&model.Like{}); err != nil {
				log.Printf("Failed to recreate likes table: %v", err)
				return err
			}
			log.Println("Successfully fixed likes table structure")
		}
	}

	// Fix the bookmarks table structure to support reply bookmarks
	log.Println("Checking if bookmarks table needs updates for reply support...")
	if db.Migrator().HasTable(&model.Bookmark{}) {
		// Check if the table has a reply_id column
		hasReplyIDColumn := db.Migrator().HasColumn(&model.Bookmark{}, "reply_id")

		if !hasReplyIDColumn {
			log.Println("Updating bookmarks table to support reply bookmarks...")

			// Add reply_id column if it doesn't exist
			if err := db.Exec("ALTER TABLE bookmarks ADD COLUMN reply_id UUID NULL").Error; err != nil {
				log.Printf("Failed to add reply_id column to bookmarks table: %v", err)
				return err
			}

			// We need to modify the primary key constraint to work with replies
			// First drop the primary key
			if err := db.Exec("ALTER TABLE bookmarks DROP CONSTRAINT IF EXISTS bookmarks_pkey").Error; err != nil {
				log.Printf("Failed to drop primary key from bookmarks table: %v", err)
				return err
			}

			// Add new primary key constraint
			if err := db.Exec("ALTER TABLE bookmarks ADD CONSTRAINT bookmarks_pkey PRIMARY KEY (user_id, thread_id)").Error; err != nil {
				log.Printf("Failed to add new primary key to bookmarks table: %v", err)
				return err
			}

			// Create a unique index for reply bookmarks to ensure uniqueness when reply_id is used
			if err := db.Exec("CREATE UNIQUE INDEX idx_bookmarks_user_reply ON bookmarks (user_id, reply_id) WHERE reply_id IS NOT NULL").Error; err != nil {
				log.Printf("Failed to create unique index for reply bookmarks: %v", err)
				return err
			}

			log.Println("Successfully updated bookmarks table to support reply bookmarks")
		}
	}

	// Run migrations for all thread-related models
	err = db.AutoMigrate(
		// Base entities
		&model.Thread{},
		&model.Reply{},
		&model.Media{},
		&model.Hashtag{},
		&model.Category{},
		&model.Poll{},
		&model.PollOption{},
		&model.PollVote{},
		&model.UserMention{},

		// Junction tables and interaction models
		&model.ThreadHashtag{},
		&model.ThreadCategory{},
		&model.Like{},
		&model.Repost{},
		&model.Bookmark{},
	)

	if err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

	log.Println("Thread service database migrations completed successfully")
	return nil
}
