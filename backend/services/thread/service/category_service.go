package service

import (
	"context"
	"errors"
	"time"

	"aycom/backend/services/thread/model"
	"aycom/backend/services/thread/repository"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, name string, categoryType string) (*model.Category, error)
	GetCategoryByID(ctx context.Context, categoryID string) (*model.Category, error)
	GetAllCategories(ctx context.Context, categoryType string) ([]*model.Category, error)
	UpdateCategory(ctx context.Context, categoryID string, name string) (*model.Category, error)
	DeleteCategory(ctx context.Context, categoryID string) error

	AddCategoryToThread(ctx context.Context, threadID string, categoryID string) error
	RemoveCategoryFromThread(ctx context.Context, threadID string, categoryID string) error
	GetThreadCategories(ctx context.Context, threadID string) ([]*model.Category, error)

	GetOrCreateCategoriesByNames(ctx context.Context, categoryNames []string, categoryType string) ([]string, error)
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *categoryService) CreateCategory(ctx context.Context, name string, categoryType string) (*model.Category, error) {
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "Category name is required")
	}

	if categoryType == "" {
		categoryType = "Thread"
	}

	existingCategory, err := s.categoryRepo.FindCategoryByName(name, categoryType)
	if err == nil {
		return existingCategory, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.Internal, "Failed to check for existing category: %v", err)
	}

	category := &model.Category{
		CategoryID: uuid.New(),
		Name:       name,
		Type:       categoryType,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.categoryRepo.CreateCategory(category); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create category: %v", err)
	}

	return category, nil
}

func (s *categoryService) GetCategoryByID(ctx context.Context, categoryID string) (*model.Category, error) {
	if categoryID == "" {
		return nil, status.Error(codes.InvalidArgument, "Category ID is required")
	}

	category, err := s.categoryRepo.FindCategoryByID(categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Category with ID %s not found", categoryID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve category: %v", err)
	}

	return category, nil
}

func (s *categoryService) GetAllCategories(ctx context.Context, categoryType string) ([]*model.Category, error) {
	categories, err := s.categoryRepo.FindAllCategories(categoryType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve categories: %v", err)
	}

	return categories, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, categoryID string, name string) (*model.Category, error) {
	if categoryID == "" {
		return nil, status.Error(codes.InvalidArgument, "Category ID is required")
	}

	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "Category name is required")
	}

	category, err := s.categoryRepo.FindCategoryByID(categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Category with ID %s not found", categoryID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve category: %v", err)
	}

	category.Name = name
	category.UpdatedAt = time.Now()

	if err := s.categoryRepo.UpdateCategory(category); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update category: %v", err)
	}

	return category, nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, categoryID string) error {
	if categoryID == "" {
		return status.Error(codes.InvalidArgument, "Category ID is required")
	}

	if err := s.categoryRepo.DeleteCategory(categoryID); err != nil {
		return status.Errorf(codes.Internal, "Failed to delete category: %v", err)
	}

	return nil
}

func (s *categoryService) AddCategoryToThread(ctx context.Context, threadID string, categoryID string) error {
	if threadID == "" {
		return status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	if categoryID == "" {
		return status.Error(codes.InvalidArgument, "Category ID is required")
	}

	if err := s.categoryRepo.AddCategoryToThread(threadID, categoryID); err != nil {
		return status.Errorf(codes.Internal, "Failed to add category to thread: %v", err)
	}

	return nil
}

func (s *categoryService) RemoveCategoryFromThread(ctx context.Context, threadID string, categoryID string) error {
	if threadID == "" {
		return status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	if categoryID == "" {
		return status.Error(codes.InvalidArgument, "Category ID is required")
	}

	if err := s.categoryRepo.RemoveCategoryFromThread(threadID, categoryID); err != nil {
		return status.Errorf(codes.Internal, "Failed to remove category from thread: %v", err)
	}

	return nil
}

func (s *categoryService) GetThreadCategories(ctx context.Context, threadID string) ([]*model.Category, error) {
	if threadID == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	categories, err := s.categoryRepo.GetThreadCategories(threadID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve thread categories: %v", err)
	}

	return categories, nil
}

func (s *categoryService) GetOrCreateCategoriesByNames(ctx context.Context, categoryNames []string, categoryType string) ([]string, error) {
	if len(categoryNames) == 0 {
		return []string{}, nil
	}

	if categoryType == "" {
		categoryType = "Thread"
	}

	categoryIDs := make([]string, 0, len(categoryNames))

	for _, name := range categoryNames {
		if name == "" {
			continue
		}

		category, err := s.categoryRepo.FindCategoryByName(name, categoryType)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newCategory, err := s.CreateCategory(ctx, name, categoryType)
				if err != nil {
					return nil, err
				}
				categoryIDs = append(categoryIDs, newCategory.CategoryID.String())
			} else {
				return nil, status.Errorf(codes.Internal, "Failed to process category %s: %v", name, err)
			}
		} else {
			categoryIDs = append(categoryIDs, category.CategoryID.String())
		}
	}

	return categoryIDs, nil
}