package repository

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/thread/model"
)

type ReplyRepository interface {
	CreateReply(reply *model.Reply) error
	FindReplyByID(id string) (*model.Reply, error)
	FindRepliesByThreadID(threadID string, page, limit int) ([]*model.Reply, error)
	FindRepliesByParentID(parentReplyID string, page, limit int) ([]*model.Reply, error)
	FindRepliesByUserID(userID string, page, limit int) ([]*model.Reply, error)
	UpdateReply(reply *model.Reply) error
	DeleteReply(id string) error
	CountRepliesByParentID(parentID string) (int64, error)
}

type PostgresReplyRepository struct {
	db *gorm.DB
}

func NewReplyRepository(db *gorm.DB) ReplyRepository {
	return &PostgresReplyRepository{db: db}
}

func (r *PostgresReplyRepository) CreateReply(reply *model.Reply) error {
	if reply.ReplyID == uuid.Nil {
		reply.ReplyID = uuid.New()
	}
	return r.db.Create(reply).Error
}

func (r *PostgresReplyRepository) FindReplyByID(id string) (*model.Reply, error) {
	replyID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid UUID format for reply ID")
	}

	var reply model.Reply
	result := r.db.Where("reply_id = ?", replyID).First(&reply)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &reply, nil
}

func (r *PostgresReplyRepository) FindRepliesByThreadID(threadID string, page, limit int) ([]*model.Reply, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return nil, errors.New("invalid UUID format for thread ID")
	}

	var replies []*model.Reply
	offset := (page - 1) * limit
	result := r.db.Where("thread_id = ? AND parent_reply_id IS NULL", threadUUID).
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&replies)

	if result.Error != nil {
		return nil, result.Error
	}
	return replies, nil
}

func (r *PostgresReplyRepository) FindRepliesByParentID(parentReplyID string, page, limit int) ([]*model.Reply, error) {
	log.Printf("Finding replies for parent reply ID: %s (page: %d, limit: %d)", parentReplyID, page, limit)

	parentUUID, err := uuid.Parse(parentReplyID)
	if err != nil {
		log.Printf("Invalid UUID format for parent reply ID: %s - %v", parentReplyID, err)
		return nil, errors.New("invalid UUID format for parent reply ID")
	}

	var parentExists int64
	if err := r.db.Model(&model.Reply{}).Where("reply_id = ? AND deleted_at IS NULL", parentUUID).Count(&parentExists).Error; err != nil {
		log.Printf("Error checking if parent reply exists: %v", err)
		return nil, fmt.Errorf("error verifying parent reply: %w", err)
	}

	if parentExists == 0 {
		log.Printf("Parent reply not found with ID: %s", parentReplyID)
		return nil, gorm.ErrRecordNotFound
	}

	var replies []*model.Reply
	offset := (page - 1) * limit

	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to begin transaction: %v", tx.Error)
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic in FindRepliesByParentID: %v", r)
		}
	}()

	result := tx.Where("parent_reply_id = ? AND deleted_at IS NULL", parentUUID).
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&replies)

	if result.Error != nil {
		tx.Rollback()
		log.Printf("Error finding replies by parent ID: %v", result.Error)
		return nil, result.Error
	}

	var totalCount int64
	if err := tx.Model(&model.Reply{}).Where("parent_reply_id = ? AND deleted_at IS NULL", parentUUID).Count(&totalCount).Error; err != nil {
		tx.Rollback()
		log.Printf("Error counting replies by parent ID: %v", err)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		return nil, err
	}

	log.Printf("Found %d replies for parent reply ID: %s (total: %d)", len(replies), parentReplyID, totalCount)
	return replies, nil
}

func (r *PostgresReplyRepository) FindRepliesByUserID(userID string, page, limit int) ([]*model.Reply, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid UUID format for user ID")
	}

	var replies []*model.Reply
	offset := (page - 1) * limit
	result := r.db.Where("user_id = ?", userUUID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&replies)

	if result.Error != nil {
		return nil, result.Error
	}
	return replies, nil
}

func (r *PostgresReplyRepository) UpdateReply(reply *model.Reply) error {
	return r.db.Save(reply).Error
}

func (r *PostgresReplyRepository) DeleteReply(id string) error {
	replyID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid UUID format for reply ID")
	}

	return r.db.Delete(&model.Reply{}, "reply_id = ?", replyID).Error
}

func (r *PostgresReplyRepository) CountRepliesByParentID(parentID string) (int64, error) {
	log.Printf("Counting replies for parent reply ID: %s", parentID)

	parentUUID, err := uuid.Parse(parentID)
	if err != nil {
		log.Printf("Invalid UUID format for parent reply ID: %s - %v", parentID, err)
		return 0, errors.New("invalid UUID format for parent reply ID")
	}

	var count int64
	err = r.db.Model(&model.Reply{}).
		Where("parent_reply_id = ? AND deleted_at IS NULL", parentUUID).
		Count(&count).Error

	if err != nil {
		log.Printf("Error counting replies for parent ID %s: %v", parentID, err)
		return 0, err
	}

	log.Printf("Found %d replies for parent reply ID: %s", count, parentID)
	return count, nil
}