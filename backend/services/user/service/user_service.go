package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/model"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// UserService defines the methods for user-related operations
type UserService interface {
	CreateUser(ctx context.Context, userId, username, email, name, gender, dateOfBirth, profilePicture, banner, secQuestion, secAnswer string, subscribeToNewsletter bool) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUserProfile(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error)
	UpdateUserVerification(ctx context.Context, userID string, isVerified bool) error
	DeleteUser(ctx context.Context, id string) error
}

// userService implements the UserService interface
type userService struct {
	repo repository.UserRepository
}

// UserService struct that also exposes DB for main.go
type UserServiceImpl struct {
	DB   *gorm.DB
	repo repository.UserRepository
	UserService
}

// NewUserService creates a new user service with DB connection for main.go compatibility
func NewUserService() (*UserServiceImpl, error) {
	// Get database connection details from environment variables
	dbHost := os.Getenv("DATABASE_HOST")
	if dbHost == "" {
		dbHost = "postgres" // Default in docker-compose
	}

	dbPort := os.Getenv("DATABASE_PORT")
	if dbPort == "" {
		dbPort = "5432" // Default PostgreSQL port
	}

	dbUser := os.Getenv("DATABASE_USER")
	if dbUser == "" {
		dbUser = "postgres" // Default user
	}

	dbPassword := os.Getenv("DATABASE_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres_password" // Default password
	}

	dbName := os.Getenv("DATABASE_NAME")
	if dbName == "" {
		dbName = "user_db" // Default database name
	}

	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database with retry mechanism
	var db *gorm.DB
	var err error
	maxRetries := 5
	retryInterval := time.Second * 5

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
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

	log.Println("Successfully connected to the database")

	// Create repository
	repo := repository.NewPostgresUserRepository(db)

	// Create service
	svc := newUserService(repo)

	// Return service implementation with DB access
	return &UserServiceImpl{
		DB:          db,
		repo:        repo,
		UserService: svc,
	}, nil
}

// GetMigrationStatus checks migration status and runs migrations if needed
func (s *UserServiceImpl) GetMigrationStatus() error {
	// Run auto migration
	if err := s.DB.AutoMigrate(&model.User{}); err != nil {
		return fmt.Errorf("failed to migrate user table: %w", err)
	}

	log.Println("User table migration complete")
	return nil
}

// Create the original user service without DB access
func newUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

// CreateUser creates a new user in the system
func (s *userService) CreateUser(ctx context.Context, userId, username, email, name, gender, dateOfBirth, profilePicture, banner, secQuestion, secAnswer string, subscribeToNewsletter bool) (*model.User, error) {
	// Check if user with same email already exists
	existingUser, err := s.repo.FindUserByEmail(email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Check if user with same username already exists
	existingUser, err = s.repo.FindUserByUsername(username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already taken")
	}

	// Create new user
	user := model.NewUser(
		userId,
		username,
		email,
		name,
		gender,
		dateOfBirth,
		profilePicture,
		banner,
		secQuestion,
		secAnswer,
		subscribeToNewsletter,
	)

	// Save user to database
	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID gets a user by their ID
func (s *userService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	return s.repo.FindUserByID(id)
}

// GetUserByUsername gets a user by their username
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.repo.FindUserByUsername(username)
}

// UpdateUserProfile updates a user's profile information
func (s *userService) UpdateUserProfile(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
	// Get current user data
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok && name != "" {
		user.Name = name
	}

	if gender, ok := updates["gender"].(string); ok && gender != "" {
		user.Gender = gender
	}

	if dobStr, ok := updates["date_of_birth"].(string); ok && dobStr != "" {
		dob, err := time.Parse("2006-01-02", dobStr)
		if err == nil {
			user.DateOfBirth = dob
		}
	}

	if profilePic, ok := updates["profile_picture"].(string); ok && profilePic != "" {
		user.ProfilePicture = profilePic
	}

	if banner, ok := updates["banner"].(string); ok && banner != "" {
		user.Banner = banner
	}

	if bio, ok := updates["bio"].(string); ok {
		user.Bio = bio
	}

	user.UpdatedAt = time.Now()

	// Save updated user
	err = s.repo.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserVerification updates a user's verification status
func (s *userService) UpdateUserVerification(ctx context.Context, userID string, isVerified bool) error {
	return s.repo.UpdateUserVerification(userID, isVerified)
}

// DeleteUser deletes a user by their ID
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(id)
}
