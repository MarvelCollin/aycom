package model

import "time"

// Chat represents a chat room in the domain model
type ChatDTO struct {
	ID          string
	Name        string
	Description string
	CreatorID   string
	CommunityID string
	IsGroupChat bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Message represents a chat message in the domain model
type MessageDTO struct {
	ID        string
	ChatID    string
	SenderID  string
	Content   string
	Timestamp time.Time
	IsRead    bool
	IsEdited  bool
	IsDeleted bool
}

// Participant represents a chat participant in the domain model
type ParticipantDTO struct {
	ChatID   string
	UserID   string
	JoinedAt time.Time
	IsAdmin  bool
}

// DeletedChat represents a chat deleted for a specific user
type DeletedChatDTO struct {
	ChatID    string
	UserID    string
	DeletedAt time.Time
}

// ChatRepository defines the interface for chat operations
type ChatRepository interface {
	CreateChat(chat *ChatDTO) error
	FindChatByID(chatID string) (*ChatDTO, error)
	ListChatsByUserID(userID string, limit, offset int) ([]*ChatDTO, error)
	UpdateChat(chat *ChatDTO) error
	DeleteChat(chatID string) error
}

// MessageRepository defines the interface for message operations
type MessageRepository interface {
	SaveMessage(message *MessageDTO) error
	FindMessageByID(messageID string) (*MessageDTO, error)
	FindMessagesByChatID(chatID string, limit, offset int) ([]*MessageDTO, error)
	MarkMessageAsRead(messageID, userID string) error
	DeleteMessage(messageID string) error
	UnsendMessage(messageID string) error
	SearchMessages(chatID, query string, limit, offset int) ([]*MessageDTO, error)
}

// ParticipantRepository defines the interface for participant operations
type ParticipantRepository interface {
	AddParticipant(participant *ParticipantDTO) error
	RemoveParticipant(chatID, userID string) error
	ListParticipantsByChatID(chatID string, limit, offset int) ([]*ParticipantDTO, error)
	IsUserInChat(chatID, userID string) (bool, error)
}

// DeletedChatRepository defines the interface for deleted chat operations
type DeletedChatRepository interface {
	MarkChatAsDeleted(chatID, userID string) error
	IsDeletedForUser(chatID, userID string) (bool, error)
}
