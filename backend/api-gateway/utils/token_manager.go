package utils

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// ResetToken represents a password reset token
type ResetToken struct {
	Token     string
	Email     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// TokenManager manages password reset tokens
type TokenManager struct {
	tokens map[string]ResetToken
	mutex  sync.RWMutex
}

// NewTokenManager creates a new token manager
func NewTokenManager() *TokenManager {
	return &TokenManager{
		tokens: make(map[string]ResetToken),
	}
}

// Generate creates a new reset token for the given email
func (tm *TokenManager) Generate(email string) ResetToken {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Create a new token
	token := ResetToken{
		Token:     uuid.New().String(),
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(15 * time.Minute), // Token valid for 15 minutes
	}

	// Store the token
	tm.tokens[token.Token] = token

	return token
}

// Validate checks if a token is valid for the given email
func (tm *TokenManager) Validate(tokenString, email string) bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// Get the token
	token, exists := tm.tokens[tokenString]
	if !exists {
		return false
	}

	// Check if the token has expired
	if token.ExpiresAt.Before(time.Now()) {
		return false
	}

	// Check if the token belongs to the user
	return token.Email == email
}

// Delete removes a token
func (tm *TokenManager) Delete(tokenString string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	delete(tm.tokens, tokenString)
}

// Cleanup removes expired tokens
func (tm *TokenManager) Cleanup() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	now := time.Now()
	for key, token := range tm.tokens {
		if token.ExpiresAt.Before(now) {
			delete(tm.tokens, key)
		}
	}
}

// Global token manager instance
var resetTokenManager *TokenManager

// GetTokenManager returns the global token manager instance
func GetTokenManager() *TokenManager {
	if resetTokenManager == nil {
		resetTokenManager = NewTokenManager()

		// Start a goroutine to clean up expired tokens
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				resetTokenManager.Cleanup()
			}
		}()
	}

	return resetTokenManager
}
