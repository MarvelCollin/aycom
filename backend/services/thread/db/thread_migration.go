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
