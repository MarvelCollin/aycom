package service

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/model"
	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/repository"
)

// AuthService manages authentication-related operations
type AuthService struct {
	DB *gorm.DB
}

// NewAuthService creates a new AuthService instance
func NewAuthService() (*AuthService, error) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get database connection string from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/aycom_auth_db"
		log.Println("Warning: DATABASE_URL not set, using default:", dbURL)
	}

	// Connect to the database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Run migrations using GORM AutoMigrate
	log.Println("Running database migrations...")
	err = repository.RunMigrations(db,
		&model.SecurityQuestion{},
		&model.UserAuth{},
		&model.UserSecurityAnswer{},
		&model.OAuthConnection{},
		&model.Token{},
	)
	if err != nil {
		return nil, err
	}

	return &AuthService{
		DB: db,
	}, nil
}

// GetMigrationStatus prints information about the database tables
func (s *AuthService) GetMigrationStatus() error {
	return repository.GetMigrationStatus(s.DB)
}
