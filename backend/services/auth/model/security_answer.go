package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserSecurityAnswer stores a user's answer to a security question
type UserSecurityAnswer struct {
	AnswerID   uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`   // Maps to UserAuth.UserID
	QuestionID uuid.UUID `gorm:"type:uuid;not null"`         // Maps to SecurityQuestion.QuestionID
	AnswerHash string    `gorm:"type:varchar(255);not null"` // Store hashed answer for security
	AnswerSalt string    `gorm:"type:varchar(64);not null"`  // Salt used for answer hashing
	CreatedAt  time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
	UpdatedAt  time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
}

// BeforeCreate will set default values before inserting a new record
func (usa *UserSecurityAnswer) BeforeCreate(tx *gorm.DB) error {
	if usa.AnswerID == uuid.Nil {
		usa.AnswerID = uuid.New()
	}
	return nil
}
