package repository

import (
	"fmt"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthSeeder handles the seeding of default users in the auth database
type AuthSeeder struct {
	db *gorm.DB
}

// NewAuthSeeder creates a new auth seeder
func NewAuthSeeder(db *gorm.DB) *AuthSeeder {
	return &AuthSeeder{db: db}
}

// SeedUsers seeds default users for authentication if they don't already exist
func (s *AuthSeeder) SeedUsers() error {
	// Check if we already have users in the database
	var count int64
	if err := s.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	// Skip seeding if users already exist
	if count > 0 {
		fmt.Println("Auth users already exist, skipping seeding")
		return nil
	}

	// Create default users
	users := getDefaultAuthUsers()

	// Insert all users
	for _, user := range users {
		if err := s.db.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create auth user %s: %w", user.Email, err)
		}
		fmt.Printf("Created auth user: %s\n", user.Email)
	}

	fmt.Println("Successfully seeded default auth users")
	return nil
}

// hashPassword creates a bcrypt hash of the password
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// getDefaultAuthUsers returns a slice of default users to seed
func getDefaultAuthUsers() []*model.User {
	// Get current time for timestamps
	now := time.Now()
	dob1, _ := time.Parse("2006-01-02", "1990-01-01")
	dob2, _ := time.Parse("2006-01-02", "1995-05-15")
	dob3, _ := time.Parse("2006-01-02", "1997-08-22")

	// Hash passwords (handle errors in a real implementation)
	adminHash, _ := hashPassword("admin123")
	johnHash, _ := hashPassword("kolin123")
	janeHash, _ := hashPassword("securePass456!")

	// We're creating auth users corresponding to the same users in the user service
	return []*model.User{
		{
			ID:                     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Email:                  "admin@aycom.com",
			Name:                   "Admin User",
			Username:               "admin",
			PasswordHash:           adminHash,
			PasswordSalt:           "",
			VerificationCode:       nil,
			IsActivated:            true,
			Gender:                 "Other",
			DateOfBirth:            dob1,
			SecurityQuestion:       "What is your first pet's name?",
			SecurityAnswer:         "Admin",
			NewsletterSubscription: false,
			JoinedAt:               now,
			CreatedAt:              now,
			UpdatedAt:              now,
		},
		{
			ID:                     uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			Email:                  "kolin@example.com",
			Name:                   "John Doe",
			Username:               "johndoe",
			PasswordHash:           johnHash,
			PasswordSalt:           "",
			VerificationCode:       nil,
			IsActivated:            true,
			Gender:                 "Male",
			DateOfBirth:            dob2,
			SecurityQuestion:       "What is your mother's maiden name?",
			SecurityAnswer:         "Doe",
			NewsletterSubscription: true,
			JoinedAt:               now,
			CreatedAt:              now,
			UpdatedAt:              now,
		},
		{
			ID:                     uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Email:                  "jane@example.com",
			Name:                   "Jane Doe",
			Username:               "janedoe",
			PasswordHash:           janeHash,
			PasswordSalt:           "",
			VerificationCode:       nil,
			IsActivated:            true,
			Gender:                 "Female",
			DateOfBirth:            dob3,
			SecurityQuestion:       "What city were you born in?",
			SecurityAnswer:         "New York",
			NewsletterSubscription: true,
			JoinedAt:               now,
			CreatedAt:              now,
			UpdatedAt:              now,
		},
	}
}
