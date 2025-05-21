package repository

import (
	"errors"
	"fmt"
	"log"
	"time"

	"aycom/backend/services/thread/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	log.Printf("LikeThread called with userID: %s, threadID: %s", userID, threadID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid UUID format for user ID: %s - %v", userID, err)
		return errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		log.Printf("Invalid UUID format for thread ID: %s - %v", threadID, err)
		return errors.New("invalid UUID format for thread ID")
	}

	// First check if like already exists outside transaction
	exists, err := r.IsThreadLikedByUser(userID, threadID)
	if err != nil {
		log.Printf("Error checking if thread is already liked: %v", err)
		return err
	}
	if exists {
		log.Printf("Thread %s is already liked by user %s", threadID, userID)
		return nil
	}

	// Start transaction with serializable isolation level
	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to begin transaction: %v", tx.Error)
		return tx.Error
	}

	// Set transaction isolation level to serializable
	if err := tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE").Error; err != nil {
		tx.Rollback()
		log.Printf("Failed to set transaction isolation level: %v", err)
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in LikeThread: %v", r)
			tx.Rollback()
		}
	}()

	// Use FOR UPDATE SKIP LOCKED to prevent deadlocks
	var count int64
	if err := tx.Model(&model.Like{}).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("user_id = ? AND thread_id = ? AND deleted_at IS NULL", userUUID, threadUUID).
		Count(&count).Error; err != nil {
		log.Printf("Error checking for existing like in transaction: %v", err)
		tx.Rollback()
		return err
	}

	// Only insert if it doesn't exist
	if count == 0 {
		log.Printf("No existing like found, creating new like")

		// Use INSERT ... ON CONFLICT DO NOTHING for additional safety
		if err := tx.Exec(`
			INSERT INTO likes (id, user_id, thread_id, created_at)
			VALUES (?, ?, ?, NOW())
			ON CONFLICT (user_id, thread_id) WHERE deleted_at IS NULL DO NOTHING`,
			uuid.New(), userUUID, threadUUID).Error; err != nil {
			log.Printf("Error creating like: %v", err)
			tx.Rollback()
			return err
		}
		log.Printf("Like created successfully")
	} else {
		log.Printf("Like already exists, skipping creation")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Printf("Transaction committed successfully, thread %s liked by user %s", threadID, userID)
	return nil
}

// LikeReply adds a like to a reply
func (r *PostgresInteractionRepository) LikeReply(userID, replyID string) error {
	log.Printf("LikeReply called with userID: %s, replyID: %s", userID, replyID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid UUID format for user ID: %s - %v", userID, err)
		return errors.New("invalid UUID format for user ID")
	}

	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		log.Printf("Invalid UUID format for reply ID: %s - %v", replyID, err)
		return errors.New("invalid UUID format for reply ID")
	}

	// Let's first check if the like already exists outside the transaction
	hasLiked, err := r.IsReplyLikedByUser(userID, replyID)
	if err != nil {
		log.Printf("Error checking if reply is already liked: %v", err)
		return err
	}

	if hasLiked {
		log.Printf("Reply %s is already liked by user %s, returning success", replyID, userID)
		return nil
	}

	// Start a transaction to make this operation atomic
	log.Printf("Starting transaction for new reply like")
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in LikeReply: %v", r)
			tx.Rollback()
		}
	}()

	// Check if the like already exists in a transaction
	var count int64
	if err := tx.Model(&model.Like{}).Where("user_id = ? AND reply_id = ? AND deleted_at IS NULL", userUUID, replyUUID).Count(&count).Error; err != nil {
		log.Printf("Error counting reply likes in transaction: %v", err)
		tx.Rollback()
		return err
	}

	// Only insert if it doesn't exist
	if count == 0 {
		log.Printf("No existing reply like found, creating new like")
		like := model.Like{
			UserID:  userUUID,
			ReplyID: &replyUUID,
		}

		if err := tx.Create(&like).Error; err != nil {
			log.Printf("Error creating reply like: %v", err)
			tx.Rollback()
			return err
		}
		log.Printf("Reply like created successfully")
	} else {
		log.Printf("Reply like already exists in transaction check, skipping creation")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Printf("Transaction committed successfully, reply %s liked by user %s", replyID, userID)
	return nil
}

// UnlikeThread removes a like from a thread
func (r *PostgresInteractionRepository) UnlikeThread(userID, threadID string) error {
	log.Printf("UnlikeThread called with userID: %s, threadID: %s", userID, threadID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid UUID format for user ID: %s - %v", userID, err)
		return errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		log.Printf("Invalid UUID format for thread ID: %s - %v", threadID, err)
		return errors.New("invalid UUID format for thread ID")
	}

	// Start a transaction for safety
	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to begin transaction: %v", tx.Error)
		return tx.Error
	}

	// Set a deferred rollback that will be ignored if we commit successfully
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in UnlikeThread: %v", r)
			tx.Rollback()
		}
	}()

	// Use soft deletion to keep history if needed
	result := tx.Where("user_id = ? AND thread_id = ?", userUUID, threadUUID).Delete(&model.Like{})
	if result.Error != nil {
		log.Printf("Error removing like: %v", result.Error)
		tx.Rollback()
		return result.Error
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	// Return success even if no rows were affected (unlike was idempotent)
	log.Printf("Successfully processed unlike for thread %s by user %s (rows affected: %d)", threadID, userID, result.RowsAffected)
	return nil
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
	err = r.db.Model(&model.Like{}).
		Where("user_id = ? AND thread_id = ? AND deleted_at IS NULL", userUUID, threadUUID).
		Count(&count).Error

	if err != nil {
		log.Printf("Error checking if thread is liked by user: %v", err)
		return false, err
	}

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
	err = r.db.Model(&model.Like{}).
		Where("user_id = ? AND reply_id = ? AND deleted_at IS NULL", userUUID, replyUUID).
		Count(&count).Error

	if err != nil {
		log.Printf("Error checking if reply is liked by user: %v", err)
		return false, err
	}

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
	log.Printf("BookmarkThread called with userID: %s, threadID: %s", userID, threadID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid UUID format for user ID: %s - %v", userID, err)
		return errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		log.Printf("Invalid UUID format for thread ID: %s - %v", threadID, err)
		return errors.New("invalid UUID format for thread ID")
	}

	// Start a transaction to make this operation atomic
	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to begin transaction: %v", tx.Error)
		return tx.Error
	}

	// Set a deferred rollback that will be ignored if we commit successfully
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in BookmarkThread: %v", r)
			tx.Rollback()
		}
	}()

	// Use FOR UPDATE locking to prevent race conditions
	var count int64
	if err := tx.Model(&model.Bookmark{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND thread_id = ? AND deleted_at IS NULL", userUUID, threadUUID).
		Count(&count).Error; err != nil {
		log.Printf("Error checking for existing bookmark in transaction: %v", err)
		tx.Rollback()
		return err
	}

	// Only insert if it doesn't exist
	if count == 0 {
		log.Printf("No existing bookmark found, creating new bookmark")

		// Use native SQL for an upsert-like operation that's more race-condition resistant
		if err := tx.Exec(`
			INSERT INTO bookmarks (user_id, thread_id, created_at)
			VALUES (?, ?, NOW())
			ON CONFLICT DO NOTHING`,
			userUUID, threadUUID).Error; err != nil {
			log.Printf("Error creating bookmark: %v", err)
			tx.Rollback()
			return err
		}
		log.Printf("Bookmark created successfully")
	} else {
		log.Printf("Bookmark already exists, skipping creation")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Printf("Transaction committed successfully, thread %s bookmarked by user %s", threadID, userID)
	return nil
}

// RemoveBookmark removes a bookmark from a thread
func (r *PostgresInteractionRepository) RemoveBookmark(userID, threadID string) error {
	log.Printf("RemoveBookmark called with userID: %s, threadID: %s", userID, threadID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid UUID format for user ID: %s - %v", userID, err)
		return errors.New("invalid UUID format for user ID")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		log.Printf("Invalid UUID format for thread ID: %s - %v", threadID, err)
		return errors.New("invalid UUID format for thread ID")
	}

	// Start a transaction for safety
	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to begin transaction: %v", tx.Error)
		return tx.Error
	}

	// Set a deferred rollback that will be ignored if we commit successfully
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in RemoveBookmark: %v", r)
			tx.Rollback()
		}
	}()

	// Delete the bookmark - use soft delete to maintain history if needed
	result := tx.Where("user_id = ? AND thread_id = ?", userUUID, threadUUID).Delete(&model.Bookmark{})
	if result.Error != nil {
		log.Printf("Error removing bookmark: %v", result.Error)
		tx.Rollback()
		return result.Error
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	// Return success even if no rows were affected (unbookmark is idempotent)
	log.Printf("Successfully processed unbookmark for thread %s by user %s (rows affected: %d)", threadID, userID, result.RowsAffected)
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

	// First, check if the bookmark already exists
	var count int64
	if err := r.db.Model(&model.Bookmark{}).
		Where("user_id = ? AND reply_id = ?", userUUID, replyUUID).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		// Bookmark already exists, nothing to do
		return nil
	}

	// Use raw SQL to ensure NULL is used for thread_id
	result := r.db.Exec(
		"INSERT INTO bookmarks (user_id, thread_id, reply_id, created_at) VALUES (?, NULL, ?, ?)",
		userUUID,
		replyUUID,
		time.Now(),
	)

	return result.Error
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

	// Use the reply_id field instead of thread_id
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
	err = r.db.Model(&model.Bookmark{}).
		Where("user_id = ? AND reply_id = ?", userUUID, replyUUID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

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

// FindLikedThreadsByUserID gets all thread IDs liked by a specific user
func (r *PostgresInteractionRepository) FindLikedThreadsByUserID(userID string, page, limit int) ([]string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid UUID format for user ID")
	}

	offset := (page - 1) * limit
	var likes []model.Like

	err = r.db.Model(&model.Like{}).
		Where("user_id = ? AND thread_id IS NOT NULL", userUUID).
		Offset(offset).
		Limit(limit).
		Find(&likes).Error

	if err != nil {
		return nil, err
	}

	// Extract thread IDs
	threadIDs := make([]string, len(likes))
	for i, like := range likes {
		threadIDs[i] = like.ThreadID.String()
	}

	return threadIDs, nil
}

// inspectBookmarksTableSchema logs information about the bookmarks table schema
func (r *PostgresInteractionRepository) inspectBookmarksTableSchema() {
	var columnNames []string
	r.db.Raw(`
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = 'bookmarks'
	`).Scan(&columnNames)

	for _, c := range columnNames {
		log.Printf("Column: %s", c)
	}

	var uniqueConstraints []string
	r.db.Raw(`
		SELECT tc.constraint_name
		FROM information_schema.table_constraints tc
		WHERE tc.constraint_type = 'UNIQUE' AND tc.table_name = 'bookmarks'
	`).Scan(&uniqueConstraints)
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
	// Check if the bookmark already exists
	var count int64
	if err := r.db.Model(&model.Bookmark{}).
		Where("user_id = ? AND thread_id = ?", bookmark.UserID, bookmark.ThreadID).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		// Bookmark already exists, nothing to do
		return nil
	}

	// Create the bookmark
	return r.db.Create(bookmark).Error
}
