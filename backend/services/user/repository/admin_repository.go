package repository

import (
	"errors"
	"time"

	"aycom/backend/services/user/model"

	"gorm.io/gorm"
)

// AdminRepository handles admin-related database operations
type AdminRepository struct {
	db *gorm.DB
}

// NewAdminRepository creates a new admin repository
func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// BanUser sets the banned status for a user
func (r *AdminRepository) BanUser(userID string, ban bool) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("is_banned", ban).Error
}

// CreateNewsletter creates a new newsletter record
func (r *AdminRepository) CreateNewsletter(newsletter *model.Newsletter) error {
	return r.db.Create(newsletter).Error
}

// GetSubscribedUsers gets all users who have subscribed to the newsletter
func (r *AdminRepository) GetSubscribedUsers() ([]model.User, error) {
	var users []model.User
	err := r.db.Where("subscribe_to_newsletter = ?", true).Find(&users).Error
	return users, err
}

// Community Request Methods

// GetCommunityRequests retrieves community creation requests with pagination
func (r *AdminRepository) GetCommunityRequests(page, limit int, status string) ([]model.CommunityRequest, int64, error) {
	var requests []model.CommunityRequest
	var total int64

	query := r.db.Model(&model.CommunityRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

// GetCommunityRequestByID retrieves a specific community request
func (r *AdminRepository) GetCommunityRequestByID(id string) (*model.CommunityRequest, error) {
	var request model.CommunityRequest
	err := r.db.Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// ProcessCommunityRequest updates the status of a community request
func (r *AdminRepository) ProcessCommunityRequest(id string, approve bool) error {
	status := "rejected"
	if approve {
		status = "approved"
	}

	return r.db.Model(&model.CommunityRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}

// Premium Request Methods

// GetPremiumRequests retrieves premium user requests with pagination
func (r *AdminRepository) GetPremiumRequests(page, limit int, status string) ([]model.PremiumRequest, int64, error) {
	var requests []model.PremiumRequest
	var total int64

	query := r.db.Model(&model.PremiumRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

// GetPremiumRequestByID retrieves a specific premium request
func (r *AdminRepository) GetPremiumRequestByID(id string) (*model.PremiumRequest, error) {
	var request model.PremiumRequest
	err := r.db.Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// ProcessPremiumRequest updates the status of a premium request
func (r *AdminRepository) ProcessPremiumRequest(id string, approve bool) error {
	status := "rejected"
	if approve {
		status = "approved"
	}

	return r.db.Model(&model.PremiumRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}

// Report Request Methods

// GetReportRequests retrieves user report requests with pagination
func (r *AdminRepository) GetReportRequests(page, limit int, status string) ([]model.ReportRequest, int64, error) {
	var requests []model.ReportRequest
	var total int64

	query := r.db.Model(&model.ReportRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

// GetReportRequestByID retrieves a specific report request
func (r *AdminRepository) GetReportRequestByID(id string) (*model.ReportRequest, error) {
	var request model.ReportRequest
	err := r.db.Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// ProcessReportRequest updates the status of a report request
func (r *AdminRepository) ProcessReportRequest(id string, approve bool) error {
	status := "rejected"
	if approve {
		status = "approved"
	}

	return r.db.Model(&model.ReportRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}

// Thread Category Methods

// GetThreadCategories retrieves thread categories with pagination
func (r *AdminRepository) GetThreadCategories(page, limit int) ([]model.ThreadCategory, int64, error) {
	var categories []model.ThreadCategory
	var total int64

	// Get total count
	err := r.db.Model(&model.ThreadCategory{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err = r.db.Offset(offset).Limit(limit).Order("name ASC").Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

// CreateThreadCategory creates a new thread category
func (r *AdminRepository) CreateThreadCategory(category *model.ThreadCategory) error {
	return r.db.Create(category).Error
}

// UpdateThreadCategory updates an existing thread category
func (r *AdminRepository) UpdateThreadCategory(id string, name, description string) error {
	return r.db.Model(&model.ThreadCategory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
		"updated_at":  time.Now(),
	}).Error
}

// DeleteThreadCategory deletes a thread category
func (r *AdminRepository) DeleteThreadCategory(id string) error {
	return r.db.Delete(&model.ThreadCategory{}, "id = ?", id).Error
}

// Community Category Methods

// GetCommunityCategories retrieves community categories with pagination
func (r *AdminRepository) GetCommunityCategories(page, limit int) ([]model.CommunityCategory, int64, error) {
	var categories []model.CommunityCategory
	var total int64

	// Get total count
	err := r.db.Model(&model.CommunityCategory{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err = r.db.Offset(offset).Limit(limit).Order("name ASC").Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

// CreateCommunityCategory creates a new community category
func (r *AdminRepository) CreateCommunityCategory(category *model.CommunityCategory) error {
	return r.db.Create(category).Error
}

// UpdateCommunityCategory updates an existing community category
func (r *AdminRepository) UpdateCommunityCategory(id string, name, description string) error {
	return r.db.Model(&model.CommunityCategory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
		"updated_at":  time.Now(),
	}).Error
}

// DeleteCommunityCategory deletes a community category
func (r *AdminRepository) DeleteCommunityCategory(id string) error {
	return r.db.Delete(&model.CommunityCategory{}, "id = ?", id).Error
}
