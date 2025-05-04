package repository

import (
	"aycom/backend/services/community/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommunityMemberRepository interface {
	Add(member *model.CommunityMember) error
	Remove(communityID, userID uuid.UUID) error
	FindByCommunity(communityID uuid.UUID) ([]*model.CommunityMember, error)
	FindByUser(userID uuid.UUID) ([]*model.CommunityMember, error)
	Update(member *model.CommunityMember) error
}

type GormCommunityMemberRepository struct {
	db *gorm.DB
}

func NewCommunityMemberRepository(db *gorm.DB) CommunityMemberRepository {
	return &GormCommunityMemberRepository{db: db}
}

func (r *GormCommunityMemberRepository) Add(member *model.CommunityMember) error {
	return r.db.Create(member).Error
}

func (r *GormCommunityMemberRepository) Remove(communityID, userID uuid.UUID) error {
	return r.db.Delete(&model.CommunityMember{}, "community_id = ? AND user_id = ?", communityID, userID).Error
}

func (r *GormCommunityMemberRepository) FindByCommunity(communityID uuid.UUID) ([]*model.CommunityMember, error) {
	var members []*model.CommunityMember
	err := r.db.Where("community_id = ?", communityID).Find(&members).Error
	return members, err
}

func (r *GormCommunityMemberRepository) FindByUser(userID uuid.UUID) ([]*model.CommunityMember, error) {
	var members []*model.CommunityMember
	err := r.db.Where("user_id = ?", userID).Find(&members).Error
	return members, err
}

func (r *GormCommunityMemberRepository) Update(member *model.CommunityMember) error {
	return r.db.Save(member).Error
}
