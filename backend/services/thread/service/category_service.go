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

// CategoryService defines the interface for category operations
type CategoryService interface {
	CreateCategory(ctx context.Context, name string, categoryType string) (*model.Category, error)
	GetCategoryByID(ctx context.Context, categoryID string) (*model.Category, error)
	GetAllCategories(ctx context.Context, categoryType string) ([]*model.Category, error)
	UpdateCategory(ctx context.Context, categoryID string, name string) (*model.Category, error)
	DeleteCategory(ctx context.Context, categoryID string) error

	// Thread category operations
	AddCategoryToThread(ctx context.Context, threadID string, categoryID string) error
	RemoveCategoryFromThread(ctx context.Context, threadID string, categoryID string) error
	GetThreadCategories(ctx context.Context, threadID string) ([]*model.Category, error)

	// Helper methods for working with category names
	GetOrCreateCategoriesByNames(ctx context.Context, categoryNames []string, categoryType string) ([]string, error)
}

// categoryService implements the CategoryService interface
type categoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

// CreateCategory creates a new category
func (s *categoryService) CreateCategory(ctx context.Context, name string, categoryType string) (*model.Category, error) {
	// Validate required fields
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "Category name is required")
	}

	if categoryType == "" {
		categoryType = "Thread" // Default to Thread type
	}

	// Check if a category with this name already exists
	existingCategory, err := s.categoryRepo.FindCategoryByName(name, categoryType)
	if err == nil {
		// Category already exists
		return existingCategory, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other error occurred
		return nil, status.Errorf(codes.Internal, "Failed to check for existing category: %v", err)
	}

	// Create a new category
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

// GetCategoryByID retrieves a category by its ID
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

// GetAllCategories retrieves all categories of a specific type
func (s *categoryService) GetAllCategories(ctx context.Context, categoryType string) ([]*model.Category, error) {
	categories, err := s.categoryRepo.FindAllCategories(categoryType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve categories: %v", err)
	}

	return categories, nil
}

// UpdateCategory updates a category
func (s *categoryService) UpdateCategory(ctx context.Context, categoryID string, name string) (*model.Category, error) {
	if categoryID == "" {
		return nil, status.Error(codes.InvalidArgument, "Category ID is required")
	}

	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "Category name is required")
	}

	// Get existing category
	category, err := s.categoryRepo.FindCategoryByID(categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Category with ID %s not found", categoryID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve category: %v", err)
	}

	// Update category name
	category.Name = name
	category.UpdatedAt = time.Now()

	if err := s.categoryRepo.UpdateCategory(category); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update category: %v", err)
	}

	return category, nil
}

// DeleteCategory deletes a category
func (s *categoryService) DeleteCategory(ctx context.Context, categoryID string) error {
	if categoryID == "" {
		return status.Error(codes.InvalidArgument, "Category ID is required")
	}

	if err := s.categoryRepo.DeleteCategory(categoryID); err != nil {
		return status.Errorf(codes.Internal, "Failed to delete category: %v", err)
	}

	return nil
}

// AddCategoryToThread adds a category to a thread
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

// RemoveCategoryFromThread removes a category from a thread
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

// GetThreadCategories gets all categories associated with a thread
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

// GetOrCreateCategoriesByNames gets or creates categories by their names and returns their IDs
func (s *categoryService) GetOrCreateCategoriesByNames(ctx context.Context, categoryNames []string, categoryType string) ([]string, error) {
	if len(categoryNames) == 0 {
		return []string{}, nil
	}

	if categoryType == "" {
		categoryType = "Thread" // Default to Thread type
	}

	categoryIDs := make([]string, 0, len(categoryNames))

	for _, name := range categoryNames {
		if name == "" {
			continue
		}

		// Try to find existing category
		category, err := s.categoryRepo.FindCategoryByName(name, categoryType)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new category
				newCategory, err := s.CreateCategory(ctx, name, categoryType)
				if err != nil {
					return nil, err
				}
				categoryIDs = append(categoryIDs, newCategory.CategoryID.String())
			} else {
				return nil, status.Errorf(codes.Internal, "Failed to process category %s: %v", name, err)
			}
		} else {
			// Use existing category
			categoryIDs = append(categoryIDs, category.CategoryID.String())
		}
	}

	return categoryIDs, nil
}
