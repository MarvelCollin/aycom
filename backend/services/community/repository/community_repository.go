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
	Search(query string, categories []string, offset, limit int) ([]*model.Community, int64, error)
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
	query := r.db.Preload("Categories")

	if len(categories) > 0 {
		query = query.Joins("JOIN community_categories cc ON cc.community_id = communities.community_id").
			Joins("JOIN categories cat ON cat.category_id = cc.category_id").
			Where("cat.name IN ?", categories).
			Group("communities.community_id")
	}

	err := query.Offset(offset).Limit(limit).Find(&communities).Error
	return communities, err
}

func (r *GormCommunityRepository) Search(query string, categories []string, offset, limit int) ([]*model.Community, int64, error) {
	var communities []*model.Community
	var count int64

	dbQuery := r.db.Model(&model.Community{}).
		Preload("Categories")

	if len(categories) > 0 {
		dbQuery = dbQuery.Joins("JOIN community_categories cc ON cc.community_id = communities.community_id").
			Joins("JOIN categories cat ON cat.category_id = cc.category_id").
			Where("cat.name IN ?", categories).
			Group("communities.community_id")
	}

	if query != "" {
		searchQuery := "%" + query + "%"
		dbQuery = dbQuery.Where("communities.name ILIKE ? OR communities.description ILIKE ?", searchQuery, searchQuery)
	}

	err := dbQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = dbQuery.Offset(offset).Limit(limit).Find(&communities).Error
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