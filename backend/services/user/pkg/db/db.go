package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDatabaseWithRetry establishes a connection to the PostgreSQL database with retries.
func ConnectDatabaseWithRetry() (*gorm.DB, error) {
	// Get database connection details from environment variables
	dbHost := os.Getenv("DATABASE_HOST")
	if dbHost == "" {
		dbHost = "user_db" // Default in docker-compose
	}
	dbPort := os.Getenv("DATABASE_PORT")
	if dbPort == "" {
		dbPort = "5432" // Default PostgreSQL port
	}
	dbUser := os.Getenv("DATABASE_USER")
	if dbUser == "" {
		dbUser = "kolin" // Default user from docker-compose
	}
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	if dbPassword == "" {
		dbPassword = "kolin" // Default password from docker-compose
	}
	dbName := os.Getenv("DATABASE_NAME")
	if dbName == "" {
		dbName = "user_db" // Default database name
	}

	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database with retry mechanism
	var db *gorm.DB
	var err error
	maxRetries := 5
	retryInterval := time.Second * 5

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr != nil {
				log.Printf("Failed to get underlying sql.DB (attempt %d/%d): %v", i+1, maxRetries, dbErr)
				db = nil    // Ensure db is nil if we can't get sql.DB
				err = dbErr // Report the error
			} else if pingErr := sqlDB.Ping(); pingErr == nil {
				log.Println("Database ping successful")
				break // Success
			} else {
				err = fmt.Errorf("failed to ping database: %w", pingErr)
				sqlDB.Close() // Close the potentially invalid connection
				db = nil
			}
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("could not connect to database after %d attempts: %w", maxRetries, err)
	}

	return db, nil
}
