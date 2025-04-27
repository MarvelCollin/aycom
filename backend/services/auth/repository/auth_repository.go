package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// User represents a user in the system
type User struct {
	ID                  string
	Name                string
	Username            string
	Email               string
	PasswordHash        string
	Gender              string
	DateOfBirth         time.Time
	ProfilePicturePath  string
	BannerPath          string
	SecurityQuestion    string
	SecurityAnswerHash  string
	SubscribeNewsletter bool
	EmailVerified       bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// VerificationCode represents an email verification code
type VerificationCode struct {
	Email     string
	Code      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// AuthRepository defines the interface for auth data operations
type AuthRepository interface {
	// Token operations
	StoreRefreshToken(ctx context.Context, userID, refreshToken string, expiresAt time.Time) error
	GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (string, error)
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
	IsTokenRevoked(ctx context.Context, refreshToken string) (bool, error)

	// User operations
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	VerifyUserEmail(ctx context.Context, email string) error

	// Verification code operations
	StoreVerificationCode(ctx context.Context, email, code string) error
	GetVerificationCode(ctx context.Context, email string) (*VerificationCode, error)
	DeleteVerificationCode(ctx context.Context, email string) error
}

// authRepository implements the AuthRepository interface
type authRepository struct {
	db *sql.DB
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

// StoreRefreshToken stores a new refresh token for a user
func (r *authRepository) StoreRefreshToken(ctx context.Context, userID, refreshToken string, expiresAt time.Time) error {
	query := `
		INSERT INTO auth_tokens (user_id, refresh_token, expires_at, created_at, revoked)
		VALUES ($1, $2, $3, $4, false)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		userID,
		refreshToken,
		expiresAt,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

// GetUserIDByRefreshToken retrieves the user ID associated with a refresh token
func (r *authRepository) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	query := `
		SELECT user_id FROM auth_tokens 
		WHERE refresh_token = $1 
		  AND expires_at > $2
		  AND revoked = false
	`

	var userID string
	err := r.db.QueryRowContext(ctx, query, refreshToken, time.Now()).Scan(&userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("token not found or expired")
		}
		return "", fmt.Errorf("failed to get user ID: %w", err)
	}

	return userID, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *authRepository) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	query := `UPDATE auth_tokens SET revoked = true WHERE refresh_token = $1`

	result, err := r.db.ExecContext(ctx, query, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (r *authRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	query := `UPDATE auth_tokens SET revoked = true WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all tokens: %w", err)
	}

	return nil
}

// IsTokenRevoked checks if a refresh token is revoked
func (r *authRepository) IsTokenRevoked(ctx context.Context, refreshToken string) (bool, error) {
	query := `SELECT revoked FROM auth_tokens WHERE refresh_token = $1`

	var revoked bool
	err := r.db.QueryRowContext(ctx, query, refreshToken).Scan(&revoked)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil // Token doesn't exist, consider it revoked
		}
		return true, fmt.Errorf("failed to check token status: %w", err)
	}

	return revoked, nil
}

// CreateUser creates a new user in the database
func (r *authRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (
			id, name, username, email, password_hash, gender, date_of_birth, 
			profile_picture_path, banner_path, security_question, security_answer_hash,
			subscribe_newsletter, email_verified, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Name,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Gender,
		user.DateOfBirth,
		user.ProfilePicturePath,
		user.BannerPath,
		user.SecurityQuestion,
		user.SecurityAnswerHash,
		user.SubscribeNewsletter,
		user.EmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT 
			id, name, username, email, password_hash, gender, date_of_birth, 
			profile_picture_path, banner_path, security_question, security_answer_hash,
			subscribe_newsletter, email_verified, created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Gender,
		&user.DateOfBirth,
		&user.ProfilePicturePath,
		&user.BannerPath,
		&user.SecurityQuestion,
		&user.SecurityAnswerHash,
		&user.SubscribeNewsletter,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (r *authRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT 
			id, name, username, email, password_hash, gender, date_of_birth, 
			profile_picture_path, banner_path, security_question, security_answer_hash,
			subscribe_newsletter, email_verified, created_at, updated_at
		FROM users 
		WHERE username = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Name,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Gender,
		&user.DateOfBirth,
		&user.ProfilePicturePath,
		&user.BannerPath,
		&user.SecurityQuestion,
		&user.SecurityAnswerHash,
		&user.SubscribeNewsletter,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (r *authRepository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users SET
			name = $1,
			username = $2,
			email = $3,
			password_hash = $4,
			gender = $5,
			date_of_birth = $6,
			profile_picture_path = $7,
			banner_path = $8,
			security_question = $9,
			security_answer_hash = $10,
			subscribe_newsletter = $11,
			email_verified = $12,
			updated_at = $13
		WHERE id = $14
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Name,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Gender,
		user.DateOfBirth,
		user.ProfilePicturePath,
		user.BannerPath,
		user.SecurityQuestion,
		user.SecurityAnswerHash,
		user.SubscribeNewsletter,
		user.EmailVerified,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// VerifyUserEmail marks a user's email as verified
func (r *authRepository) VerifyUserEmail(ctx context.Context, email string) error {
	query := `UPDATE users SET email_verified = true, updated_at = $1 WHERE email = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), email)
	if err != nil {
		return fmt.Errorf("failed to verify user email: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// StoreVerificationCode stores a verification code for email verification
func (r *authRepository) StoreVerificationCode(ctx context.Context, email, code string) error {
	// First, delete any existing verification codes for this email
	deleteQuery := `DELETE FROM verification_codes WHERE email = $1`
	_, err := r.db.ExecContext(ctx, deleteQuery, email)
	if err != nil {
		return fmt.Errorf("failed to delete existing verification codes: %w", err)
	}

	// Then, insert the new verification code
	insertQuery := `
		INSERT INTO verification_codes (email, code, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`

	// Code expires in 5 minutes
	expiresAt := time.Now().Add(5 * time.Minute)
	createdAt := time.Now()

	_, err = r.db.ExecContext(
		ctx,
		insertQuery,
		email,
		code,
		expiresAt,
		createdAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	return nil
}

// GetVerificationCode retrieves a verification code for an email
func (r *authRepository) GetVerificationCode(ctx context.Context, email string) (*VerificationCode, error) {
	query := `
		SELECT email, code, expires_at, created_at
		FROM verification_codes
		WHERE email = $1 AND expires_at > $2
	`

	var code VerificationCode
	err := r.db.QueryRowContext(ctx, query, email, time.Now()).Scan(
		&code.Email,
		&code.Code,
		&code.ExpiresAt,
		&code.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("verification code not found or expired")
		}
		return nil, fmt.Errorf("failed to get verification code: %w", err)
	}

	return &code, nil
}

// DeleteVerificationCode deletes a verification code for an email
func (r *authRepository) DeleteVerificationCode(ctx context.Context, email string) error {
	query := `DELETE FROM verification_codes WHERE email = $1`

	_, err := r.db.ExecContext(ctx, query, email)
	if err != nil {
		return fmt.Errorf("failed to delete verification code: %w", err)
	}

	return nil
}
