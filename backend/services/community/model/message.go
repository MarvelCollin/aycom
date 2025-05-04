package model

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	MessageID        uuid.UUID `gorm:"type:uuid;primaryKey;column:message_id"`
	ChatID           uuid.UUID `gorm:"type:uuid;not null"`
	SenderID         uuid.UUID `gorm:"type:uuid;not null"`
	Content          string    `gorm:"type:text"`
	MediaURL         string    `gorm:"type:varchar(512)"`
	MediaType        string    `gorm:"type:varchar(10)"`
	SentAt           time.Time `gorm:"autoCreateTime"`
	Unsent           bool      `gorm:"default:false;not null"`
	UnsentAt         *time.Time
	DeletedForSender bool       `gorm:"default:false;not null"`
	DeletedForAll    bool       `gorm:"default:false;not null"`
	ReplyToMessageID *uuid.UUID `gorm:"type:uuid"`
	DeletedAt        *time.Time `gorm:"index"`
	IsRead           bool       `gorm:"column:is_read"`
	IsEdited         bool       `gorm:"column:is_edited"`
	IsDeleted        bool       `gorm:"column:is_deleted"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

type ReadReceipt struct {
	MessageID uuid.UUID `gorm:"column:message_id;primaryKey"`
	UserID    uuid.UUID `gorm:"column:user_id;primaryKey"`
	ReadAt    time.Time `gorm:"column:read_at"`
}

func (m *Message) TableName() string {
	return "messages"
}

func (r *ReadReceipt) TableName() string {
	return "read_receipts"
}
