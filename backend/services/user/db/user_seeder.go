package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SeedConfig struct {
	AdminPassword  string
	RegularUserPwd string
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetSeedConfig() SeedConfig {
	return SeedConfig{
		AdminPassword:  getEnv("SEED_ADMIN_PASSWORD", "admin123"),
		RegularUserPwd: getEnv("SEED_USER_PASSWORD", "password123"),
	}
}

type UserSeeder struct {
	db     *gorm.DB
	config SeedConfig
}

func NewUserSeeder(db *gorm.DB) *UserSeeder {
	return &UserSeeder{
		db:     db,
		config: GetSeedConfig(),
	}
}

func (s *UserSeeder) SeedUsers() error {
	var count int64
	s.db.Table("users").Count(&count)
	if count > 0 {
		log.Println("Users already exist, skipping seeding")
		return nil
	}

	now := time.Now()
	dob1, _ := time.Parse("2006-01-02", "1990-01-01")
	dob2, _ := time.Parse("2006-01-02", "1995-05-15")
	dob3, _ := time.Parse("2006-01-02", "1997-08-22")

	adminHash, _ := bcrypt.GenerateFromPassword([]byte(s.config.AdminPassword), bcrypt.DefaultCost)
	userHash, _ := bcrypt.GenerateFromPassword([]byte(s.config.RegularUserPwd), bcrypt.DefaultCost)

	adminID := uuid.MustParse(getEnv("ADMIN_UUID", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"))
	johnID := uuid.MustParse(getEnv("JOHN_UUID", "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12"))
	janeID := uuid.MustParse(getEnv("JANE_UUID", "c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13"))

	type User struct {
		ID                     uuid.UUID `gorm:"type:uuid;primary_key"`
		Email                  string    `gorm:"size:255;uniqueIndex;not null"`
		Name                   string    `gorm:"size:255;not null"`
		Username               string    `gorm:"size:255;uniqueIndex;not null"`
		PasswordHash           string    `gorm:"size:255;not null"`
		PasswordSalt           string    `gorm:"size:255"`
		VerificationCode       *string
		IsActivated            bool
		Gender                 string
		DateOfBirth            time.Time
		SecurityQuestion       string
		SecurityAnswer         string
		NewsletterSubscription bool
		JoinedAt               time.Time
		CreatedAt              time.Time
		UpdatedAt              time.Time
	}

	users := []User{
		{
			ID:                     adminID,
			Email:                  "admin@aycom.com",
			Name:                   "Admin User",
			Username:               "admin",
			PasswordHash:           string(adminHash),
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
			ID:                     johnID,
			Email:                  "john@example.com",
			Name:                   "John Doe",
			Username:               "johndoe",
			PasswordHash:           string(userHash),
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
			ID:                     janeID,
			Email:                  "jane@example.com",
			Name:                   "Jane Doe",
			Username:               "janedoe",
			PasswordHash:           string(userHash),
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

	if err := s.db.Create(&users).Error; err != nil {
		return fmt.Errorf("failed to create users: %w", err)
	}

	log.Printf("Created %d users", len(users))
	return nil
}

func (s *UserSeeder) SeedSessions() error {
	var count int64
	s.db.Table("sessions").Count(&count)
	if count > 0 {
		log.Println("Sessions already exist, skipping seeding")
		return nil
	}

	var adminUser struct {
		ID uuid.UUID
	}
	s.db.Table("users").Where("username = ?", "admin").First(&adminUser)

	session := struct {
		ID           uuid.UUID
		UserID       uuid.UUID
		AccessToken  string
		RefreshToken string
		IPAddress    string
		UserAgent    string
		ExpiresAt    time.Time
	}{
		ID:           uuid.New(),
		UserID:       adminUser.ID,
		AccessToken:  "test_access_token_" + uuid.New().String(),
		RefreshToken: "test_refresh_token_" + uuid.New().String(),
		IPAddress:    "127.0.0.1",
		UserAgent:    "Test User Agent",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	if err := s.db.Table("sessions").Create(&session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	log.Println("Created test session for admin user")
	return nil
}

func (s *UserSeeder) SeedAll() error {
	if err := s.SeedUsers(); err != nil {
		return err
	}
	if err := s.SeedSessions(); err != nil {
		return err
	}
	return nil
}
