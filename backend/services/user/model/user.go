package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	// Removed gorm.Model to use explicit UUID
	ID                    uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"user_id"`
	Username              string         `gorm:"size:100;uniqueIndex;not null" json:"username"`
	Email                 string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Name                  string         `gorm:"size:255;not null" json:"name"`
	Gender                string         `gorm:"size:50" json:"gender"`
	DateOfBirth           *time.Time     `gorm:"type:date" json:"date_of_birth"`       // Use pointer for potential null
	ProfilePictureURL     string         `gorm:"type:text" json:"profile_picture_url"` // Renamed field
	BannerURL             string         `gorm:"type:text" json:"banner_url"`          // Renamed field
	Bio                   string         `gorm:"type:text" json:"bio"`
	Followers             int            `gorm:"default:0" json:"followers"`
	Following             int            `gorm:"default:0" json:"following"`
	IsVerified            bool           `gorm:"default:false" json:"is_verified"`
	SecurityQuestion      string         `gorm:"size:255" json:"security_question"` // Should this be stored here or only in auth?
	SecurityAnswer        string         `gorm:"size:255" json:"security_answer"`   // Should this be stored here or only in auth?
	SubscribeToNewsletter bool           `gorm:"default:false" json:"subscribe_to_newsletter"`
	CreatedAt             time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"` // Add soft delete
}

// NewUser creates a new user with the given details
// Note: This constructor might need adjustment based on how DateOfBirth string is handled
func NewUser(id uuid.UUID, username, email, name, gender, dateOfBirthStr, profilePictureURL, bannerURL, secQuestion, secAnswer string, subscribeToNewsletter bool) *User {
	var dobPtr *time.Time
	if dob, err := time.Parse("2006-01-02", dateOfBirthStr); err == nil {
		dobPtr = &dob
	}

	return &User{
		ID:                    id,
		Username:              username,
		Email:                 email,
		Name:                  name,
		Gender:                gender,
		DateOfBirth:           dobPtr,
		ProfilePictureURL:     profilePictureURL,
		BannerURL:             bannerURL,
		IsVerified:            false, // Typically verified via auth service interaction
		SecurityQuestion:      secQuestion,
		SecurityAnswer:        secAnswer,
		SubscribeToNewsletter: subscribeToNewsletter,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
}
