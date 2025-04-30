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

// replyRepositoryImpl implements ReplyRepository
type replyRepositoryImpl struct {
	db *gorm.DB
}

// CreateReply implements ReplyRepository
func (r *replyRepositoryImpl) CreateReply(reply *model.Reply) error {
	return r.db.Create(reply).Error
}

// FindReplyByID implements ReplyRepository
func (r *replyRepositoryImpl) FindReplyByID(id string) (*model.Reply, error) {
	var reply model.Reply
	err := r.db.First(&reply, "reply_id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReplyNotFound
		}
		return nil, err
	}
	return &reply, nil
}

// FindRepliesByThreadID implements ReplyRepository
func (r *replyRepositoryImpl) FindRepliesByThreadID(threadID string, parentReplyID *string, page, limit int) ([]*model.Reply, int64, error) {
	var replies []*model.Reply
	var totalCount int64

	offset := (page - 1) * limit
	query := r.db.Model(&model.Reply{}).Where("thread_id = ?", threadID)

	if parentReplyID != nil {
		query = query.Where("parent_reply_id = ?", *parentReplyID)
	} else {
		query = query.Where("parent_reply_id IS NULL")
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&replies).Error; err != nil {
		return nil, 0, err
	}

	return replies, totalCount, nil
}

// UpdateReply implements ReplyRepository
func (r *replyRepositoryImpl) UpdateReply(reply *model.Reply) error {
	return r.db.Save(reply).Error
}

// DeleteReply implements ReplyRepository
func (r *replyRepositoryImpl) DeleteReply(id string) error {
	return r.db.Delete(&model.Reply{}, "reply_id = ?", id).Error
}

// GetReplyStats implements ReplyRepository
func (r *replyRepositoryImpl) GetReplyStats(replyID string) (replyCount, likeCount int64, err error) {
	// Count replies to this reply
	if err := r.db.Model(&model.Reply{}).Where("parent_reply_id = ?", replyID).Count(&replyCount).Error; err != nil {
		return 0, 0, err
	}

	// Count likes on this reply
	if err := r.db.Model(&model.Like{}).Where("reply_id = ?", replyID).Count(&likeCount).Error; err != nil {
		return 0, 0, err
	}

	return replyCount, likeCount, nil
}

// likeRepositoryImpl implements LikeRepository
type likeRepositoryImpl struct {
	db *gorm.DB
}

// CreateThreadLike implements LikeRepository
func (r *likeRepositoryImpl) CreateThreadLike(userID, threadID string) error {
	like := &model.Like{
		UserID:    uuid.MustParse(userID),
		ThreadID:  &uuid.UUID{},
		CreatedAt: time.Now(),
	}
	threadUUID := uuid.MustParse(threadID)
	like.ThreadID = &threadUUID
	return r.db.Create(like).Error
}

// DeleteThreadLike implements LikeRepository
func (r *likeRepositoryImpl) DeleteThreadLike(userID, threadID string) error {
	return r.db.Where("user_id = ? AND thread_id = ?", userID, threadID).Delete(&model.Like{}).Error
}

// CreateReplyLike implements LikeRepository
func (r *likeRepositoryImpl) CreateReplyLike(userID, replyID string) error {
	like := &model.Like{
		UserID:    uuid.MustParse(userID),
		ReplyID:   &uuid.UUID{},
		CreatedAt: time.Now(),
	}
	replyUUID := uuid.MustParse(replyID)
	like.ReplyID = &replyUUID
	return r.db.Create(like).Error
}

// DeleteReplyLike implements LikeRepository
func (r *likeRepositoryImpl) DeleteReplyLike(userID, replyID string) error {
	return r.db.Where("user_id = ? AND reply_id = ?", userID, replyID).Delete(&model.Like{}).Error
}

// HasUserLikedThread implements LikeRepository
func (r *likeRepositoryImpl) HasUserLikedThread(userID, threadID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Like{}).Where("user_id = ? AND thread_id = ?", userID, threadID).Count(&count).Error
	return count > 0, err
}

// HasUserLikedReply implements LikeRepository
func (r *likeRepositoryImpl) HasUserLikedReply(userID, replyID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Like{}).Where("user_id = ? AND reply_id = ?", userID, replyID).Count(&count).Error
	return count > 0, err
}

// GetThreadLikeCount implements LikeRepository
func (r *likeRepositoryImpl) GetThreadLikeCount(threadID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Like{}).Where("thread_id = ?", threadID).Count(&count).Error
	return count, err
}

// GetReplyLikeCount implements LikeRepository
func (r *likeRepositoryImpl) GetReplyLikeCount(replyID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Like{}).Where("reply_id = ?", replyID).Count(&count).Error
	return count, err
}

// repostRepositoryImpl implements RepostRepository
type repostRepositoryImpl struct {
	db *gorm.DB
}

// CreateRepost implements RepostRepository
func (r *repostRepositoryImpl) CreateRepost(userID, threadID string, repostText *string) error {
	repost := &model.Repost{
		UserID:     uuid.MustParse(userID),
		ThreadID:   uuid.MustParse(threadID),
		RepostText: repostText,
		CreatedAt:  time.Now(),
	}
	return r.db.Create(repost).Error
}

// DeleteRepost implements RepostRepository
func (r *repostRepositoryImpl) DeleteRepost(userID, threadID string) error {
	return r.db.Where("user_id = ? AND thread_id = ?", userID, threadID).Delete(&model.Repost{}).Error
}

// HasUserReposted implements RepostRepository
func (r *repostRepositoryImpl) HasUserReposted(userID, threadID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Repost{}).Where("user_id = ? AND thread_id = ?", userID, threadID).Count(&count).Error
	return count > 0, err
}

// GetRepostCount implements RepostRepository
func (r *repostRepositoryImpl) GetRepostCount(threadID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Repost{}).Where("thread_id = ?", threadID).Count(&count).Error
	return count, err
}

// bookmarkRepositoryImpl implements BookmarkRepository
type bookmarkRepositoryImpl struct {
	db *gorm.DB
}

// CreateBookmark implements BookmarkRepository
func (r *bookmarkRepositoryImpl) CreateBookmark(userID, threadID string) error {
	bookmark := &model.Bookmark{
		UserID:    uuid.MustParse(userID),
		ThreadID:  uuid.MustParse(threadID),
		CreatedAt: time.Now(),
	}
	return r.db.Create(bookmark).Error
}

// DeleteBookmark implements BookmarkRepository
func (r *bookmarkRepositoryImpl) DeleteBookmark(userID, threadID string) error {
	return r.db.Where("user_id = ? AND thread_id = ?", userID, threadID).Delete(&model.Bookmark{}).Error
}

// HasUserBookmarked implements BookmarkRepository
func (r *bookmarkRepositoryImpl) HasUserBookmarked(userID, threadID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Bookmark{}).Where("user_id = ? AND thread_id = ?", userID, threadID).Count(&count).Error
	return count > 0, err
}

// GetUserBookmarks implements BookmarkRepository
func (r *bookmarkRepositoryImpl) GetUserBookmarks(userID string, page, limit int) ([]*model.Thread, int64, error) {
	var threads []*model.Thread
	var totalCount int64

	offset := (page - 1) * limit

	if err := r.db.Model(&model.Bookmark{}).Where("user_id = ?", userID).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Table("threads").
		Joins("JOIN bookmarks ON threads.thread_id = bookmarks.thread_id").
		Where("bookmarks.user_id = ?", userID).
		Order("bookmarks.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads).Error

	if err != nil {
		return nil, 0, err
	}

	return threads, totalCount, nil
}

// categoryRepositoryImpl implements CategoryRepository
type categoryRepositoryImpl struct {
	db *gorm.DB
}

// CreateCategory implements CategoryRepository
func (r *categoryRepositoryImpl) CreateCategory(category *model.Category) error {
	return r.db.Create(category).Error
}

// FindOrCreateCategoryByName implements CategoryRepository
func (r *categoryRepositoryImpl) FindOrCreateCategoryByName(name string, categoryType string) (*model.Category, error) {
	var category model.Category

	err := r.db.Where("name = ? AND type = ?", name, categoryType).First(&category).Error
	if err == nil {
		return &category, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	newCategory := &model.Category{
		CategoryID: uuid.New(),
		Name:       name,
		Type:       categoryType,
		CreatedAt:  time.Now(),
	}

	if err := r.db.Create(newCategory).Error; err != nil {
		return nil, err
	}

	return newCategory, nil
}

// FindCategoryByID implements CategoryRepository
func (r *categoryRepositoryImpl) FindCategoryByID(id string) (*model.Category, error) {
	var category model.Category
	err := r.db.First(&category, "category_id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

// FindCategoryByName implements CategoryRepository
func (r *categoryRepositoryImpl) FindCategoryByName(name string, categoryType string) (*model.Category, error) {
	var category model.Category
	err := r.db.Where("name = ? AND type = ?", name, categoryType).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

// AddCategoryToThread implements CategoryRepository
func (r *categoryRepositoryImpl) AddCategoryToThread(threadID, categoryID string) error {
	threadCategory := &model.ThreadCategory{
		ThreadID:   uuid.MustParse(threadID),
		CategoryID: uuid.MustParse(categoryID),
	}
	return r.db.Create(threadCategory).Error
}

// RemoveCategoryFromThread implements CategoryRepository
func (r *categoryRepositoryImpl) RemoveCategoryFromThread(threadID, categoryID string) error {
	return r.db.Where("thread_id = ? AND category_id = ?", threadID, categoryID).Delete(&model.ThreadCategory{}).Error
}

// FindCategoriesByThreadID implements CategoryRepository
func (r *categoryRepositoryImpl) FindCategoriesByThreadID(threadID string) ([]*model.Category, error) {
	var categories []*model.Category

	err := r.db.Table("categories").
		Joins("JOIN thread_categories ON categories.category_id = thread_categories.category_id").
		Where("thread_categories.thread_id = ?", threadID).
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}

// pollRepositoryImpl implements PollRepository
type pollRepositoryImpl struct {
	db *gorm.DB
}

// CreatePoll implements PollRepository
func (r *pollRepositoryImpl) CreatePoll(poll *model.Poll) error {
	return r.db.Create(poll).Error
}

// CreatePollOptions implements PollRepository
func (r *pollRepositoryImpl) CreatePollOptions(options []*model.PollOption) error {
	return r.db.Create(options).Error
}

// FindPollByID implements PollRepository
func (r *pollRepositoryImpl) FindPollByID(id string) (*model.Poll, error) {
	var poll model.Poll
	err := r.db.First(&poll, "poll_id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPollNotFound
		}
		return nil, err
	}
	return &poll, nil
}

// FindPollByThreadID implements PollRepository
func (r *pollRepositoryImpl) FindPollByThreadID(threadID string) (*model.Poll, error) {
	var poll model.Poll
	err := r.db.Where("thread_id = ?", threadID).First(&poll).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPollNotFound
		}
		return nil, err
	}
	return &poll, nil
}

// FindPollOptionByID implements PollRepository
func (r *pollRepositoryImpl) FindPollOptionByID(id string) (*model.PollOption, error) {
	var option model.PollOption
	err := r.db.First(&option, "option_id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("poll option not found")
		}
		return nil, err
	}
	return &option, nil
}

// FindPollOptionsByPollID implements PollRepository
func (r *pollRepositoryImpl) FindPollOptionsByPollID(pollID string) ([]*model.PollOption, error) {
	var options []*model.PollOption
	err := r.db.Where("poll_id = ?", pollID).Find(&options).Error
	if err != nil {
		return nil, err
	}
	return options, nil
}

// CreateVote implements PollRepository
func (r *pollRepositoryImpl) CreateVote(vote *model.PollVote) error {
	return r.db.Create(vote).Error
}

// DeleteVote implements PollRepository
func (r *pollRepositoryImpl) DeleteVote(userID, pollID string) error {
	return r.db.Where("user_id = ? AND poll_id = ?", userID, pollID).Delete(&model.PollVote{}).Error
}

// FindVoteByUserAndPoll implements PollRepository
func (r *pollRepositoryImpl) FindVoteByUserAndPoll(userID, pollID string) (*model.PollVote, error) {
	var vote model.PollVote
	err := r.db.Where("user_id = ? AND poll_id = ?", userID, pollID).First(&vote).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No vote found, not an error
		}
		return nil, err
	}
	return &vote, nil
}

// GetPollVoteCounts implements PollRepository
func (r *pollRepositoryImpl) GetPollVoteCounts(pollID string) (map[string]int64, int64, error) {
	var votes []*model.PollVote
	voteCounts := make(map[string]int64)

	err := r.db.Where("poll_id = ?", pollID).Find(&votes).Error
	if err != nil {
		return nil, 0, err
	}

	totalVotes := int64(len(votes))

	for _, vote := range votes {
		optionID := vote.OptionID.String()
		voteCounts[optionID]++
	}

	return voteCounts, totalVotes, nil
}

// IsPollClosed implements PollRepository
func (r *pollRepositoryImpl) IsPollClosed(pollID string) (bool, error) {
	var poll model.Poll
	err := r.db.Select("closes_at").First(&poll, "poll_id = ?", pollID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrPollNotFound
		}
		return false, err
	}

	return time.Now().After(poll.ClosesAt), nil
}
