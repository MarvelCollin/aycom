package repository

import (
	"errors"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Define error types
var (
	ErrThreadNotFound = errors.New("thread not found")
	ErrReplyNotFound  = errors.New("reply not found")
	ErrPollNotFound   = errors.New("poll not found")
)

// Implementation of ThreadRepository
type threadRepositoryImpl struct {
	db *gorm.DB
}

// CreateThread implements ThreadRepository
func (r *threadRepositoryImpl) CreateThread(thread *model.Thread) error {
	return r.db.Create(thread).Error
}

// FindThreadByID implements ThreadRepository
func (r *threadRepositoryImpl) FindThreadByID(id string) (*model.Thread, error) {
	var thread model.Thread

	err := r.db.
		Preload("Media").
		First(&thread, "thread_id = ?", id).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThreadNotFound
		}
		return nil, err
	}

	return &thread, nil
}

// FindThreadsByUserID implements ThreadRepository
func (r *threadRepositoryImpl) FindThreadsByUserID(userID string, page, limit int) ([]*model.Thread, int64, error) {
	var threads []*model.Thread
	var totalCount int64

	offset := (page - 1) * limit

	// Get total count
	if err := r.db.Model(&model.Thread{}).
		Where("user_id = ?", userID).
		Count(&totalCount).
		Error; err != nil {
		return nil, 0, err
	}

	// Get threads with pagination
	err := r.db.
		Preload("Media").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads).
		Error

	if err != nil {
		return nil, 0, err
	}

	return threads, totalCount, nil
}

// FindThreadsByCommunityID implements ThreadRepository
func (r *threadRepositoryImpl) FindThreadsByCommunityID(communityID string, page, limit int) ([]*model.Thread, int64, error) {
	var threads []*model.Thread
	var totalCount int64

	offset := (page - 1) * limit

	// Get total count
	if err := r.db.Model(&model.Thread{}).
		Where("community_id = ?", communityID).
		Count(&totalCount).
		Error; err != nil {
		return nil, 0, err
	}

	// Get threads with pagination
	err := r.db.
		Preload("Media").
		Where("community_id = ?", communityID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads).
		Error

	if err != nil {
		return nil, 0, err
	}

	return threads, totalCount, nil
}

// UpdateThread implements ThreadRepository
func (r *threadRepositoryImpl) UpdateThread(thread *model.Thread) error {
	return r.db.Save(thread).Error
}

// DeleteThread implements ThreadRepository
func (r *threadRepositoryImpl) DeleteThread(id string) error {
	return r.db.Delete(&model.Thread{}, "thread_id = ?", id).Error
}

// GetThreadStats implements ThreadRepository
func (r *threadRepositoryImpl) GetThreadStats(threadID string) (replyCount, likeCount, repostCount int64, err error) {
	// Get reply count
	if err := r.db.Model(&model.Reply{}).
		Where("thread_id = ?", threadID).
		Count(&replyCount).
		Error; err != nil {
		return 0, 0, 0, err
	}

	// Get like count
	if err := r.db.Model(&model.Like{}).
		Where("thread_id = ?", threadID).
		Count(&likeCount).
		Error; err != nil {
		return 0, 0, 0, err
	}

	// Get repost count
	if err := r.db.Model(&model.Repost{}).
		Where("thread_id = ?", threadID).
		Count(&repostCount).
		Error; err != nil {
		return 0, 0, 0, err
	}

	return replyCount, likeCount, repostCount, nil
}

// IncrementViewCount implements ThreadRepository
func (r *threadRepositoryImpl) IncrementViewCount(threadID string) error {
	// Implementation for view count would typically use Redis or a similar cache
	// For simplicity, we'll just stub this method for now
	return nil
}

// mediaRepositoryImpl implements MediaRepository
type mediaRepositoryImpl struct {
	db *gorm.DB
}

// CreateMedia implements MediaRepository
func (r *mediaRepositoryImpl) CreateMedia(media *model.Media) error {
	return r.db.Create(media).Error
}

// FindMediaByID implements MediaRepository
func (r *mediaRepositoryImpl) FindMediaByID(id string) (*model.Media, error) {
	var media model.Media
	err := r.db.First(&media, "media_id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("media not found")
		}
		return nil, err
	}
	return &media, nil
}

// FindMediaByThreadID implements MediaRepository
func (r *mediaRepositoryImpl) FindMediaByThreadID(threadID string) ([]*model.Media, error) {
	var mediaList []*model.Media
	err := r.db.Where("thread_id = ?", threadID).Find(&mediaList).Error
	if err != nil {
		return nil, err
	}
	return mediaList, nil
}

// FindMediaByReplyID implements MediaRepository
func (r *mediaRepositoryImpl) FindMediaByReplyID(replyID string) ([]*model.Media, error) {
	var mediaList []*model.Media
	err := r.db.Where("reply_id = ?", replyID).Find(&mediaList).Error
	if err != nil {
		return nil, err
	}
	return mediaList, nil
}

// DeleteMedia implements MediaRepository
func (r *mediaRepositoryImpl) DeleteMedia(id string) error {
	return r.db.Delete(&model.Media{}, "media_id = ?", id).Error
}

// hashtagRepositoryImpl implements HashtagRepository
type hashtagRepositoryImpl struct {
	db *gorm.DB
}

// CreateHashtag implements HashtagRepository
func (r *hashtagRepositoryImpl) CreateHashtag(hashtag *model.Hashtag) error {
	return r.db.Create(hashtag).Error
}

// FindOrCreateHashtagByText implements HashtagRepository
func (r *hashtagRepositoryImpl) FindOrCreateHashtagByText(text string) (*model.Hashtag, error) {
	var hashtag model.Hashtag

	// Try to find existing hashtag
	err := r.db.Where("text = ?", text).First(&hashtag).Error
	if err == nil {
		return &hashtag, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new hashtag
	newHashtag := &model.Hashtag{
		HashtagID: uuid.New(),
		Text:      text,
		CreatedAt: time.Now(),
	}

	if err := r.db.Create(newHashtag).Error; err != nil {
		return nil, err
	}

	return newHashtag, nil
}

// FindHashtagByID implements HashtagRepository
func (r *hashtagRepositoryImpl) FindHashtagByID(id string) (*model.Hashtag, error) {
	var hashtag model.Hashtag
	err := r.db.First(&hashtag, "hashtag_id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("hashtag not found")
		}
		return nil, err
	}
	return &hashtag, nil
}

// FindHashtagByText implements HashtagRepository
func (r *hashtagRepositoryImpl) FindHashtagByText(text string) (*model.Hashtag, error) {
	var hashtag model.Hashtag
	err := r.db.Where("text = ?", text).First(&hashtag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("hashtag not found")
		}
		return nil, err
	}
	return &hashtag, nil
}

// AddHashtagToThread implements HashtagRepository
func (r *hashtagRepositoryImpl) AddHashtagToThread(threadID, hashtagID string) error {
	threadHashtag := &model.ThreadHashtag{
		ThreadID:  uuid.MustParse(threadID),
		HashtagID: uuid.MustParse(hashtagID),
	}
	return r.db.Create(threadHashtag).Error
}

// RemoveHashtagFromThread implements HashtagRepository
func (r *hashtagRepositoryImpl) RemoveHashtagFromThread(threadID, hashtagID string) error {
	return r.db.Where("thread_id = ? AND hashtag_id = ?", threadID, hashtagID).Delete(&model.ThreadHashtag{}).Error
}

// FindHashtagsByThreadID implements HashtagRepository
func (r *hashtagRepositoryImpl) FindHashtagsByThreadID(threadID string) ([]*model.Hashtag, error) {
	var hashtags []*model.Hashtag

	err := r.db.Table("hashtags").
		Joins("JOIN thread_hashtags ON hashtags.hashtag_id = thread_hashtags.hashtag_id").
		Where("thread_hashtags.thread_id = ?", threadID).
		Find(&hashtags).Error

	if err != nil {
		return nil, err
	}

	return hashtags, nil
}

// FindTrendingHashtags implements HashtagRepository
func (r *hashtagRepositoryImpl) FindTrendingHashtags(limit int) ([]*model.Hashtag, int64, error) {
	var hashtags []*model.Hashtag
	var totalCount int64

	// Get total count
	if err := r.db.Model(&model.Hashtag{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// This is a simplified implementation - in a real system you would
	// count hashtag usage within a time period (e.g., last 24 hours)
	err := r.db.Table("hashtags").
		Select("hashtags.*, COUNT(thread_hashtags.hashtag_id) as usage_count").
		Joins("LEFT JOIN thread_hashtags ON hashtags.hashtag_id = thread_hashtags.hashtag_id").
		Group("hashtags.hashtag_id").
		Order("usage_count DESC").
		Limit(limit).
		Find(&hashtags).Error

	if err != nil {
		return nil, 0, err
	}

	return hashtags, totalCount, nil
}

// mentionRepositoryImpl implements MentionRepository
type mentionRepositoryImpl struct {
	db *gorm.DB
}

// CreateMention implements MentionRepository
func (r *mentionRepositoryImpl) CreateMention(mention *model.UserMention) error {
	return r.db.Create(mention).Error
}

// FindMentionsByThreadID implements MentionRepository
func (r *mentionRepositoryImpl) FindMentionsByThreadID(threadID string) ([]*model.UserMention, error) {
	var mentions []*model.UserMention
	err := r.db.Where("thread_id = ?", threadID).Find(&mentions).Error
	if err != nil {
		return nil, err
	}
	return mentions, nil
}

// FindMentionsByReplyID implements MentionRepository
func (r *mentionRepositoryImpl) FindMentionsByReplyID(replyID string) ([]*model.UserMention, error) {
	var mentions []*model.UserMention
	err := r.db.Where("reply_id = ?", replyID).Find(&mentions).Error
	if err != nil {
		return nil, err
	}
	return mentions, nil
}

// DeleteMention implements MentionRepository
func (r *mentionRepositoryImpl) DeleteMention(id string) error {
	return r.db.Delete(&model.UserMention{}, "mention_id = ?", id).Error
}
