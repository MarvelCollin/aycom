package repository

import (
	"time"
)


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


type DeletedChatModel struct {
	ChatID    string
	UserID    string
	DeletedAt time.Time
}
