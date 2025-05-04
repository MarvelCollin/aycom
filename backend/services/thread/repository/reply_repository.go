package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/thread/model"
)

// ReplyRepository defines the methods for reply-related database operations
type ReplyRepository interface {
	CreateReply(reply *model.Reply) error
	FindReplyByID(id string) (*model.Reply, error)
	FindRepliesByThreadID(threadID string, page, limit int) ([]*model.Reply, error)
	FindRepliesByParentID(parentReplyID string, page, limit int) ([]*model.Reply, error)
	UpdateReply(reply *model.Reply) error
	DeleteReply(id string) error
}

// PostgresReplyRepository is the PostgreSQL implementation of ReplyRepository
type PostgresReplyRepository struct {
	db *gorm.DB
}

// NewReplyRepository creates a new PostgreSQL reply repository
func NewReplyRepository(db *gorm.DB) ReplyRepository {
	return &PostgresReplyRepository{db: db}
}

// CreateReply creates a new reply
func (r *PostgresReplyRepository) CreateReply(reply *model.Reply) error {
	if reply.ReplyID == uuid.Nil {
		reply.ReplyID = uuid.New()
	}
	return r.db.Create(reply).Error
}

// FindReplyByID finds a reply by its ID
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

// FindRepliesByThreadID finds all top-level replies for a specific thread
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
	parentUUID, err := uuid.Parse(parentReplyID)
	if err != nil {
		return nil, errors.New("invalid UUID format for parent reply ID")
	}

	var replies []*model.Reply
	offset := (page - 1) * limit
	result := r.db.Where("parent_reply_id = ?", parentUUID).
		Order("created_at ASC").
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
