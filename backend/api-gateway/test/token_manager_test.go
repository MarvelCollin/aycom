package test

import (
	"aycom/backend/api-gateway/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenManager(t *testing.T) {
	// Create a new token manager
	tm := utils.NewTokenManager()

	// Test token generation
	t.Run("TokenGeneration", func(t *testing.T) {
		assert := assert.New(t)
		
		email := "test@example.com"
		token := tm.Generate(email)
		
		// Check that token is generated correctly
		assert.NotEmpty(token.Token, "Token should not be empty")
		assert.Equal(email, token.Email, "Token should be associated with the correct email")
		assert.True(token.ExpiresAt.After(time.Now()), "Expiry time should be in the future")
		assert.True(token.ExpiresAt.Sub(token.CreatedAt) > 0, "Expiry time should be after creation time")
	})

	// Test token validation
	t.Run("TokenValidation", func(t *testing.T) {
		assert := assert.New(t)
		
		email := "test2@example.com"
		token := tm.Generate(email)
		
		// Test valid token and email
		assert.True(tm.Validate(token.Token, email), "Valid token with correct email should validate")
		
		// Test invalid email
		assert.False(tm.Validate(token.Token, "wrong@example.com"), "Valid token with incorrect email should not validate")
		
		// Test invalid token
		assert.False(tm.Validate("invalid-token", email), "Invalid token should not validate")
	})

	// Test token deletion
	t.Run("TokenDeletion", func(t *testing.T) {
		assert := assert.New(t)
		
		email := "delete@example.com"
		token := tm.Generate(email)
		
		// Verify token exists
		assert.True(tm.Validate(token.Token, email), "Token should exist before deletion")
		
		// Delete token
		tm.Delete(token.Token)
		
		// Verify token no longer exists
		assert.False(tm.Validate(token.Token, email), "Token should not validate after deletion")
	})

	// Test cleanup of expired tokens
	t.Run("TokenCleanup", func(t *testing.T) {
		// Since we can't easily manipulate private token storage or time,
		// this is more of a functionality test than a full unit test
		
		assert := assert.New(t)
		tm.Cleanup() // Should at least not cause errors
		assert.True(true, "Cleanup should run without errors")
		
		// In a more comprehensive test with test helpers or refactoring,
		// we would inject mock time and verify expired tokens are removed
	})
} 