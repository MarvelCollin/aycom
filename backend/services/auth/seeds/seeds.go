package seeds

import (
	"log"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedAll runs all seeders
func SeedAll(db *gorm.DB) error {
	if err := SeedUsers(db); err != nil {
		return err
	}

	if err := SeedSessions(db); err != nil {
		return err
	}

	return nil
}

// SeedUsers adds test users to the database
func SeedUsers(db *gorm.DB) error {
	// Check if users already exist
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		log.Println("Users table already has data, skipping seeding")
		return nil
	}

	// Generate password and salt
	password := "password123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// In a real implementation, you would generate a unique salt for each user
	salt := "testsalt"

	// Current time for joined_at
	now := time.Now()

	// Fixed UUIDs for test data
	adminID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	testUserID := uuid.MustParse("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12")

	// Create test users
	users := []model.User{
		{
			ID:                        adminID,
			Name:                      "Admin User",
			Username:                  "admin",
			Email:                     "admin@example.com",
			PasswordHash:              string(passwordHash),
			PasswordSalt:              salt,
			Gender:                    "male",
			DateOfBirth:               time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			SecurityQuestion:          "What is your favorite color?",
			SecurityAnswer:            "blue",
			GoogleID:                  "",
			IsActivated:               true,
			IsBanned:                  false,
			IsDeactivated:             false,
			IsAdmin:                   true,
			NewsletterSubscription:    true,
			LastLoginAt:               &now,
			JoinedAt:                  now,
			VerificationCode:          nil,
			VerificationCodeExpiresAt: nil,
		},
		{
			ID:                        testUserID,
			Name:                      "Test User",
			Username:                  "testuser",
			Email:                     "test@example.com",
			PasswordHash:              string(passwordHash),
			PasswordSalt:              salt,
			Gender:                    "female",
			DateOfBirth:               time.Date(1995, 5, 15, 0, 0, 0, 0, time.UTC),
			SecurityQuestion:          "What is your pet's name?",
			SecurityAnswer:            "fluffy",
			GoogleID:                  "",
			IsActivated:               true,
			IsBanned:                  false,
			IsDeactivated:             false,
			IsAdmin:                   false,
			NewsletterSubscription:    false,
			LastLoginAt:               &now,
			JoinedAt:                  now,
			VerificationCode:          nil,
			VerificationCodeExpiresAt: nil,
		},
	}

	result := db.Create(&users)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Created %d test users", len(users))
	return nil
}

// SeedSessions adds test sessions to the database
func SeedSessions(db *gorm.DB) error {
	// Check if sessions already exist
	var count int64
	db.Model(&model.Session{}).Count(&count)
	if count > 0 {
		log.Println("Sessions table already has data, skipping seeding")
		return nil
	}

	// Fixed UUIDs for test data
	adminID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	sessionID := uuid.MustParse("c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13")

	// Create test session
	sessions := []model.Session{
		{
			ID:           sessionID,
			UserID:       adminID,
			AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhMGVlYmM5OS05YzBiLTRlZjgtYmI2ZC02YmI5YmQzODBhMTEiLCJuYW1lIjoiQWRtaW4gVXNlciIsImlhdCI6MTUxNjIzOTAyMn0.XG0fIRH_tga1vbRxqQr3S0aKd5OGxhXKFNZwwZDIZlc",
			RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhMGVlYmM5OS05YzBiLTRlZjgtYmI2ZC02YmI5YmQzODBhMTEiLCJyZWYiOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.8O_MaAjTDfmXYOPiQeXnP-YzpkQKfMWZ4qleDSEfB5c",
			IPAddress:    "127.0.0.1",
			UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		},
	}

	result := db.Create(&sessions)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Created %d test sessions", len(sessions))
	return nil
}
