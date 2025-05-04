package service

type ChatService struct{}

func NewChatService(chatRepo, chatParticipantRepo, messageRepo, deletedChatRepo interface{}) *ChatService {
	return &ChatService{}
}
