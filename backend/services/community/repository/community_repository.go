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
