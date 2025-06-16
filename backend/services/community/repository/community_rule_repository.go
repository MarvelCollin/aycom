package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/community/model"
)

type CommunityRuleRepository interface {
	Add(rule *model.CommunityRule) error
	Remove(ruleID uuid.UUID) error
	FindByCommunity(communityID uuid.UUID) ([]*model.CommunityRule, error)
	Update(rule *model.CommunityRule) error
}

type GormCommunityRuleRepository struct {
	db *gorm.DB
}

func NewCommunityRuleRepository(db *gorm.DB) CommunityRuleRepository {
	return &GormCommunityRuleRepository{db: db}
}

func (r *GormCommunityRuleRepository) Add(rule *model.CommunityRule) error {
	return r.db.Create(rule).Error
}

func (r *GormCommunityRuleRepository) Remove(ruleID uuid.UUID) error {
	return r.db.Delete(&model.CommunityRule{}, "rule_id = ?", ruleID).Error
}

func (r *GormCommunityRuleRepository) FindByCommunity(communityID uuid.UUID) ([]*model.CommunityRule, error) {
	var rules []*model.CommunityRule
	err := r.db.Where("community_id = ?", communityID).Find(&rules).Error
	return rules, err
}

func (r *GormCommunityRuleRepository) Update(rule *model.CommunityRule) error {
	return r.db.Save(rule).Error
}
