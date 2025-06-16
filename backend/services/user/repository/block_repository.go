package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/user/model"

)

type BlockRepositoryInterface interface {
	BlockUser(blockerID, blockedID uuid.UUID) error
	UnblockUser(blockerID, blockedID uuid.UUID) error
	IsUserBlocked(userID, blockerID uuid.UUID) (bool, error)
	GetBlockedUsers(blockerID uuid.UUID, page, limit int) ([]model.User, int64, error)
}

type ReportRepositoryInterface interface {
	ReportUser(reporterID, reportedID uuid.UUID, reason string) error
	GetUserReports(page, limit int, status string) ([]model.UserReport, int64, error)
	ProcessReport(reportID uuid.UUID, approved bool, adminID uuid.UUID, adminNotes string) error
}

type BlockRepository struct {
	db *gorm.DB
}

func NewBlockRepository(db *gorm.DB) *BlockRepository {
	return &BlockRepository{db: db}
}

func (r *BlockRepository) BlockUser(blockerID, blockedID uuid.UUID) error {

	var existingBlock model.UserBlock
	if err := r.db.Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).First(&existingBlock).Error; err == nil {

		return nil
	}

	block := model.UserBlock{
		BlockerID: blockerID,
		BlockedID: blockedID,
	}

	return r.db.Create(&block).Error
}

func (r *BlockRepository) UnblockUser(blockerID, blockedID uuid.UUID) error {
	return r.db.Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).Delete(&model.UserBlock{}).Error
}

func (r *BlockRepository) IsUserBlocked(userID, blockerID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.UserBlock{}).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, userID).
		Count(&count).Error

	return count > 0, err
}

func (r *BlockRepository) GetBlockedUsers(blockerID uuid.UUID, page, limit int) ([]model.User, int64, error) {
	var blockedUsers []model.User
	var total int64

	err := r.db.Model(&model.UserBlock{}).
		Where("blocker_id = ?", blockerID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = r.db.Table("users").
		Select("users.*").
		Joins("INNER JOIN user_blocks ON user_blocks.blocked_id = users.id").
		Where("user_blocks.blocker_id = ?", blockerID).
		Offset(offset).
		Limit(limit).
		Find(&blockedUsers).Error

	return blockedUsers, total, err
}

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) ReportUser(reporterID, reportedID uuid.UUID, reason string) error {

	var existingReport model.UserReport
	if err := r.db.Where("reporter_id = ? AND reported_id = ? AND status = ?", reporterID, reportedID, "pending").First(&existingReport).Error; err == nil {

		return nil
	}

	report := model.UserReport{
		ReporterID: reporterID,
		ReportedID: reportedID,
		Reason:     reason,
		Status:     "pending",
	}

	return r.db.Create(&report).Error
}

func (r *ReportRepository) GetUserReports(page, limit int, status string) ([]model.UserReport, int64, error) {
	var reports []model.UserReport
	var total int64

	query := r.db.Model(&model.UserReport{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Preload("Reporter").
		Preload("ReportedUser").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&reports).Error

	return reports, total, err
}

func (r *ReportRepository) ProcessReport(reportID uuid.UUID, approved bool, adminID uuid.UUID, adminNotes string) error {
	status := "rejected"
	if approved {
		status = "approved"
	}

	updates := map[string]interface{}{
		"status":       status,
		"processed_by": adminID,
		"admin_notes":  adminNotes,
	}

	return r.db.Model(&model.UserReport{}).
		Where("id = ?", reportID).
		Updates(updates).Error
}
