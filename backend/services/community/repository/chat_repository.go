package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/community/model"
)

// GormChatRepository is the implementation of ChatRepository using GORM
type GormChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) model.ChatRepository {
	return &GormChatRepository{db: db}
}

func (r *GormChatRepository) CreateChat(chat *model.ChatDTO) error {
	dbChat := &model.Chat{
		ChatID:    uuid.MustParse(chat.ID),
		Name:      chat.Name,
		IsGroup:   chat.IsGroupChat,
		CreatedBy: uuid.MustParse(chat.CreatorID),
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}
	return r.db.Create(dbChat).Error
}

func (r *GormChatRepository) FindChatByID(id string) (*model.ChatDTO, error) {
	var dbChat model.Chat

	// Parse the ID to UUID
	chatID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	err = r.db.First(&dbChat, "chat_id = ?", chatID).Error
	if err != nil {
		return nil, err
	}

	chat := &model.ChatDTO{
		ID:          dbChat.ChatID.String(),
		Name:        dbChat.Name,
		IsGroupChat: dbChat.IsGroup,
		CreatorID:   dbChat.CreatedBy.String(),
		CreatedAt:   dbChat.CreatedAt,
		UpdatedAt:   dbChat.UpdatedAt,
	}

	return chat, nil
}

func (r *GormChatRepository) ListChatsByUserID(userID string, limit, offset int) ([]*model.ChatDTO, error) {
	var dbChats []model.Chat

	// Parse the user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	err = r.db.Table("chats c").
		Joins("JOIN chat_participants cp ON c.chat_id = cp.chat_id").
		Where("cp.user_id = ?", userUUID).
		Limit(limit).
		Offset(offset).
		Find(&dbChats).Error
	if err != nil {
		return nil, err
	}

	chats := make([]*model.ChatDTO, len(dbChats))
	for i, dbChat := range dbChats {
		chats[i] = &model.ChatDTO{
			ID:          dbChat.ChatID.String(),
			Name:        dbChat.Name,
			IsGroupChat: dbChat.IsGroup,
			CreatorID:   dbChat.CreatedBy.String(),
			CreatedAt:   dbChat.CreatedAt,
			UpdatedAt:   dbChat.UpdatedAt,
		}
	}

	return chats, nil
}

func (r *GormChatRepository) UpdateChat(chat *model.ChatDTO) error {
	dbChat := &model.Chat{
		ChatID:    uuid.MustParse(chat.ID),
		Name:      chat.Name,
		IsGroup:   chat.IsGroupChat,
		UpdatedAt: chat.UpdatedAt,
	}
	return r.db.Model(&model.Chat{}).Where("chat_id = ?", dbChat.ChatID).Updates(dbChat).Error
}

func (r *GormChatRepository) DeleteChat(chatID string) error {
	// Parse the chat ID to UUID
	id, err := uuid.Parse(chatID)
	if err != nil {
		return err
	}

	return r.db.Delete(&model.Chat{}, "chat_id = ?", id).Error
}
