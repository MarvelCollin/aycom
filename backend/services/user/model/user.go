package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                    string    `json:"id" gorm:"primaryKey"`
	Username              string    `json:"username" gorm:"uniqueIndex"`
	Email                 string    `json:"email" gorm:"uniqueIndex"`
	Name                  string    `json:"name"`
	Gender                string    `json:"gender"`
	DateOfBirth           time.Time `json:"date_of_birth"`
	ProfilePicture        string    `json:"profile_picture"`
	Banner                string    `json:"banner"`
	Bio                   string    `json:"bio"`
	Followers             int       `json:"followers"`
	Following             int       `json:"following"`
	IsVerified            bool      `json:"is_verified" gorm:"default:false"`
	SecurityQuestion      string    `json:"security_question"`
	SecurityAnswer        string    `json:"security_answer"`
	SubscribeToNewsletter bool      `json:"subscribe_to_newsletter"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// NewUser creates a new user with the given details
func NewUser(id, username, email, name, gender, dateOfBirth, profilePicture, banner, secQuestion, secAnswer string, subscribeToNewsletter bool) *User {
	dob, _ := time.Parse("2006-01-02", dateOfBirth) // Ignore error for simplicity

	// Generate UUID if not provided
	if id == "" {
		id = uuid.New().String()
	}

	return &User{
		ID:                    id,
		Username:              username,
		Email:                 email,
		Name:                  name,
		Gender:                gender,
		DateOfBirth:           dob,
		ProfilePicture:        profilePicture,
		Banner:                banner,
		IsVerified:            false,
		SecurityQuestion:      secQuestion,
		SecurityAnswer:        secAnswer,
		SubscribeToNewsletter: subscribeToNewsletter,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
}
