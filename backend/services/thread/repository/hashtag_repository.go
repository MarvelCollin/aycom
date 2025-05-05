package repository

import (
	"errors"

	"aycom/backend/services/thread/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HashtagRepository defines the methods for hashtag-related database operations
type HashtagRepository interface {
	CreateHashtag(hashtag *model.Hashtag) error
	FindHashtagByText(text string) (*model.Hashtag, error)
	FindHashtagsByThreadID(threadID string) ([]*model.Hashtag, error)
	AddHashtagToThread(threadID string, hashtagID string) error
	RemoveHashtagFromThread(threadID string, hashtagID string) error
	GetTrendingHashtags(limit int) ([]*model.Hashtag, error)
	CountThreadsWithHashtag(hashtagID string) (int, error)
}

// PostgresHashtagRepository is the PostgreSQL implementation of HashtagRepository
type PostgresHashtagRepository struct {
	db *gorm.DB
}

// NewHashtagRepository creates a new PostgreSQL hashtag repository
func NewHashtagRepository(db *gorm.DB) HashtagRepository {
	return &PostgresHashtagRepository{db: db}
}

// CreateHashtag creates a new hashtag if it doesn't exist
func (r *PostgresHashtagRepository) CreateHashtag(hashtag *model.Hashtag) error {
	if hashtag.HashtagID == uuid.Nil {
		hashtag.HashtagID = uuid.New()
	}

	var existingHashtag model.Hashtag
	err := r.db.Where("text = ?", hashtag.Text).First(&existingHashtag).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new hashtag
			return r.db.Create(hashtag).Error
		}
		return err
	}

	// Hashtag already exists, return the existing one
	*hashtag = existingHashtag
	return nil
}

// FindHashtagByText finds a hashtag by its text
func (r *PostgresHashtagRepository) FindHashtagByText(text string) (*model.Hashtag, error) {
	var hashtag model.Hashtag
	result := r.db.Where("text = ?", text).First(&hashtag)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &hashtag, nil
}

// FindHashtagsByThreadID finds all hashtags for a specific thread
func (r *PostgresHashtagRepository) FindHashtagsByThreadID(threadID string) ([]*model.Hashtag, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return nil, errors.New("invalid UUID format for thread ID")
	}

	var hashtags []*model.Hashtag
	result := r.db.Table("hashtags").
		Joins("JOIN thread_hashtags ON hashtags.hashtag_id = thread_hashtags.hashtag_id").
		Where("thread_hashtags.thread_id = ?", threadUUID).
		Find(&hashtags)

	if result.Error != nil {
		return nil, result.Error
	}
	return hashtags, nil
}

// AddHashtagToThread adds a hashtag to a thread
func (r *PostgresHashtagRepository) AddHashtagToThread(threadID string, hashtagID string) error {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return errors.New("invalid UUID format for thread ID")
	}

	hashtagUUID, err := uuid.Parse(hashtagID)
	if err != nil {
		return errors.New("invalid UUID format for hashtag ID")
	}

	threadHashtag := model.ThreadHashtag{
		ThreadID:  threadUUID,
		HashtagID: hashtagUUID,
	}

	return r.db.Create(&threadHashtag).Error
}

// RemoveHashtagFromThread removes a hashtag from a thread
func (r *PostgresHashtagRepository) RemoveHashtagFromThread(threadID string, hashtagID string) error {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return errors.New("invalid UUID format for thread ID")
	}

	hashtagUUID, err := uuid.Parse(hashtagID)
	if err != nil {
		return errors.New("invalid UUID format for hashtag ID")
	}

	return r.db.Where("thread_id = ? AND hashtag_id = ?", threadUUID, hashtagUUID).Delete(&model.ThreadHashtag{}).Error
}

// GetTrendingHashtags gets the most used hashtags
func (r *PostgresHashtagRepository) GetTrendingHashtags(limit int) ([]*model.Hashtag, error) {
	var hashtags []*model.Hashtag

	result := r.db.Table("hashtags").
		Joins("JOIN thread_hashtags ON hashtags.hashtag_id = thread_hashtags.hashtag_id").
		Select("hashtags.*, COUNT(thread_hashtags.thread_id) as thread_count").
		Group("hashtags.hashtag_id").
		Order("thread_count DESC").
		Limit(limit).
		Find(&hashtags)

	if result.Error != nil {
		return nil, result.Error
	}

	return hashtags, nil
}

// CountThreadsWithHashtag returns the number of threads that use this hashtag
func (r *PostgresHashtagRepository) CountThreadsWithHashtag(hashtagID string) (int, error) {
	hashtagUUID, err := uuid.Parse(hashtagID)
	if err != nil {
		return 0, errors.New("invalid UUID format for hashtag ID")
	}

	var count int64
	result := r.db.Model(&model.ThreadHashtag{}).
		Where("hashtag_id = ?", hashtagUUID).
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return int(count), nil
}
