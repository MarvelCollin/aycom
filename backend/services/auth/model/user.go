package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the authentication system
type User struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name                  string     `gorm:"size:255;not null" json:"name"`
	Username              string     `gorm:"size:255;not null;unique" json:"username"`
	Email                 string     `gorm:"size:255;not null;unique" json:"email"`
	PasswordHash          string     `gorm:"size:255;not null" json:"-"`
	PasswordSalt          string     `gorm:"size:64;not null" json:"-"`
	Gender                string     `gorm:"size:10" json:"gender"`
	DateOfBirth           time.Time  `json:"date_of_birth"`
	SecurityQuestion      string     `gorm:"size:255" json:"security_question"`
	SecurityAnswer        string     `gorm:"size:255" json:"-"`
	GoogleID              string     `gorm:"size:255;unique" json:"google_id,omitempty"`
	IsActivated           bool       `gorm:"default:true" json:"is_activated"`
	IsBanned              bool       `gorm:"default:false" json:"is_banned"`
	IsDeactivated         bool       `gorm:"default:false" json:"is_deactivated"`
	IsAdmin               bool       `gorm:"default:false" json:"is_admin"`
	NewsletterSubscription bool       `gorm:"default:false" json:"newsletter_subscription"`
	LastLoginAt           *time.Time `json:"last_login_at"`
	JoinedAt              time.Time  `gorm:"default:CURRENT_TIMESTAMP;not null" json:"joined_at"`
	VerificationCode      *string    `gorm:"size:64" json:"-"`
	VerificationCodeExpiresAt *time.Time `json:"-"`
	CreatedAt             time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Sessions              []Session  `gorm:"foreignKey:UserID" json:"-"`
}

// Session represents a user session
type Session struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	AccessToken  string    `gorm:"size:500;not null" json:"-"`
	RefreshToken string    `gorm:"size:500;not null;unique" json:"-"`
	IPAddress    string    `gorm:"size:45" json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	User         User      `gorm:"foreignKey:UserID" json:"-"`
}
