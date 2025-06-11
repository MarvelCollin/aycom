package repository

import (
	"aycom/backend/services/community/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommunityMemberRepository interface {
	IsMember(communityID, userID uuid.UUID) (bool, error)
	Add(member *model.CommunityMember) error
	Remove(communityID, userID uuid.UUID) error
	FindByID(communityID, userID uuid.UUID) (*model.CommunityMember, error)
	FindByCommunity(communityID uuid.UUID) ([]*model.CommunityMember, error)
	FindByUser(userID uuid.UUID) ([]*model.CommunityMember, error)
	Update(member *model.CommunityMember) error
	CountByCommunity(communityID uuid.UUID) (int64, error)

	// Transaction support
	AddTx(tx *gorm.DB, member *model.CommunityMember) error
	UpdateTx(tx *gorm.DB, member *model.CommunityMember) error
}

type GormCommunityMemberRepository struct {
	db *gorm.DB
}

func NewCommunityMemberRepository(db *gorm.DB) CommunityMemberRepository {
	return &GormCommunityMemberRepository{db: db}
}

func (r *GormCommunityMemberRepository) IsMember(communityID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.CommunityMember{}).
		Where("community_id = ? AND user_id = ?", communityID, userID).
		Count(&count).Error

	return count > 0, err
}

func (r *GormCommunityMemberRepository) Add(member *model.CommunityMember) error {
	// Ensure timestamps are set
	member.CreatedAt = time.Now()
	member.UpdatedAt = time.Now()
	return r.db.Create(member).Error
}

func (r *GormCommunityMemberRepository) Remove(communityID, userID uuid.UUID) error {
	return r.db.Delete(&model.CommunityMember{}, "community_id = ? AND user_id = ?", communityID, userID).Error
}

func (r *GormCommunityMemberRepository) FindByID(communityID, userID uuid.UUID) (*model.CommunityMember, error) {
	var member model.CommunityMember
	err := r.db.Where("community_id = ? AND user_id = ?", communityID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
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
	member.UpdatedAt = time.Now()
	return r.db.Save(member).Error
}

func (r *GormCommunityMemberRepository) CountByCommunity(communityID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.CommunityMember{}).
		Where("community_id = ?", communityID).
		Count(&count).Error
	return count, err
}

// Transaction support
func (r *GormCommunityMemberRepository) AddTx(tx *gorm.DB, member *model.CommunityMember) error {
	// Ensure timestamps are set
	member.CreatedAt = time.Now()
	member.UpdatedAt = time.Now()
	return tx.Create(member).Error
}

func (r *GormCommunityMemberRepository) UpdateTx(tx *gorm.DB, member *model.CommunityMember) error {
	member.UpdatedAt = time.Now()
	return tx.Save(member).Error
}
