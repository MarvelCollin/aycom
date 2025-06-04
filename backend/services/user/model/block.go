package model

import (
	"time"

	"github.com/google/uuid"
)

type UserBlock struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	BlockerID uuid.UUID `gorm:"type:uuid;not null;index"` 
	BlockedID uuid.UUID `gorm:"type:uuid;not null;index"` 
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (b *UserBlock) BeforeCreate() error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

func (UserBlock) TableName() string {
	return "user_blocks"
}

type UserReport struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ReporterID  uuid.UUID  `gorm:"type:uuid;not null;index"`         
	ReportedID  uuid.UUID  `gorm:"type:uuid;not null;index"`         
	Reason      string     `gorm:"type:text;not null"`               
	Status      string     `gorm:"type:varchar(20);default:pending"` 
	AdminNotes  string     `gorm:"type:text"`                        
	ProcessedBy *uuid.UUID `gorm:"type:uuid"`                        
	ProcessedAt *time.Time `gorm:"type:timestamp"`                   
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}

func (r *UserReport) BeforeCreate() error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (UserReport) TableName() string {
	return "user_reports"
}