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
