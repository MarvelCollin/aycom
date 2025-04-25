package repository

import (
	"context"

	"github.com/Acad600-TPA/WEB-MV-242/auth/internal/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// FindByID retrieves a user by ID
	FindByID(ctx context.Context, id string) (*models.User, error)

	// FindByEmail retrieves a user by email
	FindByEmail(ctx context.Context, email string) (*models.User, error)

	// Create saves a new user
	Create(ctx context.Context, user *models.User) error

	// Update updates an existing user
	Update(ctx context.Context, user *models.User) error

	// Delete removes a user
	Delete(ctx context.Context, id string) error
}
