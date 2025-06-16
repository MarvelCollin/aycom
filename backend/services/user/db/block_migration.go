package db

import (
	"log"

	"gorm.io/gorm"

	"aycom/backend/services/user/model"
)

func AddBlockAndReportTables(db *gorm.DB) error {

	if err := db.AutoMigrate(&model.UserBlock{}); err != nil {
		log.Printf("Failed to migrate user_blocks table: %v", err)
		return err
	}
	log.Println("Successfully migrated user_blocks table")

	if err := db.AutoMigrate(&model.UserReport{}); err != nil {
		log.Printf("Failed to migrate user_reports table: %v", err)
		return err
	}
	log.Println("Successfully migrated user_reports table")

	return nil
}
