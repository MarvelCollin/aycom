package model

import "time"

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

type MessageDTO struct {
	ID               string
	ChatID           string
	SenderID         string
	Content          string
	MediaURL         string
	MediaType        string
	Timestamp        time.Time
	Unsent           bool
	UnsentAt         *time.Time
	DeletedForSender bool
	DeletedForAll    bool
	ReplyToMessageID string
	IsRead           bool
	IsEdited         bool
	IsDeleted        bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ParticipantDTO struct {
	ChatID   string
	UserID   string
	JoinedAt time.Time
	IsAdmin  bool
}

type DeletedChatDTO struct {
	ChatID    string
	UserID    string
	DeletedAt time.Time
}

type ChatRepository interface {
	CreateChat(chat *ChatDTO) error
	FindChatByID(chatID string) (*ChatDTO, error)
	ListChatsByUserID(userID string, limit, offset int) ([]*ChatDTO, error)
	UpdateChat(chat *ChatDTO) error
	DeleteChat(chatID string) error
}

type MessageRepository interface {
	SaveMessage(message *MessageDTO) error
	FindMessageByID(messageID string) (*MessageDTO, error)
	FindMessagesByChatID(chatID string, limit, offset int) ([]*MessageDTO, error)
	MarkMessageAsRead(messageID, userID string) error
	DeleteMessage(messageID string) error
	UnsendMessage(messageID string) error
	SearchMessages(chatID, query string, limit, offset int) ([]*MessageDTO, error)
	UpdateMessage(message *MessageDTO) error
}

type ParticipantRepository interface {
	AddParticipant(participant *ParticipantDTO) error
	RemoveParticipant(chatID, userID string) error
	ListParticipantsByChatID(chatID string, limit, offset int) ([]*ParticipantDTO, error)
	IsUserInChat(chatID, userID string) (bool, error)
}

type DeletedChatRepository interface {
	MarkChatAsDeleted(chatID, userID string) error
	IsDeletedForUser(chatID, userID string) (bool, error)
}
