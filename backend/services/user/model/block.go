package model

import (
	"time"

	"github.com/google/uuid"
)

type UserBlock struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	BlockerID uuid.UUID `gorm:"type:uuid;not null;index"` // User who is blocking
	BlockedID uuid.UUID `gorm:"type:uuid;not null;index"` // User who is being blocked
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// BeforeCreate sets the ID for the block record
func (b *UserBlock) BeforeCreate() error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for the UserBlock model
func (UserBlock) TableName() string {
	return "user_blocks"
}

type UserReport struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ReporterID  uuid.UUID  `gorm:"type:uuid;not null;index"`         // User who is reporting
	ReportedID  uuid.UUID  `gorm:"type:uuid;not null;index"`         // User who is being reported
	Reason      string     `gorm:"type:text;not null"`               // Reason for the report
	Status      string     `gorm:"type:varchar(20);default:pending"` // pending, approved, rejected
	AdminNotes  string     `gorm:"type:text"`                        // Admin notes for the report
	ProcessedBy *uuid.UUID `gorm:"type:uuid"`                        // Admin who processed the report
	ProcessedAt *time.Time `gorm:"type:timestamp"`                   // When the report was processed
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}

// BeforeCreate sets the ID for the report record
func (r *UserReport) BeforeCreate() error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for the UserReport model
func (UserReport) TableName() string {
	return "user_reports"
}
