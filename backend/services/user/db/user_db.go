package db

import (
	"errors"
	"log"
	"time"

	"aycom/backend/services/user/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *model.User) error
	FindUserByID(id string) (*model.User, error)
	FindUserByEmail(email string) (*model.User, error)
	FindUserByUsername(username string) (*model.User, error)
	UpdateUser(user *model.User) error
	UpdateUserVerification(userID string, isVerified bool) error
	DeleteUser(id string) error

	CheckFollowExists(followerID, followedID uuid.UUID) (bool, error)
	CreateFollow(follow *model.Follow) error
	DeleteFollow(followerID, followedID uuid.UUID) error
	GetFollowers(userID uuid.UUID, page, limit int) ([]*model.User, int, error)
	GetFollowing(userID uuid.UUID, page, limit int) ([]*model.User, int, error)
	SearchUsers(query, filter string, page, limit int) ([]*model.User, int, error)
	GetRecommendedUsers(limit int, excludeUserID string) ([]*model.User, error)
	GetAllUsers(page, limit int, sortBy string, ascending bool) ([]*model.User, int, error)
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

func (r *PostgresUserRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *PostgresUserRepository) UpdateUserVerification(userID string, isVerified bool) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("is_verified", isVerified).Error
}

func (r *PostgresUserRepository) DeleteUser(id string) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

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

func (r *PostgresUserRepository) CreateFollow(follow *model.Follow) error {
	return r.db.Create(follow).Error
}

func (r *PostgresUserRepository) DeleteFollow(followerID, followedID uuid.UUID) error {
	return r.db.Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Delete(&model.Follow{}).Error
}

func (r *PostgresUserRepository) GetFollowers(userID uuid.UUID, page, limit int) ([]*model.User, int, error) {
	var followers []*model.User
	var total int64

	offset := (page - 1) * limit

	err := r.db.Model(&model.Follow{}).
		Where("followed_id = ?", userID).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

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

func (r *PostgresUserRepository) GetFollowing(userID uuid.UUID, page, limit int) ([]*model.User, int, error) {
	var following []*model.User
	var total int64

	offset := (page - 1) * limit

	err := r.db.Model(&model.Follow{}).
		Where("follower_id = ?", userID).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

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

func (r *PostgresUserRepository) SearchUsers(query, filter string, page, limit int) ([]*model.User, int, error) {
	var users []*model.User
	var total int64

	offset := (page - 1) * limit

	if filter == "popular" {
		countQuery := r.db.Table("users AS u").
			Select("u.*, COUNT(f.follower_id) as follower_count").
			Joins("LEFT JOIN follows AS f ON u.id = f.followed_id").
			Group("u.id").
			Order("follower_count DESC, u.created_at DESC")

		var tempUsers []*model.User
		err := countQuery.Find(&tempUsers).Error
		if err != nil {
			return nil, 0, err
		}
		total = int64(len(tempUsers))

		err = countQuery.
			Offset(offset).
			Limit(limit).
			Find(&users).Error

		if err != nil {
			return nil, 0, err
		}

		return users, int(total), nil
	}

	baseQuery := r.db.Model(&model.User{})

	if query != "" {
		baseQuery = baseQuery.Where("username ILIKE ? OR name ILIKE ? OR email ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	switch filter {
	case "verified":
		baseQuery = baseQuery.Where("is_verified = ?", true)
	}

	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = baseQuery.
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}

func (r *PostgresUserRepository) GetRecommendedUsers(limit int, excludeUserID string) ([]*model.User, error) {
	var users []*model.User

	// Check if follows table exists
	var hasFollowsTable bool
	err := r.db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'follows')").Scan(&hasFollowsTable).Error
	if err != nil {
		log.Printf("Error checking if follows table exists: %v", err)
		// Continue with a fallback approach
		hasFollowsTable = false
	}

	query := r.db.Model(&model.User{})

	// If follows table exists, use it to rank users by follower count
	if hasFollowsTable {
		query = r.db.Table("users").
			Joins("LEFT JOIN follows ON users.id = follows.followed_id").
			Group("users.id").
			Select("users.*, COUNT(follows.follower_id) as follower_count").
			Order("follower_count DESC, users.created_at DESC")
	} else {
		// Fallback to sorting by creation date if follows table doesn't exist
		query = query.Order("created_at DESC")
	}

	// Exclude the current user if ID is provided
	if excludeUserID != "" {
		if _, err := uuid.Parse(excludeUserID); err == nil {
			query = query.Where("users.id != ?", excludeUserID)
		}
	}

	// Apply limit
	err = query.Limit(limit).Find(&users).Error
	if err != nil {
		log.Printf("Error retrieving recommended users: %v", err)
		return nil, err
	}

	// If we didn't get any users and follows table exists, try the fallback method
	if len(users) == 0 && hasFollowsTable {
		log.Printf("No users found with joins method, trying fallback")
		fallbackQuery := r.db.Model(&model.User{})

		// Apply exclude filter if we have a valid UUID
		if excludeUserID != "" {
			if _, err := uuid.Parse(excludeUserID); err == nil {
				fallbackQuery = fallbackQuery.Where("id != ?", excludeUserID)
			}
		}

		err = fallbackQuery.Order("created_at DESC").
			Limit(limit).
			Find(&users).Error

		if err != nil {
			log.Printf("Error retrieving recommended users with fallback: %v", err)
			return nil, err
		}
	}

	return users, nil
}

func (r *PostgresUserRepository) GetAllUsers(page, limit int, sortBy string, ascending bool) ([]*model.User, int, error) {
	var users []*model.User
	var total int64

	offset := (page - 1) * limit

	// Count total users
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Build the query with sorting
	query := r.db.Model(&model.User{})

	// Sort by specified field or default to created_at descending
	if sortBy != "" {
		// Make sure the sort field exists to prevent SQL injection
		validSortFields := map[string]bool{
			"username": true, "created_at": true, "name": true, "id": true,
		}

		if _, ok := validSortFields[sortBy]; ok {
			if ascending {
				query = query.Order(sortBy + " ASC")
			} else {
				query = query.Order(sortBy + " DESC")
			}
		} else {
			// Default sorting if invalid field provided
			query = query.Order("created_at DESC")
		}
	} else {
		// Default sorting if no sort field provided
		query = query.Order("created_at DESC")
	}

	// Execute the query with pagination
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}

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

type UserAuthRepository interface {
	UserRepository
	SaveToken(token *Token) error
	FindTokenByRefreshToken(refreshToken string) (*Token, error)
	DeleteToken(refreshToken string) error
	SaveOAuthConnection(conn *OAuthConnection) error
	FindOAuthConnection(provider, providerID string) (*OAuthConnection, error)
}

type PostgresUserAuthRepository struct {
	PostgresUserRepository
	db *gorm.DB
}

func NewPostgresUserAuthRepository(db *gorm.DB) UserAuthRepository {
	return &PostgresUserAuthRepository{
		PostgresUserRepository: PostgresUserRepository{db: db},
		db:                     db,
	}
}

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
