package db

import (
	"log"

	"aycom/backend/services/community/model"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	log.Println("Running community service database migrations...")
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		return err
	}
	// Migrate all models
	err = db.AutoMigrate(
		&model.Community{},
		&model.CommunityMember{},
		&model.CommunityRule{},
		&model.CommunityJoinRequest{},
		&model.Category{},
		&model.CommunityCategory{},
		&model.Chat{},
		&model.ChatParticipant{},
		&model.Message{},
		&model.DeletedChat{},
	)
	if err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}
	log.Println("Community service database migrations completed successfully")
	return nil
}
