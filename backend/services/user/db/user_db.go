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
	db *gorm.DB
}

func NewPostgresUserAuthRepository(db *gorm.DB) UserAuthRepository {
	return &PostgresUserAuthRepository{db: db}
}

// UserRepository methods (reuse existing PostgresUserRepository logic)
// ... existing code ...
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
