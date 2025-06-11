package repository

import (
	"errors"
	"fmt"

	"aycom/backend/services/thread/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ThreadRepository interface {
	CreateThread(thread *model.Thread) error
	FindThreadByID(id string) (*model.Thread, error)
	FindThreadsByUserID(userID string, page, limit int) ([]*model.Thread, error)
	FindAllThreads(page, limit int) ([]*model.Thread, error)
	CountAllThreads() (int64, error)
	UpdateThread(thread *model.Thread) error
	DeleteThread(id string) error
	ThreadExists(threadID string) (bool, error)
	RunInTransaction(fn func(tx *gorm.DB) error) error
}

type PostgresThreadRepository struct {
	db *gorm.DB
}

func NewThreadRepository(db *gorm.DB) ThreadRepository {
	return &PostgresThreadRepository{db: db}
}

// RunInTransaction executes the given function within a database transaction
func (r *PostgresThreadRepository) RunInTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *PostgresThreadRepository) CreateThread(thread *model.Thread) error {
	if thread.ThreadID == uuid.Nil {
		thread.ThreadID = uuid.New()
	}
	return r.db.Create(thread).Error
}

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

func (r *PostgresThreadRepository) FindThreadsByUserID(userID string, page, limit int) ([]*model.Thread, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format for user ID: %w", err)
	}

	var threads []*model.Thread
	offset := (page - 1) * limit

	// Add debug logging
	fmt.Printf("Executing FindThreadsByUserID query: userID=%s, offset=%d, limit=%d\n", userUUID, offset, limit)

	result := r.db.Where("user_id = ?", userUUID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads)

	// Check if there was an actual database error
	// Note: If no records were found, result.Error will be nil
	if result.Error != nil {
		return nil, fmt.Errorf("database error in FindThreadsByUserID: %w", result.Error)
	}

	// If no threads found, return an empty slice (not an error)
	fmt.Printf("Found %d threads for user %s\n", len(threads), userID)
	return threads, nil
}

func (r *PostgresThreadRepository) FindAllThreads(page, limit int) ([]*model.Thread, error) {
	var threads []*model.Thread
	offset := (page - 1) * limit

	// Use a more efficient query with proper ordering
	// and ensure we're not fetching soft-deleted threads
	result := r.db.
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads)

	if result.Error != nil {
		return nil, result.Error
	}

	return threads, nil
}

func (r *PostgresThreadRepository) UpdateThread(thread *model.Thread) error {
	return r.db.Save(thread).Error
}

func (r *PostgresThreadRepository) DeleteThread(id string) error {
	threadID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid UUID format for thread ID")
	}

	return r.db.Delete(&model.Thread{}, "thread_id = ?", threadID).Error
}

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

func (r *PostgresThreadRepository) CountAllThreads() (int64, error) {
	var count int64
	result := r.db.Model(&model.Thread{}).Where("deleted_at IS NULL").Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}
