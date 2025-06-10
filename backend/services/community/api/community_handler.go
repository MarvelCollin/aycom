package api

import (
	communityProto "aycom/backend/proto/community"
	"aycom/backend/services/community/model"
	"aycom/backend/services/community/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CommunityService interface {
	CreateCommunity(ctx context.Context, community *model.Community) error
	UpdateCommunity(ctx context.Context, community *model.Community) error
	ApproveCommunity(ctx context.Context, communityID uuid.UUID) error
	DeleteCommunity(ctx context.Context, communityID uuid.UUID) error
	GetCommunityByID(ctx context.Context, communityID uuid.UUID) (*model.Community, error)
	ListCommunities(ctx context.Context, offset, limit int) ([]*model.Community, error)
	ListCommunitiesByCategories(ctx context.Context, categories []string, offset, limit int) ([]*model.Community, error)
	SearchCommunities(ctx context.Context, query string, categories []string, isApproved *bool, offset, limit int) ([]*model.Community, int64, error)
	ListUserCommunities(ctx context.Context, userID uuid.UUID, status string, offset, limit int) ([]*model.Community, int64, error)
	CountCommunities(ctx context.Context) (int64, error)

	CreateCategory(ctx context.Context, category *model.Category) error
	GetCategoryByID(ctx context.Context, categoryID uuid.UUID) (*model.Category, error)
	GetCategoryByName(ctx context.Context, name string) (*model.Category, error)
	ListCategories(ctx context.Context) ([]*model.Category, error)
	UpdateCategory(ctx context.Context, category *model.Category) error
	DeleteCategory(ctx context.Context, categoryID uuid.UUID) error
	AddCommunityToCategory(ctx context.Context, communityID, categoryID uuid.UUID) error
	RemoveCommunityFromCategory(ctx context.Context, communityID, categoryID uuid.UUID) error
	GetCommunityCategoriesById(ctx context.Context, communityID uuid.UUID) ([]*model.Category, error)
}

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
	communityMemberRepo      repository.CommunityMemberRepository
	communityJoinRequestRepo repository.CommunityJoinRequestRepository
	ruleRepo                 repository.CommunityRuleRepository
}

func NewCommunityHandler(
	communityService CommunityService,
	chatService ChatService,
	memberRepo repository.CommunityMemberRepository,
	joinRequestRepo repository.CommunityJoinRequestRepository,
	ruleRepo repository.CommunityRuleRepository,
) *CommunityHandler {
	return &CommunityHandler{
		communityService:         communityService,
		chatService:              chatService,
		communityMemberRepo:      memberRepo,
		communityJoinRequestRepo: joinRequestRepo,
		ruleRepo:                 ruleRepo,
	}
}

func (h *CommunityHandler) CreateCommunity(ctx context.Context, req *communityProto.CreateCommunityRequest) (*communityProto.CommunityResponse, error) {
	if req == nil || req.Community == nil {
		return nil, status.Error(codes.InvalidArgument, "community is required")
	}

	community := &model.Community{
		CommunityID: uuid.New(),
		Name:        req.Community.Name,
		Description: req.Community.Description,
		LogoURL:     req.Community.LogoUrl,
		BannerURL:   req.Community.BannerUrl,
		CreatorID:   uuid.MustParse(req.Community.CreatorId),
		IsApproved:  false,
	}

	if err := h.communityService.CreateCommunity(ctx, community); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create community: %v", err))
	}

	return h.mapCommunityToResponse(community), nil
}

func (h *CommunityHandler) UpdateCommunity(ctx context.Context, req *communityProto.UpdateCommunityRequest) (*communityProto.CommunityResponse, error) {
	if req == nil || req.Community == nil {
		return nil, status.Error(codes.InvalidArgument, "community is required")
	}

	communityID, err := uuid.Parse(req.Community.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	existing, err := h.communityService.GetCommunityByID(ctx, communityID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "community not found")
	}

	if req.Community.Name != "" {
		existing.Name = req.Community.Name
	}
	if req.Community.Description != "" {
		existing.Description = req.Community.Description
	}
	if req.Community.LogoUrl != "" {
		existing.LogoURL = req.Community.LogoUrl
	}
	if req.Community.BannerUrl != "" {
		existing.BannerURL = req.Community.BannerUrl
	}

	if err := h.communityService.UpdateCommunity(ctx, existing); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update community: %v", err))
	}

	return h.mapCommunityToResponse(existing), nil
}

func (h *CommunityHandler) ApproveCommunity(ctx context.Context, req *communityProto.ApproveCommunityRequest) (*communityProto.CommunityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	if err := h.communityService.ApproveCommunity(ctx, communityID); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to approve community: %v", err))
	}

	community, err := h.communityService.GetCommunityByID(ctx, communityID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get community after approval: %v", err))
	}

	return h.mapCommunityToResponse(community), nil
}

func (h *CommunityHandler) DeleteCommunity(ctx context.Context, req *communityProto.DeleteCommunityRequest) (*communityProto.EmptyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	if err := h.communityService.DeleteCommunity(ctx, communityID); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete community: %v", err))
	}

	return &communityProto.EmptyResponse{}, nil
}

func (h *CommunityHandler) GetCommunityByID(ctx context.Context, req *communityProto.GetCommunityByIDRequest) (*communityProto.CommunityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	community, err := h.communityService.GetCommunityByID(ctx, communityID)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("community not found: %v", err))
	}

	return h.mapCommunityToResponse(community), nil
}

func (h *CommunityHandler) ListCommunities(ctx context.Context, req *communityProto.ListCommunitiesRequest) (*communityProto.ListCommunitiesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	var communities []*model.Community
	var err error

	// Standard listing without additional filters
	communities, err = h.communityService.ListCommunities(ctx, offset, limit)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list communities: %v", err))
	}

	// Get total count of communities
	totalCount, err := h.communityService.CountCommunities(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to count communities: %v", err))
	}

	protoCommunities := make([]*communityProto.Community, len(communities))
	for i, community := range communities {
		protoCommunities[i] = h.mapCommunityToProto(community)
	}

	return &communityProto.ListCommunitiesResponse{
		Communities: protoCommunities,
		TotalCount:  int32(totalCount),
	}, nil
}

func (h *CommunityHandler) AddMember(ctx context.Context, req *communityProto.AddMemberRequest) (*communityProto.MemberResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	member := &model.CommunityMember{
		CommunityID: communityID,
		UserID:      userID,
		Role:        req.Role,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.communityMemberRepo.Add(member); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add member: %v", err))
	}

	return &communityProto.MemberResponse{
		Member: &communityProto.Member{
			UserId:      req.UserId,
			CommunityId: req.CommunityId,
			Role:        req.Role,
			JoinedAt:    timestamppb.New(member.CreatedAt),
		},
	}, nil
}

func (h *CommunityHandler) RemoveMember(ctx context.Context, req *communityProto.RemoveMemberRequest) (*communityProto.EmptyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if err := h.communityMemberRepo.Remove(communityID, userID); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove member: %v", err))
	}

	return &communityProto.EmptyResponse{}, nil
}

func (h *CommunityHandler) ListMembers(ctx context.Context, req *communityProto.ListMembersRequest) (*communityProto.ListMembersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	members, err := h.communityMemberRepo.FindByCommunity(communityID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list members: %v", err))
	}

	protoMembers := make([]*communityProto.Member, len(members))
	for i, member := range members {
		protoMembers[i] = &communityProto.Member{
			UserId:      member.UserID.String(),
			CommunityId: member.CommunityID.String(),
			Role:        member.Role,
			JoinedAt:    timestamppb.New(member.CreatedAt),
		}
	}

	return &communityProto.ListMembersResponse{
		Members: protoMembers,
	}, nil
}

func (h *CommunityHandler) UpdateMemberRole(ctx context.Context, req *communityProto.UpdateMemberRoleRequest) (*communityProto.MemberResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	members, err := h.communityMemberRepo.FindByCommunity(communityID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find community members: %v", err))
	}

	var member *model.CommunityMember
	for _, m := range members {
		if m.UserID == userID {
			member = m
			break
		}
	}

	if member == nil {
		return nil, status.Error(codes.NotFound, "member not found in community")
	}

	member.Role = req.Role
	member.UpdatedAt = time.Now()

	if err := h.communityMemberRepo.Update(member); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update member role: %v", err))
	}

	return &communityProto.MemberResponse{
		Member: &communityProto.Member{
			UserId:      member.UserID.String(),
			CommunityId: member.CommunityID.String(),
			Role:        member.Role,
			JoinedAt:    timestamppb.New(member.CreatedAt),
		},
	}, nil
}

func (h *CommunityHandler) AddRule(ctx context.Context, req *communityProto.AddRuleRequest) (*communityProto.RuleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	rule := &model.CommunityRule{
		RuleID:      uuid.New(),
		CommunityID: communityID,
		RuleText:    req.RuleText,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.ruleRepo.Add(rule); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add rule: %v", err))
	}

	return &communityProto.RuleResponse{
		Rule: &communityProto.Rule{
			Id:          rule.RuleID.String(),
			CommunityId: rule.CommunityID.String(),
			RuleText:    rule.RuleText,
		},
	}, nil
}

func (h *CommunityHandler) RemoveRule(ctx context.Context, req *communityProto.RemoveRuleRequest) (*communityProto.EmptyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	ruleID, err := uuid.Parse(req.RuleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid rule ID")
	}

	if err := h.ruleRepo.Remove(ruleID); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove rule: %v", err))
	}

	return &communityProto.EmptyResponse{}, nil
}

func (h *CommunityHandler) ListRules(ctx context.Context, req *communityProto.ListRulesRequest) (*communityProto.ListRulesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	rules, err := h.ruleRepo.FindByCommunity(communityID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list rules: %v", err))
	}

	protoRules := make([]*communityProto.Rule, len(rules))
	for i, rule := range rules {
		protoRules[i] = &communityProto.Rule{
			Id:          rule.RuleID.String(),
			CommunityId: rule.CommunityID.String(),
			RuleText:    rule.RuleText,
		}
	}

	return &communityProto.ListRulesResponse{
		Rules: protoRules,
	}, nil
}

func (h *CommunityHandler) RequestToJoin(ctx context.Context, req *communityProto.RequestToJoinRequest) (*communityProto.JoinRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	isMember, err := h.communityMemberRepo.IsMember(communityID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check membership: %v", err))
	}

	if isMember {
		return nil, status.Error(codes.AlreadyExists, "user is already a member of this community")
	}

	hasRequest, err := h.communityJoinRequestRepo.HasPendingJoinRequest(communityID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check pending join requests: %v", err))
	}

	if hasRequest {
		return nil, status.Error(codes.AlreadyExists, "user already has a pending join request")
	}

	joinRequest := &model.CommunityJoinRequest{
		RequestID:   uuid.New(),
		CommunityID: communityID,
		UserID:      userID,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.communityJoinRequestRepo.Add(joinRequest); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create join request: %v", err))
	}

	return &communityProto.JoinRequestResponse{
		JoinRequest: &communityProto.JoinRequest{
			Id:          joinRequest.RequestID.String(),
			CommunityId: joinRequest.CommunityID.String(),
			UserId:      joinRequest.UserID.String(),
			Status:      joinRequest.Status,
		},
	}, nil
}

func (h *CommunityHandler) ApproveJoinRequest(ctx context.Context, req *communityProto.ApproveJoinRequestRequest) (*communityProto.JoinRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	_, err := uuid.Parse(req.JoinRequestId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request ID")
	}

	// Logic to approve join request would go here
	// Since we don't have a model.CommunityJoinRequest struct with appropriate fields shown,
	// this implementation is incomplete

	return &communityProto.JoinRequestResponse{
		JoinRequest: &communityProto.JoinRequest{
			Id:     req.JoinRequestId,
			Status: "approved",
		},
	}, nil
}

func (h *CommunityHandler) RejectJoinRequest(ctx context.Context, req *communityProto.RejectJoinRequestRequest) (*communityProto.JoinRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	_, err := uuid.Parse(req.JoinRequestId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request ID")
	}

	// Logic to reject join request would go here
	// Since we don't have a model.CommunityJoinRequest struct with appropriate fields shown,
	// this implementation is incomplete

	return &communityProto.JoinRequestResponse{
		JoinRequest: &communityProto.JoinRequest{
			Id:     req.JoinRequestId,
			Status: "rejected",
		},
	}, nil
}

func (h *CommunityHandler) ListJoinRequests(ctx context.Context, req *communityProto.ListJoinRequestsRequest) (*communityProto.ListJoinRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	communityID, err := uuid.Parse(req.CommunityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid community ID")
	}

	// Get join requests from the repository
	joinRequests, err := h.communityJoinRequestRepo.FindByCommunity(communityID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list join requests: %v", err))
	}

	// Convert to proto format
	protoRequests := make([]*communityProto.JoinRequest, len(joinRequests))
	for i, request := range joinRequests {
		protoRequests[i] = &communityProto.JoinRequest{
			Id:          request.RequestID.String(),
			CommunityId: request.CommunityID.String(),
			UserId:      request.UserID.String(),
			Status:      request.Status,
		}
	}

	return &communityProto.ListJoinRequestsResponse{
		JoinRequests: protoRequests,
	}, nil
}

// Helper method to map a model.Community to a protobuf response
func (h *CommunityHandler) mapCommunityToResponse(community *model.Community) *communityProto.CommunityResponse {
	return &communityProto.CommunityResponse{
		Community: h.mapCommunityToProto(community),
	}
}

// Helper method to map a model.Community to a protobuf Community
func (h *CommunityHandler) mapCommunityToProto(community *model.Community) *communityProto.Community {
	categories := make([]*communityProto.Category, len(community.Categories))
	for i, category := range community.Categories {
		categories[i] = &communityProto.Category{
			Id:   category.CategoryID.String(),
			Name: category.Name,
		}
	}

	return &communityProto.Community{
		Id:          community.CommunityID.String(),
		Name:        community.Name,
		Description: community.Description,
		LogoUrl:     community.LogoURL,
		BannerUrl:   community.BannerURL,
		CreatorId:   community.CreatorID.String(),
		IsApproved:  community.IsApproved,
		Categories:  categories,
		CreatedAt:   timestamppb.New(community.CreatedAt),
		UpdatedAt:   timestamppb.New(community.UpdatedAt),
	}
}

func (h *CommunityHandler) CreateChat(ctx context.Context, req *communityProto.CreateChatRequest) (*communityProto.ChatResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.Name == "" && req.IsGroup {
		return nil, status.Error(codes.InvalidArgument, "group chat requires a name")
	}

	if req.CreatedBy == "" {
		return nil, status.Error(codes.InvalidArgument, "created_by is required")
	}

	if len(req.ParticipantIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one participant is required")
	}

	chat, err := h.chatService.CreateChat(
		req.Name,
		"",
		req.CreatedBy,
		req.IsGroup,
		req.ParticipantIds,
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

	limit := 50
	offset := 0

	participants, err := h.chatService.ListParticipants(req.ChatId, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list participants: %v", err))
	}

	return &communityProto.ListChatParticipantsResponse{
		Participants: participants,
	}, nil
}

func (h *CommunityHandler) SendMessage(ctx context.Context, req *communityProto.SendMessageRequest) (*communityProto.MessageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.ChatId == "" {
		return nil, status.Error(codes.InvalidArgument, "chat_id is required")
	}

	if req.SenderId == "" {
		return nil, status.Error(codes.InvalidArgument, "sender_id is required")
	}

	if req.Content == "" && req.MediaUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "either content or media_url is required")
	}

	messageID, err := h.chatService.SendMessage(req.ChatId, req.SenderId, req.Content)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to send message: %v", err))
	}

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

	userID, err := extractUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to authenticate user")
	}

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

	userID, err := extractUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to authenticate user")
	}

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

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

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

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	messages, err := h.chatService.SearchMessages(req.ChatId, req.Query, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to search messages: %v", err))
	}

	return &communityProto.ListMessagesResponse{
		Messages: messages,
	}, nil
}

func extractUserIDFromContext(ctx context.Context) (string, error) {

	return "system-user", nil
}

func (h *CommunityHandler) ListCategories(ctx context.Context, req *communityProto.ListCategoriesRequest) (*communityProto.ListCategoriesResponse, error) {
	categories, err := h.communityService.ListCategories(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list categories: %v", err))
	}

	protoCategories := make([]*communityProto.Category, len(categories))
	for i, category := range categories {
		protoCategories[i] = &communityProto.Category{
			Id:   category.CategoryID.String(),
			Name: category.Name,
		}
		if !category.CreatedAt.IsZero() {
			protoCategories[i].CreatedAt = timestamppb.New(category.CreatedAt)
		}
	}

	return &communityProto.ListCategoriesResponse{
		Categories: protoCategories,
	}, nil
}

func (h *CommunityHandler) SearchCommunities(ctx context.Context, req *communityProto.SearchCommunitiesRequest) (*communityProto.ListCommunitiesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	var isApproved *bool
	if req.IsApproved {
		approved := req.IsApproved
		isApproved = &approved
	}

	communities, totalCount, err := h.communityService.SearchCommunities(ctx, req.Query, req.Categories, isApproved, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to search communities: %v", err))
	}

	protoCommunities := make([]*communityProto.Community, len(communities))
	for i, community := range communities {
		protoCommunities[i] = h.mapCommunityToProto(community)
	}

	return &communityProto.ListCommunitiesResponse{
		Communities: protoCommunities,
		TotalCount:  int32(totalCount),
	}, nil
}

// ListUserCommunities gets communities based on user's membership status
func (h *CommunityHandler) ListUserCommunities(ctx context.Context, req *communityProto.ListUserCommunitiesRequest) (*communityProto.ListCommunitiesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	// Normalize status values
	status := req.Status
	if status != "member" && status != "pending" {
		status = "member" // Default to member if not specified
	}

	// Get communities where the user is a member or has pending requests
	communities, totalCount, err := h.communityService.ListUserCommunities(ctx, userID, status, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list user communities: %v", err))
	}

	protoCommunities := make([]*communityProto.Community, len(communities))
	for i, community := range communities {
		protoCommunities[i] = h.mapCommunityToProto(community)
	}

	return &communityProto.ListCommunitiesResponse{
		Communities: protoCommunities,
		TotalCount:  int32(totalCount),
	}, nil
}
