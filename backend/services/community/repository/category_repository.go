package repository

import (
	"aycom/backend/services/community/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *model.Category) error
	FindByID(id uuid.UUID) (*model.Category, error)
	FindByName(name string) (*model.Category, error)
	List() ([]*model.Category, error)
	Update(category *model.Category) error
	Delete(id uuid.UUID) error
	AddCommunityToCategory(communityID, categoryID uuid.UUID) error
	RemoveCommunityFromCategory(communityID, categoryID uuid.UUID) error
	GetCategoriesByCommunity(communityID uuid.UUID) ([]*model.Category, error)
}

type GormCategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &GormCategoryRepository{db: db}
}

func (r *GormCategoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

func (r *GormCategoryRepository) FindByID(id uuid.UUID) (*model.Category, error) {
	var category model.Category
	err := r.db.First(&category, "category_id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *GormCategoryRepository) FindByName(name string) (*model.Category, error) {
	var category model.Category
	err := r.db.Where("name = ?", name).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *GormCategoryRepository) List() ([]*model.Category, error) {
	var categories []*model.Category
	err := r.db.Find(&categories).Error
	return categories, err
}

func (r *GormCategoryRepository) Update(category *model.Category) error {
	return r.db.Save(category).Error
}

func (r *GormCategoryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Category{}, "category_id = ?", id).Error
}

func (r *GormCategoryRepository) AddCommunityToCategory(communityID, categoryID uuid.UUID) error {
	communityCategory := model.CommunityCategory{
		CommunityID: communityID,
		CategoryID:  categoryID,
	}
	return r.db.Create(&communityCategory).Error
}

func (r *GormCategoryRepository) RemoveCommunityFromCategory(communityID, categoryID uuid.UUID) error {
	return r.db.Delete(&model.CommunityCategory{}, "community_id = ? AND category_id = ?", communityID, categoryID).Error
}

func (r *GormCategoryRepository) GetCategoriesByCommunity(communityID uuid.UUID) ([]*model.Category, error) {
	var categories []*model.Category
	err := r.db.Joins("JOIN community_categories cc ON cc.category_id = categories.category_id").
		Where("cc.community_id = ?", communityID).
		Find(&categories).Error
	return categories, err
}
