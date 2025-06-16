package repository

import (
	"aycom/backend/services/thread/model"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InteractionRepository interface {
	LikeThread(userID, threadID string) error
	LikeReply(userID, replyID string) error
	UnlikeThread(userID, threadID string) error
	UnlikeReply(userID, replyID string) error
	IsThreadLikedByUser(userID, threadID string) (bool, error)
	IsReplyLikedByUser(userID, replyID string) (bool, error)
	CountThreadLikes(threadID string) (int64, error)
	CountReplyLikes(replyID string) (int64, error)

	RepostThread(repost *model.Repost) error
	RemoveRepost(userID, threadID string) error
	IsThreadRepostedByUser(userID, threadID string) (bool, error)
	CountThreadReposts(threadID string) (int64, error)

	BookmarkThread(userID, threadID string) error
	RemoveBookmark(userID, threadID string) error
	IsThreadBookmarkedByUser(userID, threadID string) (bool, error)
	GetUserBookmarks(userID string, page, limit int) ([]*model.Thread, error)
	CountThreadBookmarks(threadID string) (int64, error)

	BookmarkReply(userID, replyID string) error
	RemoveReplyBookmark(userID, replyID string) error
	IsReplyBookmarkedByUser(userID, replyID string) (bool, error)
	CountReplyBookmarks(replyID string) (int64, error)

	FindLikedThreadsByUserID(userID string, page, limit int) ([]string, error)

	BookmarkExists(userID, threadID string) (bool, error)
	CreateBookmark(bookmark *model.Bookmark) error

	BatchCountThreadLikes(threadIDs []string) (map[string]int64, error)
	BatchCheckThreadsLikedByUser(userID string, threadIDs []string) (map[string]bool, error)
}

type PostgresInteractionRepository struct {
	db *gorm.DB
}

func NewInteractionRepository(db *gorm.DB) InteractionRepository {
	return &PostgresInteractionRepository{db: db}
}

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
	err = r.db.Exec(`
		INSERT INTO likes (user_id, thread_id, created_at, deleted_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, thread_id) WHERE thread_id IS NOT NULL AND reply_id IS NULL AND deleted_at IS NULL DO NOTHING`,
		userUUID, threadUUID, time.Now(), nil).Error

	if err != nil {
		log.Printf("Error creating like: %v", err)
		return fmt.Errorf("failed to create like: %w", err)
	}

	log.Printf("Thread %s liked by user %s", threadID, userID)
	return nil
}

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
	err = r.db.Exec(`
		INSERT INTO likes (user_id, reply_id, created_at, deleted_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, reply_id) WHERE reply_id IS NOT NULL AND thread_id IS NULL AND deleted_at IS NULL DO NOTHING`,
		userUUID, replyUUID, time.Now(), nil).Error

	if err != nil {
		log.Printf("Error creating reply like: %v", err)
		return fmt.Errorf("failed to create reply like: %w", err)
	}

	log.Printf("Reply %s liked by user %s", replyID, userID)
	return nil
}

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

	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to begin transaction: %v", tx.Error)
		return tx.Error
	}

	errChan := make(chan error, 1)

	defer func() {
		if r := recover(); r != nil {

			errStr := fmt.Sprintf("Recovered from panic in UnlikeThread: %v", r)
			log.Printf(errStr)
			tx.Rollback()
			errChan <- errors.New(errStr)
		}
	}()

	result := tx.Model(&model.Like{}).
		Where("user_id = ? AND thread_id = ? AND deleted_at IS NULL", userUUID, threadUUID).
		Update("deleted_at", time.Now())

	if result.Error != nil {
		log.Printf("Error soft-deleting like: %v", result.Error)
		tx.Rollback()
		return result.Error
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	select {
	case err := <-errChan:
		return err
	default:

	}

	log.Printf("Successfully processed unlike for thread %s by user %s (rows affected: %d)", threadID, userID, result.RowsAffected)
	return nil
}

func (r *PostgresInteractionRepository) CleanupSoftDeletedLikes(cutoffTime time.Time) (int64, error) {
	log.Printf("Running cleanup of soft-deleted likes older than %v", cutoffTime)

	result := r.db.Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoffTime).
		Delete(&model.Like{})

	if result.Error != nil {
		log.Printf("Error cleaning up soft-deleted likes: %v", result.Error)
		return 0, result.Error
	}

	log.Printf("Successfully cleaned up %d soft-deleted like records", result.RowsAffected)
	return result.RowsAffected, nil
}

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

func (r *PostgresInteractionRepository) CountThreadLikes(threadID string) (int64, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return 0, errors.New("invalid UUID format for thread ID")
	}

	var count int64
	r.db.Model(&model.Like{}).Where("thread_id = ? AND deleted_at IS NULL", threadUUID).Count(&count)
	return count, nil
}

func (r *PostgresInteractionRepository) CountReplyLikes(replyID string) (int64, error) {
	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return 0, errors.New("invalid UUID format for reply ID")
	}

	var count int64
	r.db.Model(&model.Like{}).Where("reply_id = ? AND deleted_at IS NULL", replyUUID).Count(&count)
	return count, nil
}

func (r *PostgresInteractionRepository) RepostThread(repost *model.Repost) error {

	userUUID := repost.UserID
	threadUUID := repost.ThreadID

	var count int64
	if err := r.db.Model(&model.Repost{}).
		Where("user_id = ? AND thread_id = ? AND deleted_at IS NULL", userUUID, threadUUID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing repost: %w", err)
	}

	if count > 0 {
		log.Printf("Repost already exists for user %s and thread %s", userUUID, threadUUID)
		return nil
	}

	if err := r.db.Create(repost).Error; err != nil {
		return fmt.Errorf("failed to create repost: %w", err)
	}

	log.Printf("Successfully created repost for user %s and thread %s", userUUID, threadUUID)
	return nil
}

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
	r.db.Model(&model.Repost{}).Where("user_id = ? AND thread_id = ? AND deleted_at IS NULL", userUUID, threadUUID).Count(&count)
	return count > 0, nil
}

func (r *PostgresInteractionRepository) CountThreadReposts(threadID string) (int64, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return 0, errors.New("invalid UUID format for thread ID")
	}

	var count int64
	r.db.Model(&model.Repost{}).Where("thread_id = ? AND deleted_at IS NULL", threadUUID).Count(&count)
	return count, nil
}

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

	bookmark := &model.Bookmark{
		UserID:    userUUID,
		ThreadID:  threadUUID,
		CreatedAt: time.Now(),
	}

	if err := r.CreateBookmark(bookmark); err != nil {
		log.Printf("Error creating bookmark: %v", err)
		return fmt.Errorf("failed to create bookmark: %w", err)
	}

	log.Printf("Thread %s bookmarked by user %s", threadID, userID)
	return nil
}

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

	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("Failed to begin transaction: %v", tx.Error)
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in RemoveBookmark: %v", r)
			tx.Rollback()
		}
	}()

	result := tx.Where("user_id = ? AND thread_id = ?", userUUID, threadUUID).Delete(&model.Bookmark{})
	if result.Error != nil {
		log.Printf("Error removing bookmark: %v", result.Error)
		tx.Rollback()
		return result.Error
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Printf("Successfully processed unbookmark for thread %s by user %s (rows affected: %d)", threadID, userID, result.RowsAffected)
	return nil
}

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

func (r *PostgresInteractionRepository) GetUserBookmarks(userID string, page, limit int) ([]*model.Thread, error) {
	log.Printf("GetUserBookmarks called with userID: %s, page: %d, limit: %d", userID, page, limit)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("ERROR: Invalid UUID format for user ID: %s - %v", userID, err)
		return nil, errors.New("invalid UUID format for user ID")
	}

	var threads []*model.Thread
	offset := (page - 1) * limit

	log.Printf("Executing query to get bookmarks for user %s with offset %d and limit %d", userID, offset, limit)

	result := r.db.Table("threads").
		Joins("JOIN bookmarks ON threads.thread_id = bookmarks.thread_id").
		Where("bookmarks.user_id = ? AND bookmarks.deleted_at IS NULL", userUUID).
		Order("bookmarks.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads)

	if result.Error != nil {
		log.Printf("ERROR: Failed to get bookmarks: %v", result.Error)
		return nil, result.Error
	}

	log.Printf("Successfully retrieved %d bookmarks for user %s", len(threads), userID)

	return threads, nil
}

func (r *PostgresInteractionRepository) CountThreadBookmarks(threadID string) (int64, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return 0, errors.New("invalid UUID format for thread ID")
	}

	var count int64
	r.db.Model(&model.Bookmark{}).Where("thread_id = ? AND deleted_at IS NULL", threadUUID).Count(&count)
	return count, nil
}

func (r *PostgresInteractionRepository) BookmarkReply(userID, replyID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid UUID format for user ID")
	}

	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return errors.New("invalid UUID format for reply ID")
	}

	result := r.db.Exec(`
		INSERT INTO bookmarks (user_id, thread_id, reply_id, created_at, deleted_at)
		VALUES ($1, NULL, $2, $3, NULL)
		ON CONFLICT (user_id, reply_id) WHERE reply_id IS NOT NULL AND thread_id IS NULL AND deleted_at IS NULL DO NOTHING`,
		userUUID, replyUUID, time.Now())

	return result.Error
}

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
		Where("user_id = ? AND reply_id = ? AND deleted_at IS NULL", userUUID, replyUUID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *PostgresInteractionRepository) CountReplyBookmarks(replyID string) (int64, error) {
	replyUUID, err := uuid.Parse(replyID)
	if err != nil {
		return 0, errors.New("invalid UUID format for reply ID")
	}

	var count int64
	r.db.Model(&model.Bookmark{}).Where("reply_id = ? AND deleted_at IS NULL", replyUUID).Count(&count)
	return count, nil
}

func (r *PostgresInteractionRepository) FindLikedThreadsByUserID(userID string, page, limit int) ([]string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid UUID format for user ID")
	}

	offset := (page - 1) * limit
	var likes []model.Like

	err = r.db.Model(&model.Like{}).
		Where("user_id = ? AND thread_id IS NOT NULL AND deleted_at IS NULL", userUUID).
		Offset(offset).
		Limit(limit).
		Find(&likes).Error

	if err != nil {
		return nil, err
	}

	threadIDs := make([]string, len(likes))
	for i, like := range likes {
		threadIDs[i] = like.ThreadID.String()
	}

	return threadIDs, nil
}

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
		Where("user_id = ? AND thread_id = ? AND deleted_at IS NULL", userUUID, threadUUID).
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("error checking bookmark existence: %w", result.Error)
	}

	return count > 0, nil
}

func (r *PostgresInteractionRepository) CreateBookmark(bookmark *model.Bookmark) error {
	var count int64
	if err := r.db.Model(&model.Bookmark{}).
		Where("user_id = ? AND thread_id = ? AND deleted_at IS NULL", bookmark.UserID, bookmark.ThreadID).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	return r.db.Create(bookmark).Error
}

func (r *PostgresInteractionRepository) BatchCountThreadLikes(threadIDs []string) (map[string]int64, error) {
	if len(threadIDs) == 0 {
		return map[string]int64{}, nil
	}

	threadUUIDs := make([]uuid.UUID, 0, len(threadIDs))
	for _, idStr := range threadIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Skipping invalid thread ID %s: %v", idStr, err)
			continue
		}
		threadUUIDs = append(threadUUIDs, id)
	}

	if len(threadUUIDs) == 0 {
		log.Printf("No valid thread UUIDs to query")
		return map[string]int64{}, nil
	}

	type Result struct {
		ThreadID uuid.UUID `gorm:"column:thread_id"`
		Count    int64     `gorm:"column:count"`
	}
	var results []Result

	err := r.db.Model(&model.Like{}).
		Select("thread_id, COUNT(*) as count").
		Where("thread_id IN ? AND deleted_at IS NULL", threadUUIDs).
		Group("thread_id").
		Find(&results).Error

	if err != nil {
		log.Printf("Error batch counting likes: %v", err)
		return nil, err
	}

	countMap := make(map[string]int64, len(results))
	for _, result := range results {
		countMap[result.ThreadID.String()] = result.Count
	}

	for _, idStr := range threadIDs {
		if _, exists := countMap[idStr]; !exists {
			countMap[idStr] = 0
		}
	}

	return countMap, nil
}

func (r *PostgresInteractionRepository) BatchCheckThreadsLikedByUser(userID string, threadIDs []string) (map[string]bool, error) {
	if len(threadIDs) == 0 {
		return map[string]bool{}, nil
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid UUID format for user ID")
	}

	threadUUIDs := make([]uuid.UUID, 0, len(threadIDs))
	for _, idStr := range threadIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Skipping invalid thread ID %s: %v", idStr, err)
			continue
		}
		threadUUIDs = append(threadUUIDs, id)
	}

	if len(threadUUIDs) == 0 {
		log.Printf("No valid thread UUIDs to query")
		return map[string]bool{}, nil
	}

	type Result struct {
		ThreadID uuid.UUID `gorm:"column:thread_id"`
	}
	var results []Result

	err = r.db.Model(&model.Like{}).
		Select("thread_id").
		Where("user_id = ? AND thread_id IN ? AND deleted_at IS NULL", userUUID, threadUUIDs).
		Find(&results).Error

	if err != nil {
		log.Printf("Error batch checking liked threads: %v", err)
		return nil, err
	}

	likedMap := make(map[string]bool, len(threadIDs))

	for _, idStr := range threadIDs {
		likedMap[idStr] = false
	}

	for _, result := range results {
		likedMap[result.ThreadID.String()] = true
	}

	return likedMap, nil
}