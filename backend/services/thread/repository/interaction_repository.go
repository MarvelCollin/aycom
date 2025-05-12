package repository

import (
	"errors"
	"fmt"
	"log"
	"time"

	"aycom/backend/services/thread/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InteractionRepository defines the methods for user-thread interactions (likes, reposts, bookmarks)
type InteractionRepository interface {
	// Like methods
	LikeThread(userID, threadID string) error
	LikeReply(userID, replyID string) error
	UnlikeThread(userID, threadID string) error
	UnlikeReply(userID, replyID string) error
	IsThreadLikedByUser(userID, threadID string) (bool, error)
	IsReplyLikedByUser(userID, replyID string) (bool, error)
	CountThreadLikes(threadID string) (int64, error)
	CountReplyLikes(replyID string) (int64, error)

	// Repost methods
	RepostThread(repost *model.Repost) error
	RemoveRepost(userID, threadID string) error
	IsThreadRepostedByUser(userID, threadID string) (bool, error)
	CountThreadReposts(threadID string) (int64, error)

	// Bookmark methods
	BookmarkThread(userID, threadID string) error
	RemoveBookmark(userID, threadID string) error
	IsThreadBookmarkedByUser(userID, threadID string) (bool, error)
	GetUserBookmarks(userID string, page, limit int) ([]*model.Thread, error)
	CountThreadBookmarks(threadID string) (int64, error)

	// Reply bookmark methods
	BookmarkReply(userID, replyID string) error
	RemoveReplyBookmark(userID, replyID string) error
	IsReplyBookmarkedByUser(userID, replyID string) (bool, error)
	CountReplyBookmarks(replyID string) (int64, error)

	// New methods
	FindLikedThreadsByUserID(userID string, page, limit int) ([]string, error)

	// Bookmark operations
	BookmarkExists(userID, threadID string) (bool, error)
	CreateBookmark(bookmark *model.Bookmark) error
}

// PostgresInteractionRepository implements the InteractionRepository interface
type PostgresInteractionRepository struct {
	db *gorm.DB
}

// NewInteractionRepository creates a new interaction repository
func NewInteractionRepository(db *gorm.DB) InteractionRepository {
	return &PostgresInteractionRepository{db: db}
}

// LikeThread adds a like to a thread
func (r *PostgresInteractionRepository) LikeThread(userID, threadID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return errors.New("invalid UUID format for thread ID")
	}

	like := model.Like{
		UserID:   userUUID,
		ThreadID: &threadUUID,
	}

	return r.db.Create(&like).Error
}

// LikeReply adds a like to a reply
func (r *PostgresInteractionRepository) LikeReply(userID, replyID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid UUID format for user ID")
	}

	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return errors.New("invalid UUID format for reply ID")
	}

	like := model.Like{
		UserID:  userUUID,
		ReplyID: &replyUUID,
	}

	return r.db.Create(&like).Error
}

// UnlikeThread removes a like from a thread
func (r *PostgresInteractionRepository) UnlikeThread(userID, threadID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return errors.New("invalid UUID format for thread ID")
	}

	return r.db.Where("user_id = ? AND thread_id = ?", userUUID, threadUUID).Delete(&model.Like{}).Error
}

// UnlikeReply removes a like from a reply
func (r *PostgresInteractionRepository) UnlikeReply(userID, replyID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid UUID format for user ID")
	}

	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return errors.New("invalid UUID format for reply ID")
	}

	return r.db.Where("user_id = ? AND reply_id = ?", userUUID, replyUUID).Delete(&model.Like{}).Error
}

// IsThreadLikedByUser checks if a thread is liked by a specific user
func (r *PostgresInteractionRepository) IsThreadLikedByUser(userID, threadID string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return false, errors.New("invalid UUID format for thread ID")
	}

	var count int64
	r.db.Model(&model.Like{}).Where("user_id = ? AND thread_id = ?", userUUID, threadUUID).Count(&count)
	return count > 0, nil
}

// IsReplyLikedByUser checks if a reply is liked by a specific user
func (r *PostgresInteractionRepository) IsReplyLikedByUser(userID, replyID string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, errors.New("invalid UUID format for user ID")
	}

	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return false, errors.New("invalid UUID format for reply ID")
	}

	var count int64
	r.db.Model(&model.Like{}).Where("user_id = ? AND reply_id = ?", userUUID, replyUUID).Count(&count)
	return count > 0, nil
}

// CountThreadLikes counts the number of likes for a thread
func (r *PostgresInteractionRepository) CountThreadLikes(threadID string) (int64, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return 0, errors.New("invalid UUID format for thread ID")
	}

	var count int64
	r.db.Model(&model.Like{}).Where("thread_id = ?", threadUUID).Count(&count)
	return count, nil
}

// CountReplyLikes counts the number of likes for a reply
func (r *PostgresInteractionRepository) CountReplyLikes(replyID string) (int64, error) {
	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return 0, errors.New("invalid UUID format for reply ID")
	}

	var count int64
	r.db.Model(&model.Like{}).Where("reply_id = ?", replyUUID).Count(&count)
	return count, nil
}

// RepostThread reposts a thread
func (r *PostgresInteractionRepository) RepostThread(repost *model.Repost) error {
	return r.db.Create(repost).Error
}

// RemoveRepost removes a thread repost
func (r *PostgresInteractionRepository) RemoveRepost(userID, threadID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return errors.New("invalid UUID format for thread ID")
	}

	return r.db.Where("user_id = ? AND thread_id = ?", userUUID, threadUUID).Delete(&model.Repost{}).Error
}

// IsThreadRepostedByUser checks if a thread is reposted by a specific user
func (r *PostgresInteractionRepository) IsThreadRepostedByUser(userID, threadID string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return false, errors.New("invalid UUID format for thread ID")
	}

	var count int64
	r.db.Model(&model.Repost{}).Where("user_id = ? AND thread_id = ?", userUUID, threadUUID).Count(&count)
	return count > 0, nil
}

// CountThreadReposts counts the number of reposts for a thread
func (r *PostgresInteractionRepository) CountThreadReposts(threadID string) (int64, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return 0, errors.New("invalid UUID format for thread ID")
	}

	var count int64
	r.db.Model(&model.Repost{}).Where("thread_id = ?", threadUUID).Count(&count)
	return count, nil
}

// BookmarkThread adds a bookmark to a thread
func (r *PostgresInteractionRepository) BookmarkThread(userID, threadID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return errors.New("invalid UUID format for thread ID")
	}

	// Check if user has already bookmarked this thread
	hasBookmarked, err := r.IsThreadBookmarkedByUser(userID, threadID)
	if err != nil {
		return err
	}

	if hasBookmarked {
		return nil // Return success for idempotence
	}

	bookmark := model.Bookmark{
		UserID:   userUUID,
		ThreadID: threadUUID,
	}

	return r.db.Create(&bookmark).Error
}

// RemoveBookmark removes a bookmark from a thread
func (r *PostgresInteractionRepository) RemoveBookmark(userID, threadID string) error {
	log.Printf("Repository RemoveBookmark called with userID: %s, threadID: %s", userID, threadID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("ERROR: Invalid UUID format for user ID: %s - %v", userID, err)
		return errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		log.Printf("ERROR: Invalid UUID format for thread ID: %s - %v", threadID, err)
		return errors.New("invalid UUID format for thread ID")
	}

	// Check if the bookmark exists before attempting to remove
	exists, err := r.IsThreadBookmarkedByUser(userID, threadID)
	if err != nil {
		log.Printf("ERROR: Failed to check if bookmark exists: %v", err)
		return err
	}

	if !exists {
		log.Printf("Bookmark does not exist for userID: %s, threadID: %s, returning success", userID, threadID)
		return nil // Return success for idempotence
	}

	// Use soft delete by setting deleted_at timestamp
	result := r.db.Model(&model.Bookmark{}).Where("user_id = ? AND thread_id = ?", userUUID, threadUUID).Update("deleted_at", time.Now())

	if result.Error != nil {
		log.Printf("ERROR: Failed to remove bookmark: %v", result.Error)
		return result.Error
	}

	log.Printf("Successfully removed bookmark for userID: %s, threadID: %s, rows affected: %d", userID, threadID, result.RowsAffected)
	return nil
}

// IsThreadBookmarkedByUser checks if a thread is bookmarked by a specific user
func (r *PostgresInteractionRepository) IsThreadBookmarkedByUser(userID, threadID string) (bool, error) {
	log.Printf("Checking if thread is bookmarked - userID: %s, threadID: %s", userID, threadID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("ERROR: Invalid UUID format for user ID: %s - %v", userID, err)
		return false, errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		log.Printf("ERROR: Invalid UUID format for thread ID: %s - %v", threadID, err)
		return false, errors.New("invalid UUID format for thread ID")
	}

	var count int64
	result := r.db.Model(&model.Bookmark{}).Where("user_id = ? AND thread_id = ? AND deleted_at IS NULL", userUUID, threadUUID).Count(&count)

	if result.Error != nil {
		log.Printf("ERROR: Failed to check if thread is bookmarked: %v", result.Error)
		return false, result.Error
	}

	isBookmarked := count > 0
	log.Printf("Thread bookmarked check result - userID: %s, threadID: %s, isBookmarked: %v", userID, threadID, isBookmarked)

	return isBookmarked, nil
}

// GetUserBookmarks gets all bookmarked threads for a user with pagination
func (r *PostgresInteractionRepository) GetUserBookmarks(userID string, page, limit int) ([]*model.Thread, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid UUID format for user ID")
	}

	var threads []*model.Thread
	offset := (page - 1) * limit

	result := r.db.Table("threads").
		Joins("JOIN bookmarks ON threads.thread_id = bookmarks.thread_id").
		Where("bookmarks.user_id = ?", userUUID).
		Order("bookmarks.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads)

	if result.Error != nil {
		return nil, result.Error
	}

	return threads, nil
}

// CountThreadBookmarks counts the number of bookmarks for a thread
func (r *PostgresInteractionRepository) CountThreadBookmarks(threadID string) (int64, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return 0, errors.New("invalid UUID format for thread ID")
	}

	var count int64
	r.db.Model(&model.Bookmark{}).Where("thread_id = ?", threadUUID).Count(&count)
	return count, nil
}

// BookmarkReply adds a bookmark to a reply
func (r *PostgresInteractionRepository) BookmarkReply(userID, replyID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid UUID format for user ID")
	}

	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return errors.New("invalid UUID format for reply ID")
	}

	bookmark := model.Bookmark{
		UserID:  userUUID,
		ReplyID: &replyUUID,
		// ThreadID is required by the model, so we use a zero UUID
		ThreadID: uuid.Nil,
	}

	return r.db.Create(&bookmark).Error
}

// RemoveReplyBookmark removes a bookmark from a reply
func (r *PostgresInteractionRepository) RemoveReplyBookmark(userID, replyID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid UUID format for user ID")
	}

	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return errors.New("invalid UUID format for reply ID")
	}

	return r.db.Where("user_id = ? AND reply_id = ?", userUUID, replyUUID).Delete(&model.Bookmark{}).Error
}

// IsReplyBookmarkedByUser checks if a reply is bookmarked by a specific user
func (r *PostgresInteractionRepository) IsReplyBookmarkedByUser(userID, replyID string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, errors.New("invalid UUID format for user ID")
	}

	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return false, errors.New("invalid UUID format for reply ID")
	}

	var count int64
	r.db.Model(&model.Bookmark{}).Where("user_id = ? AND reply_id = ?", userUUID, replyUUID).Count(&count)
	return count > 0, nil
}

// CountReplyBookmarks counts the number of bookmarks for a reply
func (r *PostgresInteractionRepository) CountReplyBookmarks(replyID string) (int64, error) {
	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return 0, errors.New("invalid UUID format for reply ID")
	}

	var count int64
	r.db.Model(&model.Bookmark{}).Where("reply_id = ?", replyUUID).Count(&count)
	return count, nil
}

// FindLikedThreadsByUserID finds all thread IDs that were liked by a specific user
func (r *PostgresInteractionRepository) FindLikedThreadsByUserID(userID string, page, limit int) ([]string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid UUID format for user ID")
	}

	var threadIDs []string
	offset := (page - 1) * limit

	// Find all thread likes by the user
	rows, err := r.db.Table("likes").
		Select("thread_id").
		Where("user_id = ? AND thread_id IS NOT NULL", userUUID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Convert UUIDs to strings
	for rows.Next() {
		var threadID uuid.UUID
		if err := rows.Scan(&threadID); err != nil {
			return nil, err
		}
		threadIDs = append(threadIDs, threadID.String())
	}

	return threadIDs, nil
}

// inspectBookmarksTableSchema checks if the bookmarks table exists and logs its schema
func (r *PostgresInteractionRepository) inspectBookmarksTableSchema() {
	log.Println("Inspecting bookmarks table schema...")

	// Check if table exists
	var tableExists bool
	r.db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'bookmarks')").Scan(&tableExists)

	if !tableExists {
		log.Println("ERROR: bookmarks table does not exist!")
		return
	}

	log.Println("bookmarks table exists, checking columns...")

	// Get column information
	type ColumnInfo struct {
		ColumnName string
		DataType   string
		IsNullable string
	}

	var columns []ColumnInfo
	r.db.Raw(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'bookmarks'
		ORDER BY ordinal_position
	`).Scan(&columns)

	for _, col := range columns {
		log.Printf("Column: %s, Type: %s, Nullable: %s", col.ColumnName, col.DataType, col.IsNullable)
	}

	// Check primary key constraints
	var primaryKey string
	r.db.Raw(`
		SELECT tc.constraint_name
		FROM information_schema.table_constraints tc
		WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_name = 'bookmarks'
	`).Scan(&primaryKey)

	if primaryKey == "" {
		log.Println("WARNING: bookmarks table has no primary key constraint")
	} else {
		log.Printf("Primary key constraint: %s", primaryKey)
	}

	// Check unique constraints
	var uniqueConstraints []string
	r.db.Raw(`
		SELECT tc.constraint_name
		FROM information_schema.table_constraints tc
		WHERE tc.constraint_type = 'UNIQUE' AND tc.table_name = 'bookmarks'
	`).Scan(&uniqueConstraints)

	if len(uniqueConstraints) == 0 {
		log.Println("WARNING: bookmarks table has no unique constraints")
	} else {
		for _, uc := range uniqueConstraints {
			log.Printf("Unique constraint: %s", uc)
		}
	}
}

// CheckDBConnection verifies that the database connection is working
func (r *PostgresInteractionRepository) CheckDBConnection() error {
	log.Println("Checking database connection...")
	sqlDB, err := r.db.DB()
	if err != nil {
		log.Printf("ERROR: Failed to get SQL DB from GORM: %v", err)
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Printf("ERROR: Database ping failed: %v", err)
		return err
	}

	// Inspect bookmarks table schema
	r.inspectBookmarksTableSchema()

	// Try a simple query
	var count int64
	result := r.db.Table("bookmarks").Count(&count)
	if result.Error != nil {
		log.Printf("ERROR: Failed to execute count query on bookmarks table: %v", result.Error)
		return result.Error
	}

	log.Printf("Database connection is healthy. Total bookmarks in database: %d", count)
	return nil
}

// BookmarkExists checks if a bookmark already exists
func (r *PostgresInteractionRepository) BookmarkExists(userID, threadID string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID format: %w", err)
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return false, fmt.Errorf("invalid thread ID format: %w", err)
	}

	var count int64
	result := r.db.Model(&model.Bookmark{}).
		Where("user_id = ? AND thread_id = ?", userUUID, threadUUID).
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("error checking bookmark existence: %w", result.Error)
	}

	return count > 0, nil
}

// CreateBookmark creates a new bookmark record
func (r *PostgresInteractionRepository) CreateBookmark(bookmark *model.Bookmark) error {
	// Debugging: print the bookmark details
	log.Printf("Creating bookmark: user=%s, thread=%s", bookmark.UserID, bookmark.ThreadID)

	// Explicitly check if bookmark exists first
	var exists bool
	err := r.db.Raw("SELECT EXISTS(SELECT 1 FROM bookmarks WHERE user_id = ? AND thread_id = ?)",
		bookmark.UserID, bookmark.ThreadID).Scan(&exists).Error

	if err != nil {
		log.Printf("Error checking existence before create: %v", err)
		return err
	}

	if exists {
		log.Printf("Bookmark already exists, skipping creation")
		return nil
	}

	// If not exists, create the bookmark
	log.Printf("Bookmark doesn't exist, creating now")
	result := r.db.Create(bookmark)

	if result.Error != nil {
		log.Printf("Error in Create operation: %v", result.Error)
		return result.Error
	}

	log.Printf("Bookmark created successfully, rows affected: %d", result.RowsAffected)
	return nil
}
