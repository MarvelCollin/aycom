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
	ListCommunitiesByCategories(ctx context.Context, categories []string, offset, limit int) ([]*model.Community, error)
	SearchCommunities(ctx context.Context, query string, categories []string, offset, limit int) ([]*model.Community, int64, error)
	ListUserCommunities(ctx context.Context, userID uuid.UUID, status string, offset, limit int) ([]*model.Community, int64, error)
	CountCommunities(ctx context.Context) (int64, error)

	// Category methods
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

type communityService struct {
	communityRepo repository.CommunityRepository
	categoryRepo  repository.CategoryRepository
	memberRepo    repository.CommunityMemberRepository
	joinRepo      repository.CommunityJoinRequestRepository
	ruleRepo      repository.CommunityRuleRepository
}

func NewCommunityService(
	communityRepo repository.CommunityRepository,
	categoryRepo repository.CategoryRepository,
	memberRepo repository.CommunityMemberRepository,
	joinRepo repository.CommunityJoinRequestRepository,
	ruleRepo repository.CommunityRuleRepository,
) CommunityService {
	return &communityService{
		communityRepo: communityRepo,
		categoryRepo:  categoryRepo,
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

func (s *communityService) ListCommunitiesByCategories(ctx context.Context, categories []string, offset, limit int) ([]*model.Community, error) {
	return s.communityRepo.ListByCategories(categories, offset, limit)
}

func (s *communityService) SearchCommunities(ctx context.Context, query string, categories []string, offset, limit int) ([]*model.Community, int64, error) {
	return s.communityRepo.Search(query, categories, offset, limit)
}

func (s *communityService) ListUserCommunities(ctx context.Context, userID uuid.UUID, status string, offset, limit int) ([]*model.Community, int64, error) {
	return s.communityRepo.ListByUserMembership(userID, status, offset, limit)
}

func (s *communityService) CountCommunities(ctx context.Context) (int64, error) {
	return s.communityRepo.CountAll()
}

// Category methods
func (s *communityService) CreateCategory(ctx context.Context, category *model.Category) error {
	return s.categoryRepo.Create(category)
}

func (s *communityService) GetCategoryByID(ctx context.Context, categoryID uuid.UUID) (*model.Category, error) {
	return s.categoryRepo.FindByID(categoryID)
}

func (s *communityService) GetCategoryByName(ctx context.Context, name string) (*model.Category, error) {
	return s.categoryRepo.FindByName(name)
}

func (s *communityService) ListCategories(ctx context.Context) ([]*model.Category, error) {
	return s.categoryRepo.List()
}

func (s *communityService) UpdateCategory(ctx context.Context, category *model.Category) error {
	return s.categoryRepo.Update(category)
}

func (s *communityService) DeleteCategory(ctx context.Context, categoryID uuid.UUID) error {
	return s.categoryRepo.Delete(categoryID)
}

func (s *communityService) AddCommunityToCategory(ctx context.Context, communityID, categoryID uuid.UUID) error {
	return s.categoryRepo.AddCommunityToCategory(communityID, categoryID)
}

func (s *communityService) RemoveCommunityFromCategory(ctx context.Context, communityID, categoryID uuid.UUID) error {
	return s.categoryRepo.RemoveCommunityFromCategory(communityID, categoryID)
}

func (s *communityService) GetCommunityCategoriesById(ctx context.Context, communityID uuid.UUID) ([]*model.Category, error) {
	return s.categoryRepo.GetCategoriesByCommunity(communityID)
}
