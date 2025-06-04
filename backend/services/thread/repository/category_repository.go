package repository

import (
	"errors"

	"aycom/backend/services/thread/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepository interface {

	CreateCategory(category *model.Category) error
	FindCategoryByID(id string) (*model.Category, error)
	FindCategoryByName(name string, categoryType string) (*model.Category, error)
	FindAllCategories(categoryType string) ([]*model.Category, error)
	UpdateCategory(category *model.Category) error
	DeleteCategory(id string) error

	AddCategoryToThread(threadID string, categoryID string) error
	RemoveCategoryFromThread(threadID string, categoryID string) error
	GetThreadCategories(threadID string) ([]*model.Category, error)
}

type PostgresCategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

func (r *PostgresCategoryRepository) CreateCategory(category *model.Category) error {
	if category.CategoryID == uuid.Nil {
		category.CategoryID = uuid.New()
	}
	return r.db.Create(category).Error
}

func (r *PostgresCategoryRepository) FindCategoryByID(id string) (*model.Category, error) {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid UUID format for category ID")
	}

	var category model.Category
	result := r.db.Where("category_id = ?", categoryID).First(&category)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &category, nil
}

func (r *PostgresCategoryRepository) FindCategoryByName(name string, categoryType string) (*model.Category, error) {
	var category model.Category
	result := r.db.Where("name = ? AND type = ?", name, categoryType).First(&category)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &category, nil
}

func (r *PostgresCategoryRepository) FindAllCategories(categoryType string) ([]*model.Category, error) {
	var categories []*model.Category
	query := r.db.Order("name ASC")

	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}

	result := query.Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

func (r *PostgresCategoryRepository) UpdateCategory(category *model.Category) error {
	return r.db.Save(category).Error
}

func (r *PostgresCategoryRepository) DeleteCategory(id string) error {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid UUID format for category ID")
	}

	return r.db.Delete(&model.Category{}, "category_id = ?", categoryID).Error
}

func (r *PostgresCategoryRepository) AddCategoryToThread(threadID string, categoryID string) error {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return errors.New("invalid UUID format for thread ID")
	}

	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return errors.New("invalid UUID format for category ID")
	}

	var count int64
	r.db.Model(&model.ThreadCategory{}).
		Where("thread_id = ? AND category_id = ?", threadUUID, categoryUUID).
		Count(&count)

	if count > 0 {
		return nil 
	}

	threadCategory := &model.ThreadCategory{
		ThreadID:   threadUUID,
		CategoryID: categoryUUID,
	}
	return r.db.Create(threadCategory).Error
}

func (r *PostgresCategoryRepository) RemoveCategoryFromThread(threadID string, categoryID string) error {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return errors.New("invalid UUID format for thread ID")
	}

	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return errors.New("invalid UUID format for category ID")
	}

	return r.db.Where("thread_id = ? AND category_id = ?", threadUUID, categoryUUID).
		Delete(&model.ThreadCategory{}).Error
}

func (r *PostgresCategoryRepository) GetThreadCategories(threadID string) ([]*model.Category, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return nil, errors.New("invalid UUID format for thread ID")
	}

	var categories []*model.Category
	result := r.db.Table("categories").
		Joins("JOIN thread_categories ON categories.category_id = thread_categories.category_id").
		Where("thread_categories.thread_id = ?", threadUUID).
		Find(&categories)

	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}