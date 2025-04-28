package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// JSONB type for handling JSON data
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value")
	}

	result := make(map[string]interface{})
	err := json.Unmarshal(bytes, &result)
	*j = JSONB(result)
	return err
}

// StringArray type for handling string arrays
type StringArray []string

// Value implements the driver.Valuer interface for StringArray
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for StringArray
func (s *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal StringArray value")
	}

	var result []string
	err := json.Unmarshal(bytes, &result)
	*s = StringArray(result)
	return err
}

// UserProfile represents a user's profile information
type UserProfile struct {
	ID                      uuid.UUID   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID                  uuid.UUID   `gorm:"type:uuid;not null;unique" json:"user_id"`
	Bio                     string      `json:"bio"`
	ProfilePictureURL       string      `gorm:"size:500" json:"profile_picture_url"`
	BannerURL               string      `gorm:"size:500" json:"banner_url"`
	Location                string      `gorm:"size:255" json:"location"`
	Website                 string      `gorm:"size:255" json:"website"`
	SocialLinks             JSONB       `gorm:"type:jsonb" json:"social_links"`
	Interests               StringArray `gorm:"type:text[]" json:"interests"`
	Language                string      `gorm:"size:10;default:'en'" json:"language"`
	Theme                   string      `gorm:"size:20;default:'light'" json:"theme"`
	IsPrivate               bool        `gorm:"default:false" json:"is_private"`
	IsPremium               bool        `gorm:"default:false" json:"is_premium"`
	NotificationPreferences JSONB       `gorm:"type:jsonb" json:"notification_preferences"`
	CreatedAt               time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt               time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Contacts                []Contact   `gorm:"foreignKey:UserID" json:"-"`
}

// Contact represents a connection between users
type Contact struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ContactUserID uuid.UUID `gorm:"type:uuid;not null" json:"contact_user_id"`
	Relationship  string    `gorm:"size:50" json:"relationship"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}
