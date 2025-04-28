package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthRepository defines the methods for auth-related database operations
type AuthRepository interface {
	// User methods
	FindUserByEmail(email string) (*User, error)
	FindUserByID(id string) (*User, error)
	FindUserByUsername(username string) (*User, error)
	SaveUser(user *User) error
	UpdateUser(user *User) error
	DeleteUser(id string) error

	// Token methods
	SaveToken(token *Token) error
	FindTokenByUserID(userID string) (*Token, error)
	DeleteToken(userID string, refreshToken string) error

	// OAuth methods
	SaveOAuthConnection(conn *OAuthConnection) error
	FindOAuthConnection(provider, providerID string) (*OAuthConnection, error)
}

// PostgresAuthRepository is the PostgreSQL implementation of AuthRepository
type PostgresAuthRepository struct {
	db *gorm.DB
}

// NewPostgresAuthRepository creates a new PostgreSQL auth repository
func NewPostgresAuthRepository(db *gorm.DB) AuthRepository {
	return &PostgresAuthRepository{db: db}
}

// FindUserByEmail finds a user by their email
func (r *PostgresAuthRepository) FindUserByEmail(email string) (*User, error) {
	var user User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindUserByID finds a user by their ID
func (r *PostgresAuthRepository) FindUserByID(id string) (*User, error) {
	var user User
	result := r.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindUserByUsername finds a user by their username
func (r *PostgresAuthRepository) FindUserByUsername(username string) (*User, error) {
	var user User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

// SaveUser creates a new user
func (r *PostgresAuthRepository) SaveUser(user *User) error {
	return r.db.Create(user).Error
}

// UpdateUser updates an existing user
func (r *PostgresAuthRepository) UpdateUser(user *User) error {
	return r.db.Save(user).Error
}

// DeleteUser deletes a user by their ID
func (r *PostgresAuthRepository) DeleteUser(id string) error {
	return r.db.Delete(&User{}, "id = ?", id).Error
}

// SaveToken creates a new token
func (r *PostgresAuthRepository) SaveToken(token *Token) error {
	// Generate UUID for token ID if not provided
	if token.ID == "" {
		token.ID = uuid.New().String()
	}

	// Delete any existing tokens for this user first
	if err := r.db.Where("user_id = ?", token.UserID).Delete(&Token{}).Error; err != nil {
		return err
	}

	return r.db.Create(token).Error
}

// FindTokenByUserID finds a token by user ID
func (r *PostgresAuthRepository) FindTokenByUserID(userID string) (*Token, error) {
	var token Token
	result := r.db.Where("user_id = ?", userID).First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("token not found")
		}
		return nil, result.Error
	}
	return &token, nil
}

// DeleteToken deletes a token by user ID and refresh token
func (r *PostgresAuthRepository) DeleteToken(userID string, refreshToken string) error {
	return r.db.Where("user_id = ? AND refresh_token = ?", userID, refreshToken).Delete(&Token{}).Error
}

// SaveOAuthConnection creates a new OAuth connection
func (r *PostgresAuthRepository) SaveOAuthConnection(conn *OAuthConnection) error {
	// Generate UUID for connection ID if not provided
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	return r.db.Create(conn).Error
}

// FindOAuthConnection finds an OAuth connection by provider and provider ID
func (r *PostgresAuthRepository) FindOAuthConnection(provider, providerID string) (*OAuthConnection, error) {
	var conn OAuthConnection
	result := r.db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&conn)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("OAuth connection not found")
		}
		return nil, result.Error
	}
	return &conn, nil
}
