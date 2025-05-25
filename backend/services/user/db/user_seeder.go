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
	dob4, _ := time.Parse("2006-01-02", "1992-03-10")
	dob5, _ := time.Parse("2006-01-02", "1993-07-05")
	dob6, _ := time.Parse("2006-01-02", "1996-11-18")
	dob7, _ := time.Parse("2006-01-02", "1994-04-27")
	dob8, _ := time.Parse("2006-01-02", "1991-09-14")

	password := []byte("Miawmiaw123@")
	passwordHash, _ := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

	adminID := uuid.MustParse(getEnv("ADMIN_UUID", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"))
	johnID := uuid.MustParse(getEnv("JOHN_UUID", "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12"))
	janeID := uuid.MustParse(getEnv("JANE_UUID", "c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13"))
	samID := uuid.MustParse(getEnv("SAM_UUID", "d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14"))
	techGuruID := uuid.New()
	fitnessCoachID := uuid.New()
	travelBugID := uuid.New()
	foodieID := uuid.New()

	type User struct {
		ID                     uuid.UUID `gorm:"type:uuid;primary_key"`
		Email                  string    `gorm:"size:255;uniqueIndex;not null"`
		Name                   string    `gorm:"size:255;not null"`
		Username               string    `gorm:"size:255;uniqueIndex;not null"`
		PasswordHash           string    `gorm:"size:255;not null"`
		PasswordSalt           string    `gorm:"size:255"`
		VerificationCode       *string
		IsActivated            bool
		IsVerified             bool
		IsAdmin                bool `gorm:"default:false"`
		IsBanned               bool `gorm:"default:false"`
		FollowerCount          int
		Gender                 string
		DateOfBirth            time.Time
		SecurityQuestion       string
		SecurityAnswer         string
		NewsletterSubscription bool
		JoinedAt               time.Time
		CreatedAt              time.Time
		UpdatedAt              time.Time
		Bio                    string
		ProfilePictureURL      string
	}

	users := []User{
		{
			ID:                     adminID,
			Email:                  "admin@aycom.com",
			Name:                   "Admin User",
			Username:               "admin",
			PasswordHash:           string(passwordHash),
			PasswordSalt:           "",
			VerificationCode:       nil,
			IsActivated:            true,
			IsVerified:             true,
			IsAdmin:                true,
			FollowerCount:          9500,
			Gender:                 "Other",
			DateOfBirth:            dob1,
			SecurityQuestion:       "What is your first pet's name?",
			SecurityAnswer:         "Admin",
			NewsletterSubscription: false,
			JoinedAt:               now,
			CreatedAt:              now,
			UpdatedAt:              now,
			Bio:                    "Platform administrator and tech enthusiast",
			ProfilePictureURL:      "https://randomuser.me/api/portraits/lego/1.jpg",
		},
		{
			ID:                     johnID,
			Email:                  "john@example.com",
			Name:                   "John Doe",
			Username:               "johndoe",
			PasswordHash:           string(passwordHash),
			PasswordSalt:           "",
			VerificationCode:       nil,
			IsActivated:            true,
			IsVerified:             true,
			IsAdmin:                false,
			FollowerCount:          4200,
			Gender:                 "Male",
			DateOfBirth:            dob2,
			SecurityQuestion:       "What is your mother's maiden name?",
			SecurityAnswer:         "Doe",
			NewsletterSubscription: true,
			JoinedAt:               now,
			CreatedAt:              now,
			UpdatedAt:              now,
			Bio:                    "Software developer and open source contributor",
			ProfilePictureURL:      "https://randomuser.me/api/portraits/men/1.jpg",
		},
		{
			ID:                     janeID,
			Email:                  "jane@example.com",
			Name:                   "Jane Doe",
			Username:               "janedoe",
			PasswordHash:           string(passwordHash),
			PasswordSalt:           "",
			VerificationCode:       nil,
			IsActivated:            true,
			IsVerified:             true,
			IsAdmin:                false,
			FollowerCount:          6300,
			Gender:                 "Female",
			DateOfBirth:            dob3,
			SecurityQuestion:       "What city were you born in?",
			SecurityAnswer:         "New York",
			NewsletterSubscription: true,
			JoinedAt:               now,
			CreatedAt:              now,
			UpdatedAt:              now,
			Bio:                    "UX designer and art enthusiast",
			ProfilePictureURL:      "https://randomuser.me/api/portraits/women/1.jpg",
		},
		{
			ID:                     samID,
			Email:                  "sam@example.com",
			Name:                   "Sam Smith",
			Username:               "samsmith",
			PasswordHash:           string(passwordHash),
			PasswordSalt:           "",
			VerificationCode:       nil,
			IsActivated:            true,
			IsVerified:             false,
			IsAdmin:                false,
			FollowerCount:          2100,
			Gender:                 "Male",
			DateOfBirth:            dob4,
			SecurityQuestion:       "What is your favorite color?",
			SecurityAnswer:         "Blue",
			NewsletterSubscription: false,
			JoinedAt:               now,
			CreatedAt:              now,
			UpdatedAt:              now,
			Bio:                    "Marketing specialist and hobby photographer",
			ProfilePictureURL:      "https://randomuser.me/api/portraits/men/2.jpg",
		},
		{
			ID:                     techGuruID,
			Email:                  "techguru@example.com",
			Name:                   "Tech Guru",
			Username:               "techguru",
			PasswordHash:           string(passwordHash),
			PasswordSalt:           "",
			VerificationCode:       nil,
			IsActivated:            true,
			IsVerified:             true,
			IsAdmin:                false,
			FollowerCount:          45000,
			Gender:                 "Other",
			DateOfBirth:            dob5,
			SecurityQuestion:       "What was your first car?",
			SecurityAnswer:         "Tesla",
			NewsletterSubscription: true,
			JoinedAt:               now.Add(-180 * 24 * time.Hour),
			CreatedAt:              now.Add(-180 * 24 * time.Hour),
			UpdatedAt:              now,
			Bio:                    "Technology reviewer and industry analyst with 10+ years experience",
			ProfilePictureURL:      "https://randomuser.me/api/portraits/men/3.jpg",
		},
		{
			ID:                     fitnessCoachID,
			Email:                  "fitness@example.com",
			Name:                   "Fitness Coach",
			Username:               "fitnesscoach",
			PasswordHash:           string(passwordHash),
			PasswordSalt:           "",
			VerificationCode:       nil,
			IsActivated:            true,
			IsVerified:             true,
			IsAdmin:                false,
			FollowerCount:          32800,
			Gender:                 "Female",
			DateOfBirth:            dob6,
			SecurityQuestion:       "What is your favorite exercise?",
			SecurityAnswer:         "Squats",
			NewsletterSubscription: true,
			JoinedAt:               now.Add(-150 * 24 * time.Hour),
			CreatedAt:              now.Add(-150 * 24 * time.Hour),
			UpdatedAt:              now,
			Bio:                    "Certified personal trainer and nutrition specialist",
			ProfilePictureURL:      "https://randomuser.me/api/portraits/women/2.jpg",
		},
		{
			ID:                     travelBugID,
			Email:                  "travel@example.com",
			Name:                   "Travel Bug",
			Username:               "travelbug",
			PasswordHash:           string(passwordHash),
			PasswordSalt:           "",
			VerificationCode:       nil,
			IsActivated:            true,
			IsVerified:             false,
			IsAdmin:                false,
			FollowerCount:          3250,
			Gender:                 "Male",
			DateOfBirth:            dob7,
			SecurityQuestion:       "What's your favorite travel destination?",
			SecurityAnswer:         "Bali",
			NewsletterSubscription: true,
			JoinedAt:               now.Add(-120 * 24 * time.Hour),
			CreatedAt:              now.Add(-120 * 24 * time.Hour),
			UpdatedAt:              now,
			Bio:                    "Digital nomad exploring the world one country at a time",
			ProfilePictureURL:      "https://randomuser.me/api/portraits/men/4.jpg",
		},
		{
			ID:                     foodieID,
			Email:                  "foodie@example.com",
			Name:                   "Food Explorer",
			Username:               "foodie123",
			PasswordHash:           string(passwordHash),
			PasswordSalt:           "",
			VerificationCode:       nil,
			IsActivated:            true,
			IsVerified:             false,
			IsAdmin:                false,
			FollowerCount:          1820,
			Gender:                 "Female",
			DateOfBirth:            dob8,
			SecurityQuestion:       "What's your favorite food?",
			SecurityAnswer:         "Pizza",
			NewsletterSubscription: false,
			JoinedAt:               now.Add(-90 * 24 * time.Hour),
			CreatedAt:              now.Add(-90 * 24 * time.Hour),
			UpdatedAt:              now,
			Bio:                    "Food blogger and amateur chef sharing recipes and restaurant reviews",
			ProfilePictureURL:      "https://randomuser.me/api/portraits/women/3.jpg",
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

func (s *UserSeeder) SeedFollows() error {
	var count int64
	s.db.Table("follows").Count(&count)
	if count > 0 {
		log.Println("Follows already exist, skipping seeding")
		return nil
	}

	type Follow struct {
		ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
		FollowerID uuid.UUID `gorm:"type:uuid;not null;index"`
		FollowedID uuid.UUID `gorm:"type:uuid;not null;index"`
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}

	adminID := uuid.MustParse(getEnv("ADMIN_UUID", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"))
	johnID := uuid.MustParse(getEnv("JOHN_UUID", "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12"))
	janeID := uuid.MustParse(getEnv("JANE_UUID", "c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13"))
	samID := uuid.MustParse(getEnv("SAM_UUID", "d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14"))

	var users []struct {
		ID       uuid.UUID
		Username string
	}
	s.db.Table("users").Select("id, username").Where("username IN ?", []string{"techguru", "fitnesscoach", "travelbug", "foodie123"}).Find(&users)

	userMap := make(map[string]uuid.UUID)
	for _, user := range users {
		userMap[user.Username] = user.ID
	}

	follows := []Follow{
		{ID: uuid.New(), FollowerID: adminID, FollowedID: johnID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), FollowerID: adminID, FollowedID: janeID, CreatedAt: time.Now(), UpdatedAt: time.Now()},

		{ID: uuid.New(), FollowerID: johnID, FollowedID: adminID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), FollowerID: johnID, FollowedID: janeID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), FollowerID: johnID, FollowedID: samID, CreatedAt: time.Now(), UpdatedAt: time.Now()},

		{ID: uuid.New(), FollowerID: janeID, FollowedID: adminID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), FollowerID: janeID, FollowedID: johnID, CreatedAt: time.Now(), UpdatedAt: time.Now()},

		{ID: uuid.New(), FollowerID: samID, FollowedID: adminID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), FollowerID: samID, FollowedID: johnID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), FollowerID: samID, FollowedID: janeID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	if techGuruID, ok := userMap["techguru"]; ok {
		follows = append(follows,
			Follow{ID: uuid.New(), FollowerID: techGuruID, FollowedID: adminID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			Follow{ID: uuid.New(), FollowerID: johnID, FollowedID: techGuruID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			Follow{ID: uuid.New(), FollowerID: janeID, FollowedID: techGuruID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		)
	}

	if fitnessCoachID, ok := userMap["fitnesscoach"]; ok {
		follows = append(follows,
			Follow{ID: uuid.New(), FollowerID: fitnessCoachID, FollowedID: adminID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			Follow{ID: uuid.New(), FollowerID: johnID, FollowedID: fitnessCoachID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			Follow{ID: uuid.New(), FollowerID: janeID, FollowedID: fitnessCoachID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		)
	}

	if err := s.db.Table("follows").Create(&follows).Error; err != nil {
		return fmt.Errorf("failed to create follows: %w", err)
	}

	log.Printf("Created %d follow relationships", len(follows))
	return nil
}

func (s *UserSeeder) SeedAll() error {
	if err := s.SeedUsers(); err != nil {
		return err
	}
	if err := s.SeedSessions(); err != nil {
		return err
	}
	if err := s.SeedFollows(); err != nil {
		return err
	}
	return nil
}
