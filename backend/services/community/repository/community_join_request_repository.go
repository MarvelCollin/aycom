package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/community/model"
)

type CommunityJoinRequestRepository interface {
	Add(request *model.CommunityJoinRequest) error
	Remove(requestID uuid.UUID) error
	FindByID(requestID uuid.UUID) (*model.CommunityJoinRequest, error)
	FindByCommunity(communityID uuid.UUID) ([]*model.CommunityJoinRequest, error)
	FindByUser(userID uuid.UUID) ([]*model.CommunityJoinRequest, error)
	Update(request *model.CommunityJoinRequest) error
	HasPendingJoinRequest(communityID, userID uuid.UUID) (bool, error)

	BeginTx(ctx context.Context) (*gorm.DB, error)
	UpdateTx(tx *gorm.DB, request *model.CommunityJoinRequest) error
}

type GormCommunityJoinRequestRepository struct {
	db *gorm.DB
}

func NewCommunityJoinRequestRepository(db *gorm.DB) CommunityJoinRequestRepository {
	return &GormCommunityJoinRequestRepository{db: db}
}

func (r *GormCommunityJoinRequestRepository) Add(request *model.CommunityJoinRequest) error {
	return r.db.Create(request).Error
}

func (r *GormCommunityJoinRequestRepository) Remove(requestID uuid.UUID) error {
	return r.db.Delete(&model.CommunityJoinRequest{}, "request_id = ?", requestID).Error
}

func (r *GormCommunityJoinRequestRepository) FindByID(requestID uuid.UUID) (*model.CommunityJoinRequest, error) {
	var request model.CommunityJoinRequest
	err := r.db.Where("request_id = ?", requestID).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *GormCommunityJoinRequestRepository) FindByCommunity(communityID uuid.UUID) ([]*model.CommunityJoinRequest, error) {
	var requests []*model.CommunityJoinRequest
	err := r.db.Where("community_id = ? AND status = ?", communityID, "pending").Find(&requests).Error
	return requests, err
}

func (r *GormCommunityJoinRequestRepository) FindByUser(userID uuid.UUID) ([]*model.CommunityJoinRequest, error) {
	var requests []*model.CommunityJoinRequest
	err := r.db.Where("user_id = ?", userID).Find(&requests).Error
	return requests, err
}

func (r *GormCommunityJoinRequestRepository) Update(request *model.CommunityJoinRequest) error {
	return r.db.Save(request).Error
}

func (r *GormCommunityJoinRequestRepository) HasPendingJoinRequest(communityID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.CommunityJoinRequest{}).
		Where("community_id = ? AND user_id = ? AND status = ?", communityID, userID, "pending").
		Count(&count).Error

	return count > 0, err
}

func (r *GormCommunityJoinRequestRepository) BeginTx(ctx context.Context) (*gorm.DB, error) {
	tx := r.db.WithContext(ctx).Begin()
	return tx, tx.Error
}

func (r *GormCommunityJoinRequestRepository) UpdateTx(tx *gorm.DB, request *model.CommunityJoinRequest) error {
	return tx.Save(request).Error
}
