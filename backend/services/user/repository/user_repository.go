package repository

import (
	"fmt"

	"aycom/backend/services/user/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *model.User) error
	FindUserByID(id string) (*model.User, error)
	FindUserByUsername(username string) (*model.User, error)
	FindUserByEmail(email string) (*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(id string) error
	UpdateUserVerification(userID string, isVerified bool) error

	UserExists(userID string) (bool, error)

	CreateFollow(follow *model.Follow) error
	DeleteFollow(followerID, followedID uuid.UUID) error
	CheckFollowExists(followerID, followedID uuid.UUID) (bool, error)
	GetFollowers(userID uuid.UUID, page, limit int) ([]*model.User, int, error)
	GetFollowing(userID uuid.UUID, page, limit int) ([]*model.User, int, error)

	IncrementFollowerCount(userID string) error
	DecrementFollowerCount(userID string) error
	IncrementFollowingCount(userID string) error
	DecrementFollowingCount(userID string) error

	ExecuteInTransaction(fn func(tx UserRepository) error) error

	SearchUsers(query, filter string, page, limit int) ([]*model.User, int, error)
	GetRecommendedUsers(limit int, excludeUserID string) ([]*model.User, error)
	GetAllUsers(page, limit int, sortBy string, ascending bool) ([]*model.User, int, error)

	BlockUser(blockerID, blockedID string) error
	UnblockUser(unblockerID, unblockedID string) error
	IsUserBlocked(userID, blockedByID string) (bool, error)
	ReportUser(reporterID, reportedID, reason string) error
	GetBlockedUsers(userID string, page, limit int) ([]map[string]interface{}, int64, error)
	
	GetDB() *gorm.DB
}

type PostgresUserRepository struct {
	db *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *PostgresUserRepository) FindUserByID(id string) (*model.User, error) {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	var user model.User
	if err := r.db.Where("id = ?", userUUID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) FindUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *PostgresUserRepository) DeleteUser(id string) error {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}
	return r.db.Delete(&model.User{}, "id = ?", userUUID).Error
}

func (r *PostgresUserRepository) UpdateUserVerification(userID string, isVerified bool) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	return r.db.Model(&model.User{}).
		Where("id = ?", userUUID).
		Update("is_verified", isVerified).
		Error
}

func (r *PostgresUserRepository) UserExists(userID string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID format: %w", err)
	}

	var count int64
	result := r.db.Model(&model.User{}).
		Where("id = ?", userUUID).
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("error checking user existence: %w", result.Error)
	}

	return count > 0, nil
}

func (r *PostgresUserRepository) CreateFollow(follow *model.Follow) error {
	return r.db.Create(follow).Error
}

func (r *PostgresUserRepository) DeleteFollow(followerID, followedID uuid.UUID) error {
	return r.db.Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Delete(&model.Follow{}).
		Error
}

func (r *PostgresUserRepository) CheckFollowExists(followerID, followedID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Follow{}).
		Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Count(&count).
		Error
	return count > 0, err
}

func (r *PostgresUserRepository) GetFollowers(userID uuid.UUID, page, limit int) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	var followers []*model.User
	var total int64

	fmt.Printf("DEBUG GetFollowers: Checking followers for userID %s, page %d, limit %d\n", userID.String(), page, limit)

	err := r.db.Model(&model.Follow{}).
		Where("followed_id = ?", userID).
		Count(&total).
		Error
	if err != nil {
		fmt.Printf("DEBUG GetFollowers: Error counting followers: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("DEBUG GetFollowers: Found %d total followers for user %s\n", total, userID.String())

	if total == 0 {

		return []*model.User{}, 0, nil
	}

	query := r.db.Model(&model.User{}).
		Joins("JOIN follows ON users.id = follows.follower_id").
		Where("follows.followed_id = ?", userID).
		Offset(offset).
		Limit(limit)

	sqlDB := r.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return query
	})
	fmt.Printf("DEBUG GetFollowers SQL: %s\n", sqlDB)

	err = query.Find(&followers).Error

	if err != nil {
		fmt.Printf("DEBUG GetFollowers: Error fetching followers: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("DEBUG GetFollowers: Returning %d followers (of %d total) for user %s\n", len(followers), total, userID.String())

	return followers, int(total), err
}

func (r *PostgresUserRepository) GetFollowing(userID uuid.UUID, page, limit int) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	var following []*model.User
	var total int64

	fmt.Printf("DEBUG GetFollowing: Checking following for userID %s, page %d, limit %d\n", userID.String(), page, limit)

	err := r.db.Model(&model.Follow{}).
		Where("follower_id = ?", userID).
		Count(&total).
		Error
	if err != nil {
		fmt.Printf("DEBUG GetFollowing: Error counting following: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("DEBUG GetFollowing: Found %d total following for user %s\n", total, userID.String())

	if total == 0 {

		return []*model.User{}, 0, nil
	}

	query := r.db.Model(&model.User{}).
		Joins("JOIN follows ON users.id = follows.followed_id").
		Where("follows.follower_id = ?", userID).
		Offset(offset).
		Limit(limit)

	sqlDB := r.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return query
	})
	fmt.Printf("DEBUG GetFollowing SQL: %s\n", sqlDB)

	err = query.Find(&following).Error

	if err != nil {
		fmt.Printf("DEBUG GetFollowing: Error fetching following: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("DEBUG GetFollowing: Returning %d following (of %d total) for user %s\n", len(following), total, userID.String())

	return following, int(total), err
}

func (r *PostgresUserRepository) SearchUsers(query, filter string, page, limit int) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	var users []*model.User
	var total int64

	db := r.db.Model(&model.User{}).
		Where("username ILIKE ? OR name ILIKE ?", "%"+query+"%", "%"+query+"%")

	if filter == "verified" {
		db = db.Where("is_verified = ?", true)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Offset(offset).Limit(limit).Find(&users).Error

	return users, int(total), err
}

func (r *PostgresUserRepository) GetRecommendedUsers(limit int, excludeUserID string) ([]*model.User, error) {
	var users []*model.User

	query := r.db.Model(&model.User{})

	if excludeUserID != "" {
		userUUID, err := uuid.Parse(excludeUserID)
		if err == nil {
			query = query.Where("id != ?", userUUID)
		}
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Find(&users).
		Error

	return users, err
}

func (r *PostgresUserRepository) GetAllUsers(page, limit int, sortBy string, ascending bool) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	var users []*model.User
	var total int64

	err := r.db.Model(&model.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if sortBy == "" {
		sortBy = "created_at"
	}

	sortDirection := "DESC"
	if ascending {
		sortDirection = "ASC"
	}

	err = r.db.Model(&model.User{}).
		Order(fmt.Sprintf("%s %s", sortBy, sortDirection)).
		Offset(offset).
		Limit(limit).
		Find(&users).
		Error

	return users, int(total), err
}

func (r *PostgresUserRepository) BlockUser(blockerID, blockedID string) error {
	blockerUUID, err := uuid.Parse(blockerID)
	if err != nil {
		return fmt.Errorf("invalid blocker ID: %w", err)
	}

	blockedUUID, err := uuid.Parse(blockedID)
	if err != nil {
		return fmt.Errorf("invalid blocked ID: %w", err)
	}

	var existingBlock model.UserBlock
	if err := r.db.Where("blocker_id = ? AND blocked_id = ?", blockerUUID, blockedUUID).First(&existingBlock).Error; err == nil {

		return nil
	}

	block := model.UserBlock{
		BlockerID: blockerUUID,
		BlockedID: blockedUUID,
	}

	return r.db.Create(&block).Error
}

func (r *PostgresUserRepository) UnblockUser(unblockerID, unblockedID string) error {
	unblockerUUID, err := uuid.Parse(unblockerID)
	if err != nil {
		return fmt.Errorf("invalid unblocker ID: %w", err)
	}

	unblockedUUID, err := uuid.Parse(unblockedID)
	if err != nil {
		return fmt.Errorf("invalid unblocked ID: %w", err)
	}

	return r.db.Where("blocker_id = ? AND blocked_id = ?", unblockerUUID, unblockedUUID).Delete(&model.UserBlock{}).Error
}

func (r *PostgresUserRepository) IsUserBlocked(userID, blockedByID string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID: %w", err)
	}

	blockedByUUID, err := uuid.Parse(blockedByID)
	if err != nil {
		return false, fmt.Errorf("invalid blocked by ID: %w", err)
	}

	var count int64
	err = r.db.Model(&model.UserBlock{}).
		Where("blocker_id = ? AND blocked_id = ?", blockedByUUID, userUUID).
		Count(&count).Error

	return count > 0, err
}

func (r *PostgresUserRepository) ReportUser(reporterID, reportedID, reason string) error {
	reporterUUID, err := uuid.Parse(reporterID)
	if err != nil {
		return fmt.Errorf("invalid reporter ID: %w", err)
	}

	reportedUUID, err := uuid.Parse(reportedID)
	if err != nil {
		return fmt.Errorf("invalid reported ID: %w", err)
	}

	var existingReport model.UserReport
	if err := r.db.Where("reporter_id = ? AND reported_id = ? AND status = ?", reporterUUID, reportedUUID, "pending").First(&existingReport).Error; err == nil {

		return nil
	}

	report := model.UserReport{
		ReporterID: reporterUUID,
		ReportedID: reportedUUID,
		Reason:     reason,
		Status:     "pending",
	}

	return r.db.Create(&report).Error
}

func (r *PostgresUserRepository) GetBlockedUsers(userID string, page, limit int) ([]map[string]interface{}, int64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user ID: %w", err)
	}

	var total int64
	var blockedUsers []model.User

	err = r.db.Model(&model.UserBlock{}).
		Where("blocker_id = ?", userUUID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = r.db.Table("users").
		Select("users.*").
		Joins("INNER JOIN user_blocks ON user_blocks.blocked_id = users.id").
		Where("user_blocks.blocker_id = ?", userUUID).
		Offset(offset).
		Limit(limit).
		Find(&blockedUsers).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, len(blockedUsers))
	for i, user := range blockedUsers {
		result[i] = map[string]interface{}{
			"id":       user.ID.String(),
			"username": user.Username,
			"name":     user.Name,
			"email":    user.Email,
		}
	}

	return result, total, nil
}

func (r *PostgresUserRepository) IncrementFollowerCount(userID string) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("follower_count", gorm.Expr("follower_count + 1")).
		Error
}

func (r *PostgresUserRepository) DecrementFollowerCount(userID string) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("follower_count", gorm.Expr("GREATEST(follower_count - 1, 0)")).
		Error
}

func (r *PostgresUserRepository) IncrementFollowingCount(userID string) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("following_count", gorm.Expr("following_count + 1")).
		Error
}

func (r *PostgresUserRepository) DecrementFollowingCount(userID string) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("following_count", gorm.Expr("GREATEST(following_count - 1, 0)")).
		Error
}

func (r *PostgresUserRepository) ExecuteInTransaction(fn func(tx UserRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &PostgresUserRepository{db: tx}
		return fn(txRepo)
	})
}

func (r *PostgresUserRepository) GetDB() *gorm.DB {
	return r.db
}
