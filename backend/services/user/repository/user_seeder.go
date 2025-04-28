package repository

import (
	"fmt"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/model"
	"gorm.io/gorm"
)

// UserSeeder handles the seeding of default users
type UserSeeder struct {
	db *gorm.DB
}

// NewUserSeeder creates a new user seeder
func NewUserSeeder(db *gorm.DB) *UserSeeder {
	return &UserSeeder{db: db}
}

// SeedUsers seeds default users if they don't already exist
func (s *UserSeeder) SeedUsers() error {
	// Check if we already have users in the database
	var count int64
	if err := s.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	// Skip seeding if users already exist
	if count > 0 {
		fmt.Println("Users already exist, skipping seeding")
		return nil
	}

	// Create default users
	users := getDefaultUsers()

	// Insert all users
	for _, user := range users {
		if err := s.db.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
		fmt.Printf("Created user: %s\n", user.Username)
	}

	fmt.Println("Successfully seeded default users")
	return nil
}

// getDefaultUsers returns a slice of default users to seed
func getDefaultUsers() []*model.User {
	// Get current time for timestamps
	now := time.Now()

	// Define a few default user accounts
	return []*model.User{
		{
			ID:                    "550e8400-e29b-41d4-a716-446655440000",
			Username:              "admin",
			Email:                 "admin@aycom.com",
			Name:                  "Admin User",
			Gender:                "Other",
			DateOfBirth:           time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			ProfilePicture:        "https://via.placeholder.com/150",
			Banner:                "https://via.placeholder.com/1200x300",
			Bio:                   "I am the administrator of this platform.",
			Followers:             0,
			Following:             0,
			IsVerified:            true,
			SecurityQuestion:      "What is your first pet's name?",
			SecurityAnswer:        "Admin",
			SubscribeToNewsletter: false,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    "550e8400-e29b-41d4-a716-446655440001",
			Username:              "johndoe",
			Email:                 "kolin@example.com", // Fixed to match the auth service
			Name:                  "John Doe",
			Gender:                "Male",
			DateOfBirth:           time.Date(1995, 5, 15, 0, 0, 0, 0, time.UTC),
			ProfilePicture:        "https://via.placeholder.com/150",
			Banner:                "https://via.placeholder.com/1200x300",
			Bio:                   "Hello, I'm John Doe. I love coding and connecting with people.",
			Followers:             0,
			Following:             0,
			IsVerified:            true,
			SecurityQuestion:      "What is your mother's maiden name?",
			SecurityAnswer:        "Doe",
			SubscribeToNewsletter: true,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    "550e8400-e29b-41d4-a716-446655440002",
			Username:              "janedoe",
			Email:                 "jane@example.com",
			Name:                  "Jane Doe",
			Gender:                "Female",
			DateOfBirth:           time.Date(1997, 8, 22, 0, 0, 0, 0, time.UTC),
			ProfilePicture:        "https://via.placeholder.com/150",
			Banner:                "https://via.placeholder.com/1200x300",
			Bio:                   "Designer, photographer, and tech enthusiast.",
			Followers:             0,
			Following:             0,
			IsVerified:            true,
			SecurityQuestion:      "What city were you born in?",
			SecurityAnswer:        "New York",
			SubscribeToNewsletter: true,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
	}
}
