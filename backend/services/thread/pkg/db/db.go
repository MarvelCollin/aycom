package db

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDatabase connects to the PostgreSQL database using environment variables
func ConnectDatabase() (*gorm.DB, error) {
	// Get database connection details from environment variables
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost" // Default for local development only
	}
	
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432" // Default PostgreSQL port
	}
	
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres" // Default for local development only
	}
	
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		// Log a warning about using default credentials in production
		log.Println("Warning: Using default database password. This should only be used for development.")
		dbPassword = "postgres" // Default for local development only
	}
	
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "thread_db" // Default database name
	}
	
	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable" // Default to disable for local development
	}
	
	// Log connection attempt (without exposing credentials)
	log.Printf("Connecting to database %s on %s:%s with user %s (SSL mode: %s)", 
		dbName, dbHost, dbPort, dbUser, dbSSLMode)

	// Construct connection string
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode,
	)

	// Set up database logging
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Enable color
		},
	)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get DB instance: %w", err)
	}

	// Get connection pool settings from environment or use sensible defaults
	maxIdleConns := getEnvInt("DB_MAX_IDLE_CONNS", 10)
	maxOpenConns := getEnvInt("DB_MAX_OPEN_CONNS", 100)
	connMaxLifetime := getEnvDuration("DB_CONN_MAX_LIFETIME", time.Hour)

	// Set connection pool limits
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	
	log.Println("Database connection established successfully")
	return db, nil
}

// ConnectDatabaseWithRetry attempts to connect to the database with retries
func ConnectDatabaseWithRetry() (*gorm.DB, error) {
	maxRetries := 5
	retryInterval := 5 * time.Second

	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = ConnectDatabase()
		if err == nil {
			return db, nil
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)

		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt gets an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Warning: Environment variable %s is not a valid integer. Using default value %d", key, defaultValue)
		return defaultValue
	}
	
	return intValue
}

// getEnvDuration gets a duration environment variable or returns a default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("Warning: Environment variable %s is not a valid duration. Using default value %s", key, defaultValue)
		return defaultValue
	}
	
	return duration
}
