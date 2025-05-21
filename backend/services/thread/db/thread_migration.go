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

	// Migrate the base entities first
	err = db.AutoMigrate(
		&model.Thread{},
		&model.Reply{},
		&model.Media{},
		&model.Hashtag{},
		&model.Category{},
		&model.Poll{},
		&model.PollOption{},
		&model.PollVote{},
		&model.UserMention{},
		&model.ThreadHashtag{},
		&model.ThreadCategory{},
	)

	if err != nil {
		log.Printf("Base migrations failed: %v", err)
		return err
	}

	// Fix the likes table structure - use a transaction for safety
	log.Println("Checking if likes table needs fixes...")
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	if db.Migrator().HasTable(&model.Like{}) {
		// Check if there's a not-null constraint on reply_id
		var count int64
		tx.Raw("SELECT count(*) FROM information_schema.table_constraints tc JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name WHERE tc.table_name = 'likes' AND ccu.column_name = 'reply_id' AND tc.constraint_type = 'PRIMARY KEY'").Count(&count)

		if count > 0 {
			log.Println("Fixing likes table structure...")
			// Drop old table
			if err := tx.Migrator().DropTable(&model.Like{}); err != nil {
				log.Printf("Failed to drop likes table: %v", err)
				tx.Rollback()
				return err
			}

			// Create it with new structure
			if err := tx.Migrator().CreateTable(&model.Like{}); err != nil {
				log.Printf("Failed to recreate likes table: %v", err)
				tx.Rollback()
				return err
			}
			log.Println("Successfully fixed likes table structure")
		}
	} else {
		// Table doesn't exist, create it
		log.Println("Creating likes table...")
		if err := tx.Migrator().CreateTable(&model.Like{}); err != nil {
			log.Printf("Failed to create likes table: %v", err)
			tx.Rollback()
			return err
		}

		// Add proper indexes for better performance on read operations
		if err := tx.Exec("CREATE INDEX idx_likes_thread_id ON likes (thread_id) WHERE thread_id IS NOT NULL").Error; err != nil {
			log.Printf("Warning: Failed to create thread_id index on likes: %v", err)
			// Continue anyway, this is not critical
		}

		if err := tx.Exec("CREATE INDEX idx_likes_reply_id ON likes (reply_id) WHERE reply_id IS NOT NULL").Error; err != nil {
			log.Printf("Warning: Failed to create reply_id index on likes: %v", err)
			// Continue anyway, this is not critical
		}
	}

	// Fix the bookmarks table structure to support reply bookmarks
	log.Println("Checking if bookmarks table needs updates for reply support...")

	// Check if the current primary key is (user_id, thread_id) without considering reply_id
	var bookmarkPrimaryKeyCount int64
	tx.Raw("SELECT count(*) FROM information_schema.table_constraints tc JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name WHERE tc.table_name = 'bookmarks' AND tc.constraint_type = 'PRIMARY KEY' AND ccu.column_name IN ('user_id', 'thread_id') AND NOT EXISTS (SELECT 1 FROM information_schema.constraint_column_usage WHERE constraint_name = tc.constraint_name AND column_name = 'reply_id')").Count(&bookmarkPrimaryKeyCount)

	if bookmarkPrimaryKeyCount > 0 {
		log.Println("Bookmark table needs to be updated to properly support reply bookmarks...")

		// Check if we need to add reply_id column first
		var hasReplyIdColumn int64
		tx.Raw("SELECT count(*) FROM information_schema.columns WHERE table_name = 'bookmarks' AND column_name = 'reply_id'").Count(&hasReplyIdColumn)

		if hasReplyIdColumn == 0 {
			// Add reply_id column if it doesn't exist
			log.Println("Adding reply_id column to bookmarks table...")
			if err := tx.Exec("ALTER TABLE bookmarks ADD COLUMN reply_id UUID NULL").Error; err != nil {
				log.Printf("Failed to add reply_id column: %v", err)
				tx.Rollback()
				return err
			}
		}

		// Remove the existing primary key constraint
		log.Println("Removing existing primary key from bookmarks table...")
		if err := tx.Exec("ALTER TABLE bookmarks DROP CONSTRAINT bookmarks_pkey").Error; err != nil {
			log.Printf("Failed to drop primary key constraint: %v", err)
			tx.Rollback()
			return err
		}

		// Add a CHECK constraint to ensure either thread_id is not NULL or reply_id is not NULL, but not both
		log.Println("Adding check constraint to ensure valid bookmark state...")
		if err := tx.Exec("ALTER TABLE bookmarks ADD CONSTRAINT bookmark_valid_target CHECK ((thread_id IS NOT NULL AND reply_id IS NULL) OR (thread_id IS NULL AND reply_id IS NOT NULL))").Error; err != nil {
			log.Printf("Warning: Failed to add check constraint: %v", err)
			// This might fail if the constraint already exists, continue anyway
		}

		// Add appropriate unique constraints
		log.Println("Adding proper unique constraints for bookmarks...")
		if err := tx.Exec("ALTER TABLE bookmarks ADD CONSTRAINT bookmarks_user_thread_unique UNIQUE (user_id, thread_id) WHERE thread_id IS NOT NULL").Error; err != nil {
			log.Printf("Warning: Failed to add thread unique constraint: %v", err)
			// Continue anyway as this might already exist
		}

		if err := tx.Exec("ALTER TABLE bookmarks ADD CONSTRAINT bookmarks_user_reply_unique UNIQUE (user_id, reply_id) WHERE reply_id IS NOT NULL").Error; err != nil {
			log.Printf("Warning: Failed to add reply unique constraint: %v", err)
			// Continue anyway as this might already exist
		}

		// Make user_id NOT NULL
		log.Println("Setting user_id to NOT NULL if not already set...")
		if err := tx.Exec("ALTER TABLE bookmarks ALTER COLUMN user_id SET NOT NULL").Error; err != nil {
			log.Printf("Warning: Failed to set user_id NOT NULL: %v", err)
			// Continue anyway as this might already be set
		}

		log.Println("Successfully updated bookmark table structure")
	} else {
		log.Println("Bookmarks table already has proper structure")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit migration changes: %v", err)
		return err
	}

	// Now run migrations for interaction models
	err = db.AutoMigrate(
		&model.Like{},
		&model.Repost{},
		&model.Bookmark{},
	)

	if err != nil {
		log.Printf("Interaction models migration failed: %v", err)
		return err
	}

	log.Println("Thread service database migrations completed successfully")
	return nil
}
