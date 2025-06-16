package repository

import (
	"errors"
	"fmt"
	"log"
	"time"

	"aycom/backend/services/user/model"

	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) BanUser(userID string, ban bool) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	log.Printf("BanUser: Attempting to set user %s ban status to %v", userID, ban)

	// First check if the user exists
	var user model.User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		log.Printf("BanUser: Error checking if user exists: %v", err)
		return fmt.Errorf("failed to check if user exists: %w", err)
	}

	log.Printf("BanUser: Found user %s (ID: %s) with current ban status: %v", user.Username, userID, user.IsBanned)

	// Update the user's ban status
	result := r.db.Model(&model.User{}).Where("id = ?", userID).Update("is_banned", ban)
	if result.Error != nil {
		log.Printf("BanUser: Database error updating ban status for user %s: %v", userID, result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		log.Printf("BanUser: No user found with ID %s, no rows affected", userID)
		return errors.New("no user found with the given ID")
	}

	log.Printf("BanUser: Successfully updated ban status for user %s to %v", userID, ban)
	return nil
}

func (r *AdminRepository) CreateNewsletter(newsletter *model.Newsletter) error {
	return r.db.Create(newsletter).Error
}

func (r *AdminRepository) GetSubscribedUsers() ([]model.User, error) {
	var users []model.User
	err := r.db.Where("subscribe_to_newsletter = ?", true).Find(&users).Error
	return users, err
}

func (r *AdminRepository) GetCommunityRequests(page, limit int, status string) ([]model.CommunityRequest, int64, error) {
	var requests []model.CommunityRequest
	var total int64

	query := r.db.Model(&model.CommunityRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	} else {

		query = query.Where("status = ?", "pending")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (r *AdminRepository) GetCommunityRequestByID(id string) (*model.CommunityRequest, error) {
	var request model.CommunityRequest
	err := r.db.Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *AdminRepository) ProcessCommunityRequest(id string, approve bool) error {
	status := "rejected"
	if approve {
		status = "approved"
	}

	log.Printf("Processing community request %s with status: %s", id, status)
	result := r.db.Model(&model.CommunityRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		log.Printf("Error updating community request %s: %v", id, result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		log.Printf("No community request found with ID %s", id)
		return errors.New("community request not found")
	}

	log.Printf("Successfully updated community request %s to status: %s", id, status)
	return nil
}

func (r *AdminRepository) GetPremiumRequests(page, limit int, status string) ([]model.PremiumRequest, int64, error) {
	var requests []model.PremiumRequest
	var total int64

	query := r.db.Model(&model.PremiumRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (r *AdminRepository) GetPremiumRequestByID(id string) (*model.PremiumRequest, error) {
	var request model.PremiumRequest
	err := r.db.Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

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

func (r *AdminRepository) GetReportRequests(page, limit int, status string) ([]model.ReportRequest, int64, error) {
	var requests []model.ReportRequest
	var total int64

	query := r.db.Model(&model.ReportRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (r *AdminRepository) GetReportRequestByID(id string) (*model.ReportRequest, error) {
	var request model.ReportRequest
	err := r.db.Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

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

func (r *AdminRepository) GetThreadCategories(page, limit int) ([]model.ThreadCategory, int64, error) {
	var categories []model.ThreadCategory
	var total int64

	err := r.db.Model(&model.ThreadCategory{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = r.db.Offset(offset).Limit(limit).Order("name ASC").Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *AdminRepository) CreateThreadCategory(category *model.ThreadCategory) error {
	return r.db.Create(category).Error
}

func (r *AdminRepository) UpdateThreadCategory(id string, name, description string) error {
	return r.db.Model(&model.ThreadCategory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
		"updated_at":  time.Now(),
	}).Error
}

func (r *AdminRepository) DeleteThreadCategory(id string) error {
	return r.db.Delete(&model.ThreadCategory{}, "id = ?", id).Error
}

func (r *AdminRepository) GetCommunityCategories(page, limit int) ([]model.CommunityCategory, int64, error) {
	var categories []model.CommunityCategory
	var total int64

	err := r.db.Model(&model.CommunityCategory{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = r.db.Offset(offset).Limit(limit).Order("name ASC").Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *AdminRepository) CreateCommunityCategory(category *model.CommunityCategory) error {
	return r.db.Create(category).Error
}

func (r *AdminRepository) UpdateCommunityCategory(id string, name, description string) error {
	return r.db.Model(&model.CommunityCategory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
		"updated_at":  time.Now(),
	}).Error
}

func (r *AdminRepository) DeleteCommunityCategory(id string) error {
	return r.db.Delete(&model.CommunityCategory{}, "id = ?", id).Error
}

func (r *AdminRepository) CreateCommunityRequest(request *model.CommunityRequest) error {
	return r.db.Create(request).Error
}

func (r *AdminRepository) CreatePremiumRequest(request *model.PremiumRequest) error {
	return r.db.Create(request).Error
}
