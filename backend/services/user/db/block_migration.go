package db

import (
	"log"

	"aycom/backend/services/user/model"

	"gorm.io/gorm"
)

// AddBlockAndReportTables adds the block and report tables to the database
func AddBlockAndReportTables(db *gorm.DB) error {
	// Create block table
	if err := db.AutoMigrate(&model.UserBlock{}); err != nil {
		log.Printf("Failed to migrate user_blocks table: %v", err)
		return err
	}
	log.Println("Successfully migrated user_blocks table")

	// Create report table
	if err := db.AutoMigrate(&model.UserReport{}); err != nil {
		log.Printf("Failed to migrate user_reports table: %v", err)
		return err
	}
	log.Println("Successfully migrated user_reports table")

	return nil
}
