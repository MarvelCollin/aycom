package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/thread/model"
)

// MediaRepository defines the methods for media-related database operations
type MediaRepository interface {
	CreateMedia(media *model.Media) error
	FindMediaByID(id string) (*model.Media, error)
	FindMediaByThreadID(threadID string) ([]*model.Media, error)
	FindMediaByReplyID(replyID string) ([]*model.Media, error)
	FindMediaByUserID(userID string, page, limit int) ([]*model.Media, error)
	DeleteMedia(id string) error
}

// PostgresMediaRepository is the PostgreSQL implementation of MediaRepository
type PostgresMediaRepository struct {
	db *gorm.DB
}

// NewMediaRepository creates a new PostgreSQL media repository
func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &PostgresMediaRepository{db: db}
}

// CreateMedia creates a new media record
func (r *PostgresMediaRepository) CreateMedia(media *model.Media) error {
	if media.MediaID == uuid.Nil {
		media.MediaID = uuid.New()
	}
	return r.db.Create(media).Error
}

// FindMediaByID finds a media record by its ID
func (r *PostgresMediaRepository) FindMediaByID(id string) (*model.Media, error) {
	mediaID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid UUID format for media ID")
	}

	var media model.Media
	result := r.db.Where("media_id = ?", mediaID).First(&media)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &media, nil
}

func (r *PostgresMediaRepository) FindMediaByThreadID(threadID string) ([]*model.Media, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return nil, errors.New("invalid UUID format for thread ID")
	}

	var media []*model.Media
	result := r.db.Where("thread_id = ?", threadUUID).Find(&media)
	if result.Error != nil {
		return nil, result.Error
	}
	return media, nil
}

func (r *PostgresMediaRepository) FindMediaByReplyID(replyID string) ([]*model.Media, error) {
	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return nil, errors.New("invalid UUID format for reply ID")
	}

	var media []*model.Media
	result := r.db.Where("reply_id = ?", replyUUID).Find(&media)
	if result.Error != nil {
		return nil, result.Error
	}
	return media, nil
}

func (r *PostgresMediaRepository) FindMediaByUserID(userID string, page, limit int) ([]*model.Media, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid UUID format for user ID")
	}

	var media []*model.Media
	offset := (page - 1) * limit

	// Use a join query to get media for both threads and replies created by the user
	result := r.db.Raw(`
		SELECT m.* FROM media m
		LEFT JOIN threads t ON m.thread_id = t.thread_id
		LEFT JOIN replies r ON m.reply_id = r.reply_id
		WHERE (t.user_id = ? OR r.user_id = ?)
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, userUUID, userUUID, limit, offset).Scan(&media)

	if result.Error != nil {
		return nil, result.Error
	}
	return media, nil
}

func (r *PostgresMediaRepository) DeleteMedia(id string) error {
	mediaID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid UUID format for media ID")
	}

	return r.db.Delete(&model.Media{}, "media_id = ?", mediaID).Error
}
