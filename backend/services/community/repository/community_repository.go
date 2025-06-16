package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/community/model"
)

type CommunityRepository interface {
	Create(community *model.Community) error
	FindByID(id uuid.UUID) (*model.Community, error)
	FindByName(name string) (*model.Community, error)
	Update(community *model.Community) error
	Delete(id uuid.UUID) error
	List(offset, limit int) ([]*model.Community, error)
	ListByCategories(categories []string, offset, limit int) ([]*model.Community, error)
	Search(query string, categories []string, isApproved *bool, offset, limit int) ([]*model.Community, int64, error)
	ListByUserMembership(userID uuid.UUID, status string, offset, limit int) ([]*model.Community, int64, error)
	CountAll() (int64, error)
}

type GormCommunityRepository struct {
	db *gorm.DB
}

func NewCommunityRepository(db *gorm.DB) CommunityRepository {
	return &GormCommunityRepository{db: db}
}

func (r *GormCommunityRepository) Create(community *model.Community) error {
	return r.db.Create(community).Error
}

func (r *GormCommunityRepository) FindByID(id uuid.UUID) (*model.Community, error) {
	var community model.Community
	err := r.db.First(&community, "community_id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &community, nil
}

func (r *GormCommunityRepository) FindByName(name string) (*model.Community, error) {
	var community model.Community
	err := r.db.First(&community, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &community, nil
}

func (r *GormCommunityRepository) Update(community *model.Community) error {
	return r.db.Save(community).Error
}

func (r *GormCommunityRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Community{}, "community_id = ?", id).Error
}

func (r *GormCommunityRepository) List(offset, limit int) ([]*model.Community, error) {
	var communities []*model.Community
	err := r.db.Offset(offset).Limit(limit).Find(&communities).Error
	return communities, err
}

func (r *GormCommunityRepository) ListByCategories(categories []string, offset, limit int) ([]*model.Community, error) {
	var communities []*model.Community

	// Filter out empty categories
	validCategories := make([]string, 0)
	for _, category := range categories {
		if category != "" {
			validCategories = append(validCategories, category)
		}
	}

	if len(validCategories) == 0 {
		err := r.db.Preload("Categories").Offset(offset).Limit(limit).Find(&communities).Error
		return communities, err
	}

	var communityIDs []uuid.UUID
	categoryQuery := r.db.Table("community_categories").
		Select("community_id").
		Joins("JOIN categories ON categories.category_id = community_categories.category_id").
		Where("categories.name IN ?", validCategories).
		Group("community_id")

	if err := categoryQuery.Pluck("community_id", &communityIDs).Error; err != nil {
		return nil, err
	}

	if len(communityIDs) == 0 {
		return []*model.Community{}, nil
	}

	err := r.db.Preload("Categories").
		Where("community_id IN ?", communityIDs).
		Offset(offset).Limit(limit).
		Find(&communities).Error

	return communities, err
}

func (r *GormCommunityRepository) Search(query string, categories []string, isApproved *bool, offset, limit int) ([]*model.Community, int64, error) {
	var communities []*model.Community
	var count int64

	dbQuery := r.db.Model(&model.Community{})

	if query != "" {
		searchQuery := "%" + query + "%"
		dbQuery = dbQuery.Where("name ILIKE ? OR description ILIKE ?", searchQuery, searchQuery)
	}

	if isApproved != nil {
		dbQuery = dbQuery.Where("is_approved = ?", *isApproved)
	}
	if len(categories) > 0 {
		// Filter out empty categories
		validCategories := make([]string, 0)
		for _, category := range categories {
			if category != "" {
				validCategories = append(validCategories, category)
			}
		}

		// Only proceed if we have valid categories
		if len(validCategories) > 0 {
			var communityIDs []uuid.UUID
			categoryQuery := r.db.Table("community_categories").
				Select("community_id").
				Joins("JOIN categories ON categories.category_id = community_categories.category_id").
				Where("categories.name IN ?", validCategories).
				Group("community_id")

			if err := categoryQuery.Pluck("community_id", &communityIDs).Error; err != nil {
				return nil, 0, err
			}

			if len(communityIDs) > 0 {
				dbQuery = dbQuery.Where("community_id IN ?", communityIDs)
			} else {
				// No communities found with these categories
				return []*model.Community{}, 0, nil
			}
		}
	}

	err := dbQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = dbQuery.Preload("Categories").Offset(offset).Limit(limit).Find(&communities).Error
	if err != nil {
		return nil, 0, err
	}

	return communities, count, nil
}

func (r *GormCommunityRepository) ListByUserMembership(userID uuid.UUID, status string, offset, limit int) ([]*model.Community, int64, error) {
	var communities []*model.Community
	var count int64

	if status != "member" && status != "pending" {
		return nil, 0, fmt.Errorf("invalid status: %s (must be 'member' or 'pending')", status)
	}

	query := r.db.Model(&model.Community{}).
		Preload("Categories").
		Select("DISTINCT communities.*") 

	if status == "member" {

		query = query.Joins("JOIN community_members cm ON cm.community_id = communities.community_id").
			Where("cm.user_id = ? AND cm.deleted_at IS NULL", userID)
	} else if status == "pending" {

		query = query.Joins("JOIN community_join_requests cjr ON cjr.community_id = communities.community_id").
			Where("cjr.user_id = ? AND cjr.status = 'pending' AND cjr.deleted_at IS NULL", userID)
	}

	query = query.Where("communities.deleted_at IS NULL")

	countQuery := r.db.Model(&model.Community{})
	if status == "member" {
		countQuery = countQuery.Joins("JOIN community_members cm ON cm.community_id = communities.community_id").
			Where("cm.user_id = ? AND cm.deleted_at IS NULL", userID)
	} else if status == "pending" {
		countQuery = countQuery.Joins("JOIN community_join_requests cjr ON cjr.community_id = communities.community_id").
			Where("cjr.user_id = ? AND cjr.status = 'pending' AND cjr.deleted_at IS NULL", userID)
	}

	countQuery = countQuery.Where("communities.deleted_at IS NULL")

	err := countQuery.Distinct("communities.community_id").Count(&count).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count communities: %w", err)
	}

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}

	err = query.Offset(offset).Limit(limit).Find(&communities).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch communities: %w", err)
	}

	if communities == nil {
		communities = []*model.Community{}
	}

	return communities, count, nil
}

func (r *GormCommunityRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Community{}).Count(&count).Error
	return count, err
}