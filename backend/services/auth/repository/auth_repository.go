package repository

import (
	"errors"
	"time" // Added for Token struct

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the user model in the database
type User struct {
	ID                    uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name                  string    `gorm:"size:255;not null"`
	Username              string    `gorm:"size:255;uniqueIndex;not null"`
	Email                 string    `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash          string    `gorm:"size:255;not null"`
	Gender                string    `gorm:"size:50"`
	DateOfBirth           string    `gorm:"type:date"` // Store as string YYYY-MM-DD
	SecurityQuestion      string    `gorm:"size:255"`
	SecurityAnswer        string    `gorm:"size:255"` // Consider hashing
	EmailVerified         bool      `gorm:"default:false"`
	VerificationCode      string    `gorm:"size:10"`
	VerificationExpiresAt time.Time
	SubscribeToNewsletter bool `gorm:"default:false"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
	// Add other fields like ProfilePictureURL, BannerURL if managed here
	// Or link to User service data
}

// Token represents the refresh token model in the database (Optional, for stateful refresh tokens)
type Token struct {
	ID           string    `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"` // Changed to string UUID
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`                         // Foreign key to User
	RefreshToken string    `gorm:"type:text;not null;uniqueIndex"`
	ExpiresAt    time.Time `gorm:"not null"`
	CreatedAt    time.Time
	User         User `gorm:"foreignKey:UserID"` // Define relationship
}

// OAuthConnection represents third-party OAuth connections
type OAuthConnection struct {
	ID         string    `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"` // Changed to string UUID
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`                         // Foreign key to User
	Provider   string    `gorm:"size:50;not null;index:idx_oauth_provider_id"`
	ProviderID string    `gorm:"size:255;not null;index:idx_oauth_provider_id"` // User ID from the provider
	// Store provider tokens (access, refresh) if needed
	CreatedAt time.Time
	UpdatedAt time.Time
	User      User `gorm:"foreignKey:UserID"` // Define relationship
}

// AuthRepository defines the methods for auth-related database operations
type AuthRepository interface {
	// User methods
	FindUserByEmail(email string) (*User, error)
	FindUserByID(id uuid.UUID) (*User, error) // Changed id to uuid.UUID
	FindUserByUsername(username string) (*User, error)
	SaveUser(user *User) error
	UpdateUser(user *User) error
	DeleteUser(id uuid.UUID) error // Changed id to uuid.UUID

	// Token methods (for stateful refresh tokens)
	SaveToken(token *Token) error
	FindTokenByRefreshToken(refreshToken string) (*Token, error) // Find by token itself
	DeleteToken(refreshToken string) error                       // Delete by token itself

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
			// Return specific error for not found
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindUserByID finds a user by their ID
func (r *PostgresAuthRepository) FindUserByID(id uuid.UUID) (*User, error) { // Changed id to uuid.UUID
	var user User
	result := r.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
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
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// SaveUser creates a new user
func (r *PostgresAuthRepository) SaveUser(user *User) error {
	// Ensure ID is set if not already
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return r.db.Create(user).Error
}

// UpdateUser updates an existing user
func (r *PostgresAuthRepository) UpdateUser(user *User) error {
	// Use Save for full updates or Updates for partial updates
	return r.db.Save(user).Error
}

// DeleteUser deletes a user by their ID
func (r *PostgresAuthRepository) DeleteUser(id uuid.UUID) error { // Changed id to uuid.UUID
	return r.db.Delete(&User{}, "id = ?", id).Error
}

// --- Token Methods (for stateful refresh tokens) ---

// SaveToken creates or updates a refresh token
func (r *PostgresAuthRepository) SaveToken(token *Token) error {
	// Ensure ID is set if not already
	if token.ID == "" {
		token.ID = uuid.New().String()
	}
	// Use Save to handle both create and update based on primary key (ID)
	// Or use Upsert/OnConflict if needed, depending on DB and GORM version
	return r.db.Save(token).Error
}

// FindTokenByRefreshToken finds a token by the refresh token string
func (r *PostgresAuthRepository) FindTokenByRefreshToken(refreshToken string) (*Token, error) {
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

// DeleteToken deletes a token by its refresh token string
func (r *PostgresAuthRepository) DeleteToken(refreshToken string) error {
	// Delete based on the unique refresh token string
	result := r.db.Where("refresh_token = ?", refreshToken).Delete(&Token{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// Optional: Return an error if the token to be deleted wasn't found
		// return gorm.ErrRecordNotFound
	}
	return nil
}

// --- OAuth Methods ---

// SaveOAuthConnection creates a new OAuth connection
func (r *PostgresAuthRepository) SaveOAuthConnection(conn *OAuthConnection) error {
	// Ensure ID is set if not already
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	return r.db.Create(conn).Error
}

// FindOAuthConnection finds an OAuth connection by provider and provider ID
func (r *PostgresAuthRepository) FindOAuthConnection(provider, providerID string) (*OAuthConnection, error) {
	var conn OAuthConnection
	// Use the composite index defined in the struct tag
	result := r.db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&conn)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound // Return specific error
		}
		return nil, result.Error
	}
	return &conn, nil
}
