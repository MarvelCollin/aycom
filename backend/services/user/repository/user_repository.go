package repository

import (
	"errors"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/model"
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
			return nil, errors.New("user not found")
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
			return nil, errors.New("user not found")
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
			return nil, errors.New("user not found")
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
