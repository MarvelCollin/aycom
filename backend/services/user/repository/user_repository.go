package repository

import (
	"fmt"

	"aycom/backend/services/user/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository defines methods for user data access
type UserRepository interface {
	// User CRUD operations
	CreateUser(user *model.User) error
	FindUserByID(id string) (*model.User, error)
	FindUserByUsername(username string) (*model.User, error)
	FindUserByEmail(email string) (*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(id string) error
	UpdateUserVerification(userID string, isVerified bool) error

	// Check if user exists by ID
	UserExists(userID string) (bool, error)

	// Follow-related operations
	CreateFollow(follow *model.Follow) error
	DeleteFollow(followerID, followedID uuid.UUID) error
	CheckFollowExists(followerID, followedID uuid.UUID) (bool, error)
	GetFollowers(userID uuid.UUID, page, limit int) ([]*model.User, int, error)
	GetFollowing(userID uuid.UUID, page, limit int) ([]*model.User, int, error)

	// User search and listing
	SearchUsers(query, filter string, page, limit int) ([]*model.User, int, error)
	GetRecommendedUsers(limit int, excludeUserID string) ([]*model.User, error)
	GetAllUsers(page, limit int, sortBy string, ascending bool) ([]*model.User, int, error)
}

// PostgresUserRepository implements UserRepository with PostgreSQL
type PostgresUserRepository struct {
	db *gorm.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *PostgresUserRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

// FindUserByID finds a user by ID
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

// FindUserByUsername finds a user by username
func (r *PostgresUserRepository) FindUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByEmail finds a user by email
func (r *PostgresUserRepository) FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user
func (r *PostgresUserRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

// DeleteUser deletes a user by ID
func (r *PostgresUserRepository) DeleteUser(id string) error {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}
	return r.db.Delete(&model.User{}, "id = ?", userUUID).Error
}

// UpdateUserVerification updates a user's verification status
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

// UserExists checks if a user exists by ID
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

// CreateFollow creates a new follow relationship
func (r *PostgresUserRepository) CreateFollow(follow *model.Follow) error {
	return r.db.Create(follow).Error
}

// DeleteFollow removes a follow relationship
func (r *PostgresUserRepository) DeleteFollow(followerID, followedID uuid.UUID) error {
	return r.db.Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Delete(&model.Follow{}).
		Error
}

// CheckFollowExists checks if a follow relationship exists
func (r *PostgresUserRepository) CheckFollowExists(followerID, followedID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Follow{}).
		Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Count(&count).
		Error
	return count > 0, err
}

// GetFollowers gets users who follow a specific user
func (r *PostgresUserRepository) GetFollowers(userID uuid.UUID, page, limit int) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	var followers []*model.User
	var total int64

	// Count total followers
	err := r.db.Model(&model.Follow{}).
		Where("followed_id = ?", userID).
		Count(&total).
		Error
	if err != nil {
		return nil, 0, err
	}

	// Get followers with pagination
	err = r.db.Model(&model.User{}).
		Joins("JOIN follows ON users.id = follows.follower_id").
		Where("follows.followed_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&followers).
		Error

	return followers, int(total), err
}

// GetFollowing gets users that a specific user follows
func (r *PostgresUserRepository) GetFollowing(userID uuid.UUID, page, limit int) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	var following []*model.User
	var total int64

	// Count total following
	err := r.db.Model(&model.Follow{}).
		Where("follower_id = ?", userID).
		Count(&total).
		Error
	if err != nil {
		return nil, 0, err
	}

	// Get following with pagination
	err = r.db.Model(&model.User{}).
		Joins("JOIN follows ON users.id = follows.followed_id").
		Where("follows.follower_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&following).
		Error

	return following, int(total), err
}

// SearchUsers searches for users based on a query
func (r *PostgresUserRepository) SearchUsers(query, filter string, page, limit int) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	var users []*model.User
	var total int64

	db := r.db.Model(&model.User{}).
		Where("username ILIKE ? OR name ILIKE ?", "%"+query+"%", "%"+query+"%")

	// Apply filters if provided
	if filter == "verified" {
		db = db.Where("is_verified = ?", true)
	}

	// Get total count
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get users with pagination
	err = db.Offset(offset).Limit(limit).Find(&users).Error

	return users, int(total), err
}

// GetRecommendedUsers gets recommended users (e.g., users with most followers)
func (r *PostgresUserRepository) GetRecommendedUsers(limit int, excludeUserID string) ([]*model.User, error) {
	var users []*model.User

	query := r.db.Model(&model.User{})

	// Exclude the current user if ID is provided
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

// GetAllUsers gets all users with pagination and sorting
func (r *PostgresUserRepository) GetAllUsers(page, limit int, sortBy string, ascending bool) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	var users []*model.User
	var total int64

	// Count total users
	err := r.db.Model(&model.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Set default sort field if not provided
	if sortBy == "" {
		sortBy = "created_at"
	}

	// Determine sort direction
	sortDirection := "DESC"
	if ascending {
		sortDirection = "ASC"
	}

	// Get users with pagination and sorting
	err = r.db.Model(&model.User{}).
		Order(fmt.Sprintf("%s %s", sortBy, sortDirection)).
		Offset(offset).
		Limit(limit).
		Find(&users).
		Error

	return users, int(total), err
}
