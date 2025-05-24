package api

import (
	communityProto "aycom/backend/proto/community"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CommunityService interface{}

type ChatService interface {
	ListChats(userID string, limit, offset int) ([]*communityProto.Chat, error)
	CreateChat(name string, description string, creatorID string, isGroupChat bool, participantIDs []string) (*communityProto.Chat, error)
	AddParticipant(chatID, userID, addedBy string) error
	RemoveParticipant(chatID, userID, removedBy string) error
	ListParticipants(chatID string, limit, offset int) ([]*communityProto.ChatParticipant, error)
	SendMessage(chatID, userID, content string) (string, error)
	GetMessages(chatID string, limit, offset int) ([]*communityProto.Message, error)
	DeleteMessage(chatID, messageID, userID string) error
	UnsendMessage(chatID, messageID, userID string) error
	SearchMessages(chatID, query string, limit, offset int) ([]*communityProto.Message, error)
}

// Repository interfaces for membership checks
type CommunityMemberRepository interface {
	IsMember(communityID, userID uuid.UUID) (bool, error)
}

type CommunityJoinRequestRepository interface {
	HasPendingJoinRequest(communityID, userID uuid.UUID) (bool, error)
}

type CommunityHandler struct {
	communityProto.UnimplementedCommunityServiceServer
	communityService         CommunityService
	chatService              ChatService
	communityMemberRepo      CommunityMemberRepository
	communityJoinRequestRepo CommunityJoinRequestRepository
}

func NewCommunityHandler(
	communityService CommunityService,
	chatService ChatService,
	memberRepo CommunityMemberRepository,
	joinRequestRepo CommunityJoinRequestRepository,
) *CommunityHandler {
	return &CommunityHandler{
		communityService:         communityService,
		chatService:              chatService,
		communityMemberRepo:      memberRepo,
		communityJoinRequestRepo: joinRequestRepo,
	}
}

// Community management
func (h *CommunityHandler) CreateCommunity(ctx context.Context, req *communityProto.CreateCommunityRequest) (*communityProto.CommunityResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) UpdateCommunity(ctx context.Context, req *communityProto.UpdateCommunityRequest) (*communityProto.CommunityResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) ApproveCommunity(ctx context.Context, req *communityProto.ApproveCommunityRequest) (*communityProto.CommunityResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) DeleteCommunity(ctx context.Context, req *communityProto.DeleteCommunityRequest) (*communityProto.EmptyResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) GetCommunityByID(ctx context.Context, req *communityProto.GetCommunityByIDRequest) (*communityProto.CommunityResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) ListCommunities(ctx context.Context, req *communityProto.ListCommunitiesRequest) (*communityProto.ListCommunitiesResponse, error) {
	return nil, nil
}

// Member management
func (h *CommunityHandler) AddMember(ctx context.Context, req *communityProto.AddMemberRequest) (*communityProto.MemberResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) RemoveMember(ctx context.Context, req *communityProto.RemoveMemberRequest) (*communityProto.EmptyResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) ListMembers(ctx context.Context, req *communityProto.ListMembersRequest) (*communityProto.ListMembersResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) UpdateMemberRole(ctx context.Context, req *communityProto.UpdateMemberRoleRequest) (*communityProto.MemberResponse, error) {
	return nil, nil
}

// Community rules
func (h *CommunityHandler) AddRule(ctx context.Context, req *communityProto.AddRuleRequest) (*communityProto.RuleResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) RemoveRule(ctx context.Context, req *communityProto.RemoveRuleRequest) (*communityProto.EmptyResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) ListRules(ctx context.Context, req *communityProto.ListRulesRequest) (*communityProto.ListRulesResponse, error) {
	return nil, nil
}

// Join requests
func (h *CommunityHandler) RequestToJoin(ctx context.Context, req *communityProto.RequestToJoinRequest) (*communityProto.JoinRequestResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) ApproveJoinRequest(ctx context.Context, req *communityProto.ApproveJoinRequestRequest) (*communityProto.JoinRequestResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) RejectJoinRequest(ctx context.Context, req *communityProto.RejectJoinRequestRequest) (*communityProto.JoinRequestResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) ListJoinRequests(ctx context.Context, req *communityProto.ListJoinRequestsRequest) (*communityProto.ListJoinRequestsResponse, error) {
	return nil, nil
}

// Membership checks
func (h *CommunityHandler) IsMember(ctx context.Context, req *communityProto.IsMemberRequest) (*communityProto.IsMemberResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.CommunityId == "" {
		return nil, status.Error(codes.InvalidArgument, "community_id is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Convert string IDs to UUID
	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID format")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	// Check if user is a member
	isMember, err := h.communityMemberRepo.IsMember(communityID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check membership: %v", err))
	}

	return &communityProto.IsMemberResponse{
		IsMember: isMember,
	}, nil
}

func (h *CommunityHandler) HasPendingJoinRequest(ctx context.Context, req *communityProto.HasPendingJoinRequestRequest) (*communityProto.HasPendingJoinRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.CommunityId == "" {
		return nil, status.Error(codes.InvalidArgument, "community_id is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Convert string IDs to UUID
	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID format")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	// Check if user has a pending join request
	hasRequest, err := h.communityJoinRequestRepo.HasPendingJoinRequest(communityID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check join request: %v", err))
	}

	return &communityProto.HasPendingJoinRequestResponse{
		HasRequest: hasRequest,
	}, nil
}

// Chat
func (h *CommunityHandler) CreateChat(ctx context.Context, req *communityProto.CreateChatRequest) (*communityProto.ChatResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validation
	if req.Name == "" && req.IsGroup {
		return nil, status.Error(codes.InvalidArgument, "group chat requires a name")
	}

	if req.CreatedBy == "" {
		return nil, status.Error(codes.InvalidArgument, "created_by is required")
	}

	if len(req.ParticipantIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one participant is required")
	}

	// Create chat using service
	chat, err := h.chatService.CreateChat(
		req.Name,           // name
		"",                 // description (not in proto)
		req.CreatedBy,      // creatorID
		req.IsGroup,        // isGroupChat
		req.ParticipantIds, // participantIDs
	)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create chat: %v", err))
	}

	return &communityProto.ChatResponse{
		Chat: chat,
	}, nil
}
func (h *CommunityHandler) AddChatParticipant(ctx context.Context, req *communityProto.AddChatParticipantRequest) (*communityProto.ChatParticipantResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) RemoveChatParticipant(ctx context.Context, req *communityProto.RemoveChatParticipantRequest) (*communityProto.EmptyResponse, error) {
	return nil, nil
}
func (h *CommunityHandler) ListChats(ctx context.Context, req *communityProto.ListChatsRequest) (*communityProto.ListChatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Default limit and offset if not provided
	limit := 50
	offset := 0

	chats, err := h.chatService.ListChats(req.UserId, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list chats: %v", err))
	}

	return &communityProto.ListChatsResponse{
		Chats: chats,
	}, nil
}
func (h *CommunityHandler) ListChatParticipants(ctx context.Context, req *communityProto.ListChatParticipantsRequest) (*communityProto.ListChatParticipantsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.ChatId == "" {
		return nil, status.Error(codes.InvalidArgument, "chat_id is required")
	}

	// Default limit and offset
	limit := 50
	offset := 0

	// Get participants using service
	participants, err := h.chatService.ListParticipants(req.ChatId, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list participants: %v", err))
	}

	return &communityProto.ListChatParticipantsResponse{
		Participants: participants,
	}, nil
}

// Messages
func (h *CommunityHandler) SendMessage(ctx context.Context, req *communityProto.SendMessageRequest) (*communityProto.MessageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validation
	if req.ChatId == "" {
		return nil, status.Error(codes.InvalidArgument, "chat_id is required")
	}

	if req.SenderId == "" {
		return nil, status.Error(codes.InvalidArgument, "sender_id is required")
	}

	// No content and no media is invalid
	if req.Content == "" && req.MediaUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "either content or media_url is required")
	}

	// Send message using service
	messageID, err := h.chatService.SendMessage(req.ChatId, req.SenderId, req.Content)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to send message: %v", err))
	}

	// Create message response
	now := time.Now()
	message := &communityProto.Message{
		Id:        messageID,
		ChatId:    req.ChatId,
		SenderId:  req.SenderId,
		Content:   req.Content,
		MediaUrl:  req.MediaUrl,
		MediaType: req.MediaType,
		SentAt:    timestamppb.New(now),
	}

	return &communityProto.MessageResponse{
		Message: message,
	}, nil
}
func (h *CommunityHandler) DeleteMessage(ctx context.Context, req *communityProto.DeleteMessageRequest) (*communityProto.EmptyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.MessageId == "" {
		return nil, status.Error(codes.InvalidArgument, "message_id is required")
	}

	// NOTE: Since the proto DeleteMessageRequest only contains messageId, we need to:
	// 1. Extract userID from the context (usually from authentication)
	// 2. Look up the chatID that this message belongs to

	// For now, we'll use a simplified approach since proper auth context extraction
	// would depend on how authentication is implemented in the system

	// Extract userID from context (this is a placeholder - implement based on your auth system)
	userID, err := extractUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to authenticate user")
	}

	// In a real implementation, we would look up the chatID for this message
	// For now, we'll use a temporary workaround by passing empty string
	// and rely on the service implementation to look up the chat from the message
	chatID := ""

	err = h.chatService.DeleteMessage(chatID, req.MessageId, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete message: %v", err))
	}

	return &communityProto.EmptyResponse{}, nil
}
func (h *CommunityHandler) UnsendMessage(ctx context.Context, req *communityProto.UnsendMessageRequest) (*communityProto.EmptyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.MessageId == "" {
		return nil, status.Error(codes.InvalidArgument, "message_id is required")
	}

	// Extract userID from context (this is a placeholder - implement based on your auth system)
	userID, err := extractUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to authenticate user")
	}

	// In a real implementation, we would look up the chatID for this message
	// For now, we'll use a temporary workaround by passing empty string
	chatID := ""

	err = h.chatService.UnsendMessage(chatID, req.MessageId, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to unsend message: %v", err))
	}

	return &communityProto.EmptyResponse{}, nil
}
func (h *CommunityHandler) ListMessages(ctx context.Context, req *communityProto.ListMessagesRequest) (*communityProto.ListMessagesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.ChatId == "" {
		return nil, status.Error(codes.InvalidArgument, "chat_id is required")
	}

	// Default limit and offset if not provided
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50 // Default limit
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0 // Default offset
	}

	// Get messages using service
	messages, err := h.chatService.GetMessages(req.ChatId, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get messages: %v", err))
	}

	return &communityProto.ListMessagesResponse{
		Messages: messages,
	}, nil
}
func (h *CommunityHandler) SearchMessages(ctx context.Context, req *communityProto.SearchMessagesRequest) (*communityProto.ListMessagesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.ChatId == "" {
		return nil, status.Error(codes.InvalidArgument, "chat_id is required")
	}

	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	// Default limit and offset if not provided
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50 // Default limit
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0 // Default offset
	}

	// Search messages using service
	messages, err := h.chatService.SearchMessages(req.ChatId, req.Query, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to search messages: %v", err))
	}

	return &communityProto.ListMessagesResponse{
		Messages: messages,
	}, nil
}

// Helper function to extract user ID from context
// This is a placeholder - replace with your actual auth implementation
func extractUserIDFromContext(ctx context.Context) (string, error) {
	// In a real implementation, this would extract the authenticated user ID
	// from the context, which is typically set by an authentication middleware

	// For development/temporary use, return a fixed userID
	// TODO: Replace with proper authentication mechanism
	return "system-user", nil
}
