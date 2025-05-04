package service

import (
	"aycom/backend/services/community/model"
	"aycom/backend/services/community/repository"
	"context"

	"github.com/google/uuid"
)

type CommunityService interface {
	CreateCommunity(ctx context.Context, community *model.Community) error
	UpdateCommunity(ctx context.Context, community *model.Community) error
	ApproveCommunity(ctx context.Context, communityID uuid.UUID) error
	DeleteCommunity(ctx context.Context, communityID uuid.UUID) error
	GetCommunityByID(ctx context.Context, communityID uuid.UUID) (*model.Community, error)
	ListCommunities(ctx context.Context, offset, limit int) ([]*model.Community, error)
}

type communityService struct {
	communityRepo repository.CommunityRepository
	memberRepo    repository.CommunityMemberRepository
	joinRepo      repository.CommunityJoinRequestRepository
	ruleRepo      repository.CommunityRuleRepository
}

func NewCommunityService(
	communityRepo repository.CommunityRepository,
	memberRepo repository.CommunityMemberRepository,
	joinRepo repository.CommunityJoinRequestRepository,
	ruleRepo repository.CommunityRuleRepository,
) CommunityService {
	return &communityService{
		communityRepo: communityRepo,
		memberRepo:    memberRepo,
		joinRepo:      joinRepo,
		ruleRepo:      ruleRepo,
	}
}

func (s *communityService) CreateCommunity(ctx context.Context, community *model.Community) error {
	return s.communityRepo.Create(community)
}

func (s *communityService) UpdateCommunity(ctx context.Context, community *model.Community) error {
	return s.communityRepo.Update(community)
}

func (s *communityService) ApproveCommunity(ctx context.Context, communityID uuid.UUID) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return err
	}
	community.IsApproved = true
	return s.communityRepo.Update(community)
}

func (s *communityService) DeleteCommunity(ctx context.Context, communityID uuid.UUID) error {
	return s.communityRepo.Delete(communityID)
}

func (s *communityService) GetCommunityByID(ctx context.Context, communityID uuid.UUID) (*model.Community, error) {
	return s.communityRepo.FindByID(communityID)
}

func (s *communityService) ListCommunities(ctx context.Context, offset, limit int) ([]*model.Community, error) {
	return s.communityRepo.List(offset, limit)
}
