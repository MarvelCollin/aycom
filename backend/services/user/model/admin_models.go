package model

import (
	"time"

	"github.com/google/uuid"
)

// CommunityRequest represents a request to create a new community
type CommunityRequest struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Name        string    `gorm:"type:text;not null"`
	Description string    `gorm:"type:text"`
	CategoryID  uuid.UUID `gorm:"type:uuid;index"`
	Status      string    `gorm:"type:varchar(20);not null;default:'pending'"` // pending, approved, rejected
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PremiumRequest represents a request from a user to become premium
type PremiumRequest struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Reason    string    `gorm:"type:text"`
	Status    string    `gorm:"type:varchar(20);not null;default:'pending'"` // pending, approved, rejected
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ReportRequest represents a report made by one user against another
type ReportRequest struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	ReporterID     uuid.UUID `gorm:"type:uuid;not null;index"`
	ReportedUserID uuid.UUID `gorm:"type:uuid;not null;index"`
	Reason         string    `gorm:"type:text;not null"`
	Status         string    `gorm:"type:varchar(20);not null;default:'pending'"` // pending, approved, rejected
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ThreadCategory represents a category for threads
type ThreadCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null;unique"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CommunityCategory represents a category for communities
type CommunityCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null;unique"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Newsletter represents a newsletter sent to subscribed users
type Newsletter struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Subject   string    `gorm:"type:varchar(255);not null"`
	Content   string    `gorm:"type:text;not null"`
	SentBy    uuid.UUID `gorm:"type:uuid;not null;index"`
	SentAt    time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BeforeCreate hooks to generate UUIDs for new records
func (cr *CommunityRequest) BeforeCreate() error {
	if cr.ID == uuid.Nil {
		cr.ID = uuid.New()
	}
	return nil
}

func (pr *PremiumRequest) BeforeCreate() error {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	return nil
}

func (rr *ReportRequest) BeforeCreate() error {
	if rr.ID == uuid.Nil {
		rr.ID = uuid.New()
	}
	return nil
}

func (tc *ThreadCategory) BeforeCreate() error {
	if tc.ID == uuid.Nil {
		tc.ID = uuid.New()
	}
	return nil
}

func (cc *CommunityCategory) BeforeCreate() error {
	if cc.ID == uuid.Nil {
		cc.ID = uuid.New()
	}
	return nil
}

func (n *Newsletter) BeforeCreate() error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}
