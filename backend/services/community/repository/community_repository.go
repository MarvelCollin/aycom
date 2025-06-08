package repository

import (
	"aycom/backend/services/community/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

	// If no categories specified, just return all communities
	if len(categories) == 0 {
		err := r.db.Preload("Categories").Offset(offset).Limit(limit).Find(&communities).Error
		return communities, err
	}

	// Use a subquery approach to avoid join issues
	// Get community IDs that match the category filter
	var communityIDs []uuid.UUID
	categoryQuery := r.db.Table("community_categories").
		Select("community_id").
		Joins("JOIN categories ON categories.category_id = community_categories.category_id").
		Where("categories.name IN ?", categories).
		Group("community_id")

	if err := categoryQuery.Pluck("community_id", &communityIDs).Error; err != nil {
		return nil, err
	}

	// No matches found
	if len(communityIDs) == 0 {
		return []*model.Community{}, nil
	}

	// Get communities by IDs with pagination
	err := r.db.Preload("Categories").
		Where("community_id IN ?", communityIDs).
		Offset(offset).Limit(limit).
		Find(&communities).Error

	return communities, err
}

func (r *GormCommunityRepository) Search(query string, categories []string, isApproved *bool, offset, limit int) ([]*model.Community, int64, error) {
	var communities []*model.Community
	var count int64

	// First build a GORM query that doesn't use joins to avoid the SQL issue
	dbQuery := r.db.Model(&model.Community{})

	// Apply basic filters
	if query != "" {
		searchQuery := "%" + query + "%"
		dbQuery = dbQuery.Where("name ILIKE ? OR description ILIKE ?", searchQuery, searchQuery)
	}

	// Apply is_approved filter if provided
	if isApproved != nil {
		dbQuery = dbQuery.Where("is_approved = ?", *isApproved)
	}

	// If categories are specified, use a subquery approach to avoid join issues
	if len(categories) > 0 {
		// Get community IDs that match the category filter
		var communityIDs []uuid.UUID
		categoryQuery := r.db.Table("community_categories").
			Select("community_id").
			Joins("JOIN categories ON categories.category_id = community_categories.category_id").
			Where("categories.name IN ?", categories).
			Group("community_id")

		if err := categoryQuery.Pluck("community_id", &communityIDs).Error; err != nil {
			return nil, 0, err
		}

		// Apply the community IDs filter
		if len(communityIDs) > 0 {
			dbQuery = dbQuery.Where("community_id IN ?", communityIDs)
		} else {
			// No matching communities found
			return []*model.Community{}, 0, nil
		}
	}

	// Count total matching records
	err := dbQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Execute the query with pagination and eager loading of Categories
	err = dbQuery.Preload("Categories").Offset(offset).Limit(limit).Find(&communities).Error
	if err != nil {
		return nil, 0, err
	}

	return communities, count, nil
}

func (r *GormCommunityRepository) ListByUserMembership(userID uuid.UUID, status string, offset, limit int) ([]*model.Community, int64, error) {
	var communities []*model.Community
	var count int64

	query := r.db.Model(&model.Community{}).
		Preload("Categories")

	if status == "member" {

		query = query.Joins("JOIN community_members cm ON cm.community_id = communities.community_id").
			Where("cm.user_id = ?", userID)
	} else if status == "pending" {

		query = query.Joins("JOIN community_join_requests cjr ON cjr.community_id = communities.community_id").
			Where("cjr.user_id = ? AND cjr.status = 'pending'", userID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Find(&communities).Error
	if err != nil {
		return nil, 0, err
	}

	return communities, count, nil
}

func (r *GormCommunityRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Community{}).Count(&count).Error
	return count, err
}
