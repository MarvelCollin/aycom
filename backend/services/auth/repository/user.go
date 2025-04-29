package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the authentication system
type User struct {
	gorm.Model
	ID                    uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name                  string    `gorm:"size:255;not null" json:"name"`
	Username              string    `gorm:"size:100;uniqueIndex;not null" json:"username"`
	Email                 string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash          string    `gorm:"size:255;not null" json:"-"` // Exclude from JSON responses
	Gender                string    `gorm:"size:50" json:"gender"`
	DateOfBirth           string    `gorm:"type:date" json:"dateOfBirth"`
	SecurityQuestion      string    `gorm:"size:255" json:"securityQuestion"`
	SecurityAnswer        string    `gorm:"size:255" json:"securityAnswer"`
	EmailVerified         bool      `gorm:"default:false" json:"emailVerified"`
	VerificationCode      string    `gorm:"size:10" json:"-"`      // Exclude from JSON responses
	VerificationExpiresAt time.Time `json:"verificationExpiresAt"` // Exclude from JSON responses
	SubscribeToNewsletter bool      `gorm:"default:false" json:"subscribeToNewsletter"`
	ProfilePictureURL     string    `gorm:"type:text" json:"profilePictureUrl"` // Added field
	BannerURL             string    `gorm:"type:text" json:"bannerUrl"`         // Added field
	CreatedAt             time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt             time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
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
