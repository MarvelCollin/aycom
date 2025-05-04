package repository

import (
	"aycom/backend/services/community/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommunityJoinRequestRepository interface {
	Add(request *model.CommunityJoinRequest) error
	Remove(requestID uuid.UUID) error
	FindByCommunity(communityID uuid.UUID) ([]*model.CommunityJoinRequest, error)
	FindByUser(userID uuid.UUID) ([]*model.CommunityJoinRequest, error)
	Update(request *model.CommunityJoinRequest) error
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

func (r *GormCommunityJoinRequestRepository) FindByCommunity(communityID uuid.UUID) ([]*model.CommunityJoinRequest, error) {
	var requests []*model.CommunityJoinRequest
	err := r.db.Where("community_id = ?", communityID).Find(&requests).Error
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
