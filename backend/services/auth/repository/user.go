package repository

import (
	"time"
)

// User represents a user in the authentication system
type User struct {
	ID                    string    `json:"id" gorm:"primaryKey"`
	Email                 string    `json:"email" gorm:"uniqueIndex"`
	Name                  string    `json:"name"`
	Username              string    `json:"username" gorm:"uniqueIndex"`
	HashedPassword        string    `json:"-"` // Not exposed in JSON
	VerificationCode      string    `json:"-"` // Not exposed in JSON
	IsVerified            bool      `json:"is_verified" gorm:"default:false"`
	Gender                string    `json:"gender"`
	DateOfBirth           string    `json:"date_of_birth"`
	ProfilePicture        string    `json:"profile_picture"`
	Banner                string    `json:"banner"`
	SecurityQuestion      string    `json:"security_question"`
	SecurityAnswer        string    `json:"security_answer"`
	SubscribeToNewsletter bool      `json:"subscribe_to_newsletter"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// Token represents a refresh token in the authentication system
type Token struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"index"`
	RefreshToken string    `json:"-"` // Not exposed in JSON
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OAuthConnection represents a connection to an OAuth provider
type OAuthConnection struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	UserID     string    `json:"user_id" gorm:"index"`
	Provider   string    `json:"provider"`
	ProviderID string    `json:"provider_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Tokens represents the tokens returned after authentication
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}
