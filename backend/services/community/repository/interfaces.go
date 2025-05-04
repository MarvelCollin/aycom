package repository

import (
	"time"
)

// ChatModel represents a chat for repository layer
type ChatModel struct {
	ID          string
	Name        string
	Description string
	CreatorID   string
	CommunityID string
	IsGroupChat bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// MessageModel represents a chat message for repository layer
type MessageModel struct {
	ID        string
	ChatID    string
	SenderID  string
	Content   string
	Timestamp time.Time
	IsRead    bool
	IsEdited  bool
	IsDeleted bool
}

// DeletedChatModel represents a deleted chat for repository layer
type DeletedChatModel struct {
	ChatID    string
	UserID    string
	DeletedAt time.Time
}
