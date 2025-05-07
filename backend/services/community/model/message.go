package model

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	MessageID        uuid.UUID  `gorm:"type:uuid;primaryKey;column:message_id;default:gen_random_uuid()"`
	ChatID           uuid.UUID  `gorm:"type:uuid;not null;column:chat_id"`
	SenderID         uuid.UUID  `gorm:"type:uuid;not null;column:sender_id"`
	Content          string     `gorm:"type:text;column:content"`
	MediaURL         string     `gorm:"type:varchar(512);column:media_url"`
	MediaType        string     `gorm:"type:varchar(10);column:media_type"`
	SentAt           time.Time  `gorm:"column:sent_at;default:now();not null"`
	Unsent           bool       `gorm:"default:false;not null;column:unsent"`
	UnsentAt         *time.Time `gorm:"column:unsent_at"`
	DeletedForSender bool       `gorm:"default:false;not null;column:deleted_for_sender"`
	DeletedForAll    bool       `gorm:"default:false;not null;column:deleted_for_all"`
	ReplyToMessageID *uuid.UUID `gorm:"type:uuid;column:reply_to_message_id"`
	IsRead           bool       `gorm:"default:false;column:is_read"`
	IsEdited         bool       `gorm:"default:false;column:is_edited"`
	IsDeleted        bool       `gorm:"default:false;column:is_deleted"`
	CreatedAt        time.Time  `gorm:"column:created_at;default:now()"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;default:now()"`
	DeletedAt        *time.Time `gorm:"index;column:deleted_at"`
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
