package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// AuthRepository defines the interface for auth data operations
type AuthRepository interface {
	StoreRefreshToken(ctx context.Context, userID, refreshToken string, expiresAt time.Time) error
	GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (string, error)
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
	IsTokenRevoked(ctx context.Context, refreshToken string) (bool, error)
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
