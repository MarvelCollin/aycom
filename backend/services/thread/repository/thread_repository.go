package repository

import (
	"errors"
	"fmt"

	"aycom/backend/services/thread/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ThreadRepository defines the methods for thread-related database operations
type ThreadRepository interface {
	// Thread methods
	CreateThread(thread *model.Thread) error
	FindThreadByID(id string) (*model.Thread, error)
	FindThreadsByUserID(userID string, page, limit int) ([]*model.Thread, error)
	FindAllThreads(page, limit int) ([]*model.Thread, error)
	UpdateThread(thread *model.Thread) error
	DeleteThread(id string) error
	ThreadExists(threadID string) (bool, error)
}

// PostgresThreadRepository is the PostgreSQL implementation of ThreadRepository
type PostgresThreadRepository struct {
	db *gorm.DB
}

// NewThreadRepository creates a new PostgreSQL thread repository
func NewThreadRepository(db *gorm.DB) ThreadRepository {
	return &PostgresThreadRepository{db: db}
}

// CreateThread creates a new thread
func (r *PostgresThreadRepository) CreateThread(thread *model.Thread) error {
	if thread.ThreadID == uuid.Nil {
		thread.ThreadID = uuid.New()
	}
	return r.db.Create(thread).Error
}

// FindThreadByID finds a thread by its ID
func (r *PostgresThreadRepository) FindThreadByID(id string) (*model.Thread, error) {
	threadID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid UUID format for thread ID")
	}

	var thread model.Thread
	result := r.db.Where("thread_id = ?", threadID).First(&thread)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &thread, nil
}

// FindThreadsByUserID finds all threads by a specific user ID
func (r *PostgresThreadRepository) FindThreadsByUserID(userID string, page, limit int) ([]*model.Thread, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid UUID format for user ID")
	}

	var threads []*model.Thread
	offset := (page - 1) * limit
	result := r.db.Where("user_id = ?", userUUID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads)

	if result.Error != nil {
		return nil, result.Error
	}
	return threads, nil
}

// FindAllThreads finds all threads with pagination
func (r *PostgresThreadRepository) FindAllThreads(page, limit int) ([]*model.Thread, error) {
	var threads []*model.Thread
	offset := (page - 1) * limit
	result := r.db.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads)

	if result.Error != nil {
		return nil, result.Error
	}
	return threads, nil
}

// UpdateThread updates an existing thread
func (r *PostgresThreadRepository) UpdateThread(thread *model.Thread) error {
	return r.db.Save(thread).Error
}

// DeleteThread deletes a thread by its ID
func (r *PostgresThreadRepository) DeleteThread(id string) error {
	threadID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid UUID format for thread ID")
	}

	return r.db.Delete(&model.Thread{}, "thread_id = ?", threadID).Error
}

// ThreadExists checks if a thread exists by ID
func (r *PostgresThreadRepository) ThreadExists(threadID string) (bool, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return false, fmt.Errorf("invalid thread ID format: %w", err)
	}

	var count int64
	result := r.db.Model(&model.Thread{}).
		Where("thread_id = ?", threadUUID).
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("error checking thread existence: %w", result.Error)
	}

	return count > 0, nil
}
