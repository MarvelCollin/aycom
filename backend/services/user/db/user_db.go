package db

import (
	"errors"
	"time"

	"aycom/backend/services/user/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository defines the methods for user-related database operations
type UserRepository interface {
	// User methods
	CreateUser(user *model.User) error
	FindUserByID(id string) (*model.User, error)
	FindUserByEmail(email string) (*model.User, error)
	FindUserByUsername(username string) (*model.User, error)
	UpdateUser(user *model.User) error
	UpdateUserVerification(userID string, isVerified bool) error
	DeleteUser(id string) error

	// Social follow methods
	CheckFollowExists(followerID, followedID uuid.UUID) (bool, error)
	CreateFollow(follow *model.Follow) error
	DeleteFollow(followerID, followedID uuid.UUID) error
	GetFollowers(userID uuid.UUID, page, limit int) ([]*model.User, int, error)
	GetFollowing(userID uuid.UUID, page, limit int) ([]*model.User, int, error)
	SearchUsers(query, filter string, page, limit int) ([]*model.User, int, error)
	GetRecommendedUsers(limit int, excludeUserID string) ([]*model.User, error)
}

// PostgresUserRepository is the PostgreSQL implementation of UserRepository
type PostgresUserRepository struct {
	db *gorm.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

// CreateUser creates a new user
func (r *PostgresUserRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

// FindUserByID finds a user by their ID
func (r *PostgresUserRepository) FindUserByID(id string) (*model.User, error) {
	// Check if id is a valid UUID
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid UUID format for user ID")
	}
	var user model.User
	result := r.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindUserByEmail finds a user by their email
func (r *PostgresUserRepository) FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindUserByUsername finds a user by their username
func (r *PostgresUserRepository) FindUserByUsername(username string) (*model.User, error) {
	var user model.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// UpdateUser updates an existing user
func (r *PostgresUserRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

// UpdateUserVerification updates the user's verification status
func (r *PostgresUserRepository) UpdateUserVerification(userID string, isVerified bool) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("is_verified", isVerified).Error
}

// DeleteUser deletes a user by their ID
func (r *PostgresUserRepository) DeleteUser(id string) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

// CheckFollowExists checks if a follow relationship exists between two users
func (r *PostgresUserRepository) CheckFollowExists(followerID, followedID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Follow{}).
		Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Count(&count).Error

	if err != nil {
		return false, err
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
		Delete(&model.Follow{}).Error
}

// GetFollowers gets a paginated list of users who follow the specified user
func (r *PostgresUserRepository) GetFollowers(userID uuid.UUID, page, limit int) ([]*model.User, int, error) {
	var followers []*model.User
	var total int64

	// Calculate offset
	offset := (page - 1) * limit

	// Get total count
	err := r.db.Model(&model.Follow{}).
		Where("followed_id = ?", userID).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	// Get followers with pagination
	err = r.db.Table("users").
		Joins("JOIN follows ON users.id = follows.follower_id").
		Where("follows.followed_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&followers).Error

	if err != nil {
		return nil, 0, err
	}

	return followers, int(total), nil
}

// GetFollowing gets a paginated list of users followed by the specified user
func (r *PostgresUserRepository) GetFollowing(userID uuid.UUID, page, limit int) ([]*model.User, int, error) {
	var following []*model.User
	var total int64

	// Calculate offset
	offset := (page - 1) * limit

	// Get total count
	err := r.db.Model(&model.Follow{}).
		Where("follower_id = ?", userID).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	// Get followed users with pagination
	err = r.db.Table("users").
		Joins("JOIN follows ON users.id = follows.followed_id").
		Where("follows.follower_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&following).Error

	if err != nil {
		return nil, 0, err
	}

	return following, int(total), nil
}

// SearchUsers searches for users based on query and filters with pagination
func (r *PostgresUserRepository) SearchUsers(query, filter string, page, limit int) ([]*model.User, int, error) {
	var users []*model.User
	var total int64

	// Calculate offset
	offset := (page - 1) * limit

	// Special case for 'popular' filter which gets most followed users
	if filter == "popular" {
		// This query gets users ordered by follower count
		countQuery := r.db.Table("users AS u").
			Select("u.*, COUNT(f.follower_id) as follower_count").
			Joins("LEFT JOIN follows AS f ON u.id = f.followed_id").
			Group("u.id").
			Order("follower_count DESC, u.created_at DESC")

		// Get total count
		var tempUsers []*model.User
		err := countQuery.Find(&tempUsers).Error
		if err != nil {
			return nil, 0, err
		}
		total = int64(len(tempUsers))

		// Apply pagination
		err = countQuery.
			Offset(offset).
			Limit(limit).
			Find(&users).Error

		if err != nil {
			return nil, 0, err
		}

		return users, int(total), nil
	}

	// Base query for standard search
	baseQuery := r.db.Model(&model.User{})

	// Apply query filter if provided
	if query != "" {
		baseQuery = baseQuery.Where("username ILIKE ? OR name ILIKE ? OR email ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	// Apply additional filters if needed
	switch filter {
	case "verified":
		baseQuery = baseQuery.Where("is_verified = ?", true)
		// Add more filters as needed
	}

	// Get total count
	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get users with pagination
	err = baseQuery.
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}

// GetRecommendedUsers gets users with the highest follower count
func (r *PostgresUserRepository) GetRecommendedUsers(limit int, excludeUserID string) ([]*model.User, error) {
	var users []*model.User

	// Start with a basic query to select users and count their followers
	query := r.db.Table("users").
		Joins("LEFT JOIN follows ON users.id = follows.followed_id").
		Group("users.id").
		Select("users.*, COUNT(follows.follower_id) as follower_count").
		Order("follower_count DESC, users.created_at DESC") // Sort by follower count, then registration date

	// Add exclusion if a userID is provided (to avoid recommending the current user)
	if excludeUserID != "" {
		if _, err := uuid.Parse(excludeUserID); err == nil {
			query = query.Where("users.id != ?", excludeUserID)
		}
	}

	// Apply limit
	err := query.Limit(limit).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Add Token and OAuthConnection structs

type Token struct {
	ID           string    `gorm:"type:uuid;primary_key"`
	UserID       string    `gorm:"type:uuid;not null;index"`
	RefreshToken string    `gorm:"type:text;not null;uniqueIndex"`
	ExpiresAt    time.Time `gorm:"not null"`
	CreatedAt    time.Time
}

type OAuthConnection struct {
	ID         string `gorm:"type:uuid;primary_key"`
	UserID     string `gorm:"type:uuid;not null;index"`
	Provider   string `gorm:"size:50;not null;index:idx_oauth_provider_id"`
	ProviderID string `gorm:"size:255;not null;index:idx_oauth_provider_id"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Add token and OAuth methods to the repository

type UserAuthRepository interface {
	UserRepository
	SaveToken(token *Token) error
	FindTokenByRefreshToken(refreshToken string) (*Token, error)
	DeleteToken(refreshToken string) error
	SaveOAuthConnection(conn *OAuthConnection) error
	FindOAuthConnection(provider, providerID string) (*OAuthConnection, error)
}

type PostgresUserAuthRepository struct {
	PostgresUserRepository // Embed PostgresUserRepository to inherit all its methods
	db                     *gorm.DB
}

func NewPostgresUserAuthRepository(db *gorm.DB) UserAuthRepository {
	return &PostgresUserAuthRepository{
		PostgresUserRepository: PostgresUserRepository{db: db},
		db:                     db,
	}
}

// Token methods
func (r *PostgresUserAuthRepository) SaveToken(token *Token) error {
	if token.ID == "" {
		token.ID = uuid.New().String()
	}
	return r.db.Save(token).Error
}

func (r *PostgresUserAuthRepository) FindTokenByRefreshToken(refreshToken string) (*Token, error) {
	var token Token
	result := r.db.Where("refresh_token = ?", refreshToken).First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &token, nil
}

func (r *PostgresUserAuthRepository) DeleteToken(refreshToken string) error {
	result := r.db.Where("refresh_token = ?", refreshToken).Delete(&Token{})
	return result.Error
}

// OAuth methods
func (r *PostgresUserAuthRepository) SaveOAuthConnection(conn *OAuthConnection) error {
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	return r.db.Create(conn).Error
}

func (r *PostgresUserAuthRepository) FindOAuthConnection(provider, providerID string) (*OAuthConnection, error) {
	var conn OAuthConnection
	result := r.db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&conn)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &conn, nil
}

func (r *PostgresUserAuthRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *PostgresUserAuthRepository) DeleteUser(id string) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

func (r *PostgresUserAuthRepository) FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *PostgresUserAuthRepository) FindUserByID(id string) (*model.User, error) {
	var user model.User
	result := r.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *PostgresUserAuthRepository) FindUserByUsername(username string) (*model.User, error) {
	var user model.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *PostgresUserAuthRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *PostgresUserAuthRepository) UpdateUserVerification(userID string, isVerified bool) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("is_verified", isVerified).Error
}
