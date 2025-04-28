package repository

import (
	"fmt"
	"time"

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
	if err := s.db.Model(&User{}).Count(&count).Error; err != nil {
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
func getDefaultAuthUsers() []*User {
	// Get current time for timestamps
	now := time.Now()

	// Hash passwords (handle errors in a real implementation)
	adminHash, _ := hashPassword("admin123")
	johnHash, _ := hashPassword("kolin123")
	janeHash, _ := hashPassword("securePass456!")

	// We're creating auth users corresponding to the same users in the user service
	return []*User{
		{
			ID:                    "550e8400-e29b-41d4-a716-446655440000",
			Email:                 "admin@aycom.com",
			Name:                  "Admin User",
			Username:              "admin",
			HashedPassword:        adminHash,
			VerificationCode:      "",
			IsVerified:            true,
			Gender:                "Other",
			DateOfBirth:           "1990-01-01",
			ProfilePicture:        "https://via.placeholder.com/150",
			Banner:                "https://via.placeholder.com/1200x300",
			SecurityQuestion:      "What is your first pet's name?",
			SecurityAnswer:        "Admin",
			SubscribeToNewsletter: false,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    "550e8400-e29b-41d4-a716-446655440001",
			Email:                 "kolin@example.com",
			Name:                  "John Doe",
			Username:              "johndoe",
			HashedPassword:        johnHash,
			VerificationCode:      "",
			IsVerified:            true,
			Gender:                "Male",
			DateOfBirth:           "1995-05-15",
			ProfilePicture:        "https://via.placeholder.com/150",
			Banner:                "https://via.placeholder.com/1200x300",
			SecurityQuestion:      "What is your mother's maiden name?",
			SecurityAnswer:        "Doe",
			SubscribeToNewsletter: true,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    "550e8400-e29b-41d4-a716-446655440002",
			Email:                 "jane@example.com",
			Name:                  "Jane Doe",
			Username:              "janedoe",
			HashedPassword:        janeHash,
			VerificationCode:      "",
			IsVerified:            true,
			Gender:                "Female",
			DateOfBirth:           "1997-08-22",
			ProfilePicture:        "https://via.placeholder.com/150",
			Banner:                "https://via.placeholder.com/1200x300",
			SecurityQuestion:      "What city were you born in?",
			SecurityAnswer:        "New York",
			SubscribeToNewsletter: true,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
	}
}
