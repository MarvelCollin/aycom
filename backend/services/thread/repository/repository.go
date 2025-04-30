package repository

import (
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/model"
	"gorm.io/gorm"
)

// ThreadRepository handles database operations for threads
type ThreadRepository interface {
	CreateThread(thread *model.Thread) error
	FindThreadByID(id string) (*model.Thread, error)
	FindThreadsByUserID(userID string, page, limit int) ([]*model.Thread, int64, error)
	FindThreadsByCommunityID(communityID string, page, limit int) ([]*model.Thread, int64, error)
	UpdateThread(thread *model.Thread) error
	DeleteThread(id string) error
	GetThreadStats(threadID string) (replyCount, likeCount, repostCount int64, err error)
	IncrementViewCount(threadID string) error
}

// ReplyRepository handles database operations for replies
type ReplyRepository interface {
	CreateReply(reply *model.Reply) error
	FindReplyByID(id string) (*model.Reply, error)
	FindRepliesByThreadID(threadID string, parentReplyID *string, page, limit int) ([]*model.Reply, int64, error)
	UpdateReply(reply *model.Reply) error
	DeleteReply(id string) error
	GetReplyStats(replyID string) (replyCount, likeCount int64, err error)
}

// LikeRepository handles database operations for likes
type LikeRepository interface {
	CreateThreadLike(userID, threadID string) error
	DeleteThreadLike(userID, threadID string) error
	CreateReplyLike(userID, replyID string) error
	DeleteReplyLike(userID, replyID string) error
	HasUserLikedThread(userID, threadID string) (bool, error)
	HasUserLikedReply(userID, replyID string) (bool, error)
	GetThreadLikeCount(threadID string) (int64, error)
	GetReplyLikeCount(replyID string) (int64, error)
}

// RepostRepository handles database operations for reposts
type RepostRepository interface {
	CreateRepost(userID, threadID string, repostText *string) error
	DeleteRepost(userID, threadID string) error
	HasUserReposted(userID, threadID string) (bool, error)
	GetRepostCount(threadID string) (int64, error)
}

// BookmarkRepository handles database operations for bookmarks
type BookmarkRepository interface {
	CreateBookmark(userID, threadID string) error
	DeleteBookmark(userID, threadID string) error
	HasUserBookmarked(userID, threadID string) (bool, error)
	GetUserBookmarks(userID string, page, limit int) ([]*model.Thread, int64, error)
}

// MediaRepository handles database operations for media
type MediaRepository interface {
	CreateMedia(media *model.Media) error
	FindMediaByID(id string) (*model.Media, error)
	FindMediaByThreadID(threadID string) ([]*model.Media, error)
	FindMediaByReplyID(replyID string) ([]*model.Media, error)
	DeleteMedia(id string) error
}

// HashtagRepository handles database operations for hashtags
type HashtagRepository interface {
	CreateHashtag(hashtag *model.Hashtag) error
	FindOrCreateHashtagByText(text string) (*model.Hashtag, error)
	FindHashtagByID(id string) (*model.Hashtag, error)
	FindHashtagByText(text string) (*model.Hashtag, error)
	AddHashtagToThread(threadID, hashtagID string) error
	RemoveHashtagFromThread(threadID, hashtagID string) error
	FindHashtagsByThreadID(threadID string) ([]*model.Hashtag, error)
	FindTrendingHashtags(limit int) ([]*model.Hashtag, int64, error)
}

// MentionRepository handles database operations for user mentions
type MentionRepository interface {
	CreateMention(mention *model.UserMention) error
	FindMentionsByThreadID(threadID string) ([]*model.UserMention, error)
	FindMentionsByReplyID(replyID string) ([]*model.UserMention, error)
	DeleteMention(id string) error
}

// CategoryRepository handles database operations for categories
type CategoryRepository interface {
	CreateCategory(category *model.Category) error
	FindOrCreateCategoryByName(name string, categoryType string) (*model.Category, error)
	FindCategoryByID(id string) (*model.Category, error)
	FindCategoryByName(name string, categoryType string) (*model.Category, error)
	AddCategoryToThread(threadID, categoryID string) error
	RemoveCategoryFromThread(threadID, categoryID string) error
	FindCategoriesByThreadID(threadID string) ([]*model.Category, error)
}

// PollRepository handles database operations for polls
type PollRepository interface {
	CreatePoll(poll *model.Poll) error
	CreatePollOptions(options []*model.PollOption) error
	FindPollByID(id string) (*model.Poll, error)
	FindPollByThreadID(threadID string) (*model.Poll, error)
	FindPollOptionByID(id string) (*model.PollOption, error)
	FindPollOptionsByPollID(pollID string) ([]*model.PollOption, error)
	CreateVote(vote *model.PollVote) error
	DeleteVote(userID, pollID string) error
	FindVoteByUserAndPoll(userID, pollID string) (*model.PollVote, error)
	GetPollVoteCounts(pollID string) (map[string]int64, int64, error)
	IsPollClosed(pollID string) (bool, error)
}

// ThreadRepositoryImpl implements ThreadRepository
type ThreadRepositoryImpl struct {
	db *gorm.DB
}

// NewThreadRepository creates a new thread repository
func NewThreadRepository(db *gorm.DB) ThreadRepository {
	return &threadRepositoryImpl{db: db}
}

// NewReplyRepository creates a new reply repository
func NewReplyRepository(db *gorm.DB) ReplyRepository {
	return &replyRepositoryImpl{db: db}
}

// NewLikeRepository creates a new like repository
func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepositoryImpl{db: db}
}

// NewRepostRepository creates a new repost repository
func NewRepostRepository(db *gorm.DB) RepostRepository {
	return &repostRepositoryImpl{db: db}
}

// NewBookmarkRepository creates a new bookmark repository
func NewBookmarkRepository(db *gorm.DB) BookmarkRepository {
	return &bookmarkRepositoryImpl{db: db}
}

// NewMediaRepository creates a new media repository
func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepositoryImpl{db: db}
}

// NewHashtagRepository creates a new hashtag repository
func NewHashtagRepository(db *gorm.DB) HashtagRepository {
	return &hashtagRepositoryImpl{db: db}
}

// NewMentionRepository creates a new mention repository
func NewMentionRepository(db *gorm.DB) MentionRepository {
	return &mentionRepositoryImpl{db: db}
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepositoryImpl{db: db}
}

// NewPollRepository creates a new poll repository
func NewPollRepository(db *gorm.DB) PollRepository {
	return &pollRepositoryImpl{db: db}
}

// NewThreadSeeder creates a new thread seeder
func NewThreadSeeder(db *gorm.DB) *ThreadSeeder {
	return &ThreadSeeder{db: db}
}

// ThreadSeeder seeds thread data for testing and development
type ThreadSeeder struct {
	db *gorm.DB
}

// SeedThreads seeds threads, replies, and interactions into the database
func (s *ThreadSeeder) SeedThreads() error {
	// Implementation for seeding test data would go here
	return nil
}

// Additional repository implementations would be in separate files
