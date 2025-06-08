package model

import (
	"time"

	"github.com/google/uuid"
)

type CommunityRequest struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Name        string    `gorm:"type:text;not null"`
	Description string    `gorm:"type:text"`
	CategoryID  uuid.UUID `gorm:"type:uuid;index"`
	Status      string    `gorm:"type:varchar(20);not null;default:'pending'"`
	LogoURL     string    `gorm:"type:text"`
	BannerURL   string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PremiumRequest struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID             uuid.UUID `gorm:"type:uuid;not null;index"`
	Reason             string    `gorm:"type:text"`
	IdentityCardNumber string    `gorm:"type:text"`
	FacePhotoURL       string    `gorm:"type:text"`
	Status             string    `gorm:"type:varchar(20);not null;default:'pending'"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ReportRequest struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	ReporterID     uuid.UUID `gorm:"type:uuid;not null;index"`
	ReportedUserID uuid.UUID `gorm:"type:uuid;not null;index"`
	Reason         string    `gorm:"type:text;not null"`
	Status         string    `gorm:"type:varchar(20);not null;default:'pending'"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ThreadCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null;unique"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CommunityCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null;unique"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Newsletter struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Subject   string    `gorm:"type:varchar(255);not null"`
	Content   string    `gorm:"type:text;not null"`
	SentBy    uuid.UUID `gorm:"type:uuid;not null;index"`
	SentAt    time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

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
