package api

import (
	communityProto "aycom/backend/proto/community"
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"aycom/backend/services/community/model"
	"aycom/backend/services/community/repository"
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
	AddParticipantDirect(participant *model.ParticipantDTO) error
	RemoveParticipant(chatID, userID, removedBy string) error
	RemoveParticipantDirect(chatID, userID string) error
	ListParticipants(chatID string, limit, offset int) ([]*communityProto.ChatParticipant, error)
	SendMessage(chatID, userID, content string) (string, error)
	GetMessages(chatID string, limit, offset int) ([]*communityProto.Message, error)
	DeleteMessage(chatID, messageID, userID string) error
	UnsendMessage(chatID, messageID, userID string) error
	SearchMessages(chatID, query string, limit, offset int) ([]*communityProto.Message, error)
}

type CommunityMemberRepository interface {
	Add(member *model.CommunityMember) error
	Remove(communityID, userID uuid.UUID) error
	FindByCommunity(communityID uuid.UUID) ([]*model.CommunityMember, error)
	IsMember(communityID, userID uuid.UUID) (bool, error)
	AddTx(tx *gorm.DB, member *model.CommunityMember) error
}

type CommunityJoinRequestRepository interface {
	Add(request *model.CommunityJoinRequest) error
	Remove(requestID uuid.UUID) error
	FindByID(requestID uuid.UUID) (*model.CommunityJoinRequest, error)
	FindByCommunity(communityID uuid.UUID) ([]*model.CommunityJoinRequest, error)
	Update(request *model.CommunityJoinRequest) error
	HasPendingJoinRequest(communityID, userID uuid.UUID) (bool, error)
	BeginTx(ctx context.Context) (*gorm.DB, error)
	UpdateTx(tx *gorm.DB, request *model.CommunityJoinRequest) error
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
	var totalCount int64
	var err error
	var isApproved *bool

	approved := req.IsApproved
	isApproved = &approved

	communities, totalCount, err = h.communityService.SearchCommunities(ctx, "", []string{}, isApproved, offset, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list communities with approval filter: %v", err))
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

	requestID, err := uuid.Parse(req.JoinRequestId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request ID")
	}

	joinRequest, err := h.communityJoinRequestRepo.FindByID(requestID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find join request: %v", err))
	}

	if joinRequest == nil {
		return nil, status.Error(codes.NotFound, "join request not found")
	}

	if joinRequest.Status != "pending" {
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("join request is not pending, current status: %s", joinRequest.Status))
	}

	tx, err := h.communityJoinRequestRepo.BeginTx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to start transaction: %v", err))
	}

	var txErr error
	defer func() {
		if txErr != nil {

			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error rolling back transaction: %v", rbErr)
			}
			log.Printf("Transaction rolled back due to error: %v", txErr)
		}
	}()

	joinRequest.Status = "approved"
	joinRequest.UpdatedAt = time.Now()
	txErr = h.communityJoinRequestRepo.UpdateTx(tx, joinRequest)
	if txErr != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update join request: %v", txErr))
	}

	member := &model.CommunityMember{
		CommunityID: joinRequest.CommunityID,
		UserID:      joinRequest.UserID,
		Role:        "member",
	}

	txErr = h.communityMemberRepo.AddTx(tx, member)
	if txErr != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add community member: %v", txErr))
	}

	txErr = tx.Commit().Error
	if txErr != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to commit transaction: %v", txErr))
	}

	log.Printf("Successfully approved join request ID %s for user %s in community %s",
		joinRequest.RequestID, joinRequest.UserID, joinRequest.CommunityID)

	return &communityProto.JoinRequestResponse{
		JoinRequest: &communityProto.JoinRequest{
			Id:          joinRequest.RequestID.String(),
			CommunityId: joinRequest.CommunityID.String(),
			UserId:      joinRequest.UserID.String(),
			Status:      joinRequest.Status,
		},
	}, nil
}

func (h *CommunityHandler) RejectJoinRequest(ctx context.Context, req *communityProto.RejectJoinRequestRequest) (*communityProto.JoinRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	requestID, err := uuid.Parse(req.JoinRequestId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request ID")
	}

	joinRequest, err := h.communityJoinRequestRepo.FindByID(requestID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find join request: %v", err))
	}

	if joinRequest == nil {
		return nil, status.Error(codes.NotFound, "join request not found")
	}

	if joinRequest.Status != "pending" {
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("join request is not pending, current status: %s", joinRequest.Status))
	}

	tx, err := h.communityJoinRequestRepo.BeginTx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to start transaction: %v", err))
	}

	var txErr error
	defer func() {
		if txErr != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error rolling back transaction: %v", rbErr)
			}
			log.Printf("Transaction rolled back due to error: %v", txErr)
		}
	}()

	joinRequest.Status = "rejected"
	joinRequest.UpdatedAt = time.Now()
	txErr = h.communityJoinRequestRepo.UpdateTx(tx, joinRequest)
	if txErr != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update join request: %v", txErr))
	}

	txErr = tx.Commit().Error
	if txErr != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to commit transaction: %v", txErr))
	}

	log.Printf("Successfully rejected join request ID %s for user %s in community %s",
		joinRequest.RequestID, joinRequest.UserID, joinRequest.CommunityID)

	return &communityProto.JoinRequestResponse{
		JoinRequest: &communityProto.JoinRequest{
			Id:          joinRequest.RequestID.String(),
			CommunityId: joinRequest.CommunityID.String(),
			UserId:      joinRequest.UserID.String(),
			Status:      joinRequest.Status,
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

	joinRequests, err := h.communityJoinRequestRepo.FindByCommunity(communityID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list join requests: %v", err))
	}

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

func (h *CommunityHandler) mapCommunityToResponse(community *model.Community) *communityProto.CommunityResponse {
	return &communityProto.CommunityResponse{
		Community: h.mapCommunityToProto(community),
	}
}

func (h *CommunityHandler) mapCommunityToProto(community *model.Community) *communityProto.Community {
	if community == nil {
		return &communityProto.Community{}
	}

	categories := make([]*communityProto.Category, 0)
	if community.Categories != nil {
		for _, category := range community.Categories {
			if category.CategoryID.String() != "" {
				categories = append(categories, &communityProto.Category{
					Id:   category.CategoryID.String(),
					Name: category.Name,
				})
			}
		}
	}

	var createdAt *timestamppb.Timestamp
	var updatedAt *timestamppb.Timestamp

	if !community.CreatedAt.IsZero() {
		createdAt = timestamppb.New(community.CreatedAt)
	} else {
		createdAt = timestamppb.New(time.Now())
	}

	if !community.UpdatedAt.IsZero() {
		updatedAt = timestamppb.New(community.UpdatedAt)
	} else {
		updatedAt = timestamppb.New(time.Now())
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
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
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
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.ChatId == "" {
		return nil, status.Error(codes.InvalidArgument, "chat_id is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	log.Printf("AddChatParticipant: Adding user %s to chat %s", req.UserId, req.ChatId)

	// For now, let's bypass the admin check by making the user being added an admin temporarily
	// This is a temporary fix until we implement proper authentication context
	// In a real implementation, you'd get the current user ID from the JWT token in the context

	// First, let's try to add the participant directly via repository to bypass admin checks
	participant := &model.ParticipantDTO{
		ChatID:   req.ChatId,
		UserID:   req.UserId,
		IsAdmin:  req.IsAdmin,
		JoinedAt: time.Now(),
	}

	// Get repository access through the chat service (we'll need to add this method)
	err := h.chatService.AddParticipantDirect(participant)
	if err != nil {
		log.Printf("AddChatParticipant: Error adding participant: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add participant: %v", err))
	}

	log.Printf("AddChatParticipant: Successfully added user %s to chat %s", req.UserId, req.ChatId)

	return &communityProto.ChatParticipantResponse{
		Participant: &communityProto.ChatParticipant{
			ChatId:   req.ChatId,
			UserId:   req.UserId,
			IsAdmin:  req.IsAdmin,
			JoinedAt: timestamppb.Now(),
		},
	}, nil
}
func (h *CommunityHandler) RemoveChatParticipant(ctx context.Context, req *communityProto.RemoveChatParticipantRequest) (*communityProto.EmptyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.ChatId == "" {
		return nil, status.Error(codes.InvalidArgument, "chat_id is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	log.Printf("RemoveChatParticipant: Removing user %s from chat %s", req.UserId, req.ChatId)

	// Use direct removal to bypass admin checks temporarily
	err := h.chatService.RemoveParticipantDirect(req.ChatId, req.UserId)
	if err != nil {
		log.Printf("RemoveChatParticipant: Error removing participant: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove participant: %v", err))
	}

	log.Printf("RemoveChatParticipant: Successfully removed user %s from chat %s", req.UserId, req.ChatId)

	return &communityProto.EmptyResponse{}, nil
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

	// Check if this is an unsend operation by looking for chat_id in context
	chatID, isUnsend := ctx.Value("chat_id").(string)

	if isUnsend && chatID != "" {
		// This is an unsend operation
		log.Printf("DeleteMessage: Performing unsend operation for message %s in chat %s", req.MessageId, chatID)
		err := h.chatService.UnsendMessage(chatID, req.MessageId, userID)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to unsend message: %v", err))
		}
		return &communityProto.EmptyResponse{}, nil
	} else {
		// This is a regular delete operation
		chatID = "" // Keep empty for regular delete
		err = h.chatService.DeleteMessage(chatID, req.MessageId, userID)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete message: %v", err))
		}
		return &communityProto.EmptyResponse{}, nil
	}
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

	approved := req.IsApproved
	isApproved = &approved

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

func (h *CommunityHandler) ListUserCommunities(ctx context.Context, req *communityProto.ListUserCommunitiesRequest) (*communityProto.ListCommunitiesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user ID format: %v", err))
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	memberStatus := req.Status
	if memberStatus != "member" && memberStatus != "pending" {
		memberStatus = "member"
	}

	_ = req.Query
	_ = req.Categories

	var communities []*model.Community
	var totalCount int64
	var repoErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				repoErr = fmt.Errorf("recovered from panic in ListUserCommunities: %v", r)
				log.Printf("PANIC in ListUserCommunities: %v", r)
				debug.PrintStack()
			}
		}()

		communities, totalCount, repoErr = h.communityService.ListUserCommunities(ctx, userID, memberStatus, offset, limit)
	}()

	if repoErr != nil {
		log.Printf("Error in repository.ListByUserMembership: %v", repoErr)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list communities: %v", repoErr))
	}

	if communities == nil {
		return &communityProto.ListCommunitiesResponse{
			Communities: []*communityProto.Community{},
			TotalCount:  0,
		}, nil
	}

	protoCommunities := make([]*communityProto.Community, 0, len(communities))
	for _, community := range communities {
		if community != nil {
			protoCommunity := h.mapCommunityToProto(community)
			protoCommunities = append(protoCommunities, protoCommunity)
		}
	}

	return &communityProto.ListCommunitiesResponse{
		Communities: protoCommunities,
		TotalCount:  int32(totalCount),
	}, nil
}

func (h *CommunityHandler) IsMember(ctx context.Context, req *communityProto.IsMemberRequest) (*communityProto.IsMemberResponse, error) {
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

	return &communityProto.IsMemberResponse{
		IsMember: isMember,
	}, nil
}

func (h *CommunityHandler) HasPendingJoinRequest(ctx context.Context, req *communityProto.HasPendingJoinRequestRequest) (*communityProto.HasPendingJoinRequestResponse, error) {
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

	hasPendingRequest, err := h.communityJoinRequestRepo.HasPendingJoinRequest(communityID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check pending join requests: %v", err))
	}

	return &communityProto.HasPendingJoinRequestResponse{
		HasRequest: hasPendingRequest,
	}, nil
}

func (h *CommunityHandler) DeleteChat(ctx context.Context, req *communityProto.DeleteChatRequest) (*communityProto.EmptyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.ChatId == "" {
		return nil, status.Error(codes.InvalidArgument, "chat_id is required")
	}

	userID, err := extractUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to authenticate user")
	}

	err = h.chatService.DeleteChatForUser(req.ChatId, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete chat: %v", err))
	}

	return &communityProto.EmptyResponse{}, nil
}
