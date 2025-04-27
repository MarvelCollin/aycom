package service

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Acad600-Tpa/WEB-MV-242/services/user/model"
	"github.com/Acad600-Tpa/WEB-MV-242/services/user/repository"
)

// UserService manages user-related operations
type UserService struct {
	DB *gorm.DB
}

// NewUserService creates a new UserService instance
func NewUserService() (*UserService, error) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get database connection string from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/aycom_user_db"
		log.Println("Warning: DATABASE_URL not set, using default:", dbURL)
	}

	// Connect to the database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Run migrations
	migrationDir := filepath.Join("migrations")
	err = repository.RunMigrations(db, repository.MigrationConfig{
		MigrationsDir: migrationDir,
		DBDialect:     "postgres",
	})
	if err != nil {
		return nil, err
	}

	return &UserService{
		DB: db,
	}, nil
}

// AutoMigrate uses GORM's built-in schema migration - alternative to SQL migrations
// This is less recommended for production but useful for development
func (s *UserService) AutoMigrate() error {
	return s.DB.AutoMigrate(
		&model.SecurityQuestion{},
		&model.User{},
	)
}
