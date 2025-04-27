package repository

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// RunMigrations runs database migrations using GORM's AutoMigrate
func RunMigrations(db *gorm.DB, models ...interface{}) error {
	log.Println("Running GORM auto-migrations for auth service...")

	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Auth service migrations completed successfully")
	return nil
}

// GetMigrationStatus returns the status of tables in the database
func GetMigrationStatus(db *gorm.DB) error {
	var tables []string

	// Get list of tables in the database
	if err := db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tables).Error; err != nil {
		return fmt.Errorf("failed to get table information: %w", err)
	}

	log.Println("Auth database tables:")
	for _, table := range tables {
		log.Printf("- %s", table)
	}

	return nil
}
