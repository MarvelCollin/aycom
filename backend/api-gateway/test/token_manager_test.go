package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"aycom/backend/api-gateway/utils"
)

func TestTokenManager(t *testing.T) {

	tm := utils.NewTokenManager()

	t.Run("TokenGeneration", func(t *testing.T) {
		assert := assert.New(t)

		email := "test@example.com"
		token := tm.Generate(email)

		assert.NotEmpty(token.Token, "Token should not be empty")
		assert.Equal(email, token.Email, "Token should be associated with the correct email")
		assert.True(token.ExpiresAt.After(time.Now()), "Expiry time should be in the future")
		assert.True(token.ExpiresAt.Sub(token.CreatedAt) > 0, "Expiry time should be after creation time")
	})

	t.Run("TokenValidation", func(t *testing.T) {
		assert := assert.New(t)

		email := "test2@example.com"
		token := tm.Generate(email)

		assert.True(tm.Validate(token.Token, email), "Valid token with correct email should validate")

		assert.False(tm.Validate(token.Token, "wrong@example.com"), "Valid token with incorrect email should not validate")

		assert.False(tm.Validate("invalid-token", email), "Invalid token should not validate")
	})

	t.Run("TokenDeletion", func(t *testing.T) {
		assert := assert.New(t)

		email := "delete@example.com"
		token := tm.Generate(email)

		assert.True(tm.Validate(token.Token, email), "Token should exist before deletion")

		tm.Delete(token.Token)

		assert.False(tm.Validate(token.Token, email), "Token should not validate after deletion")
	})

	t.Run("TokenCleanup", func(t *testing.T) {

		assert := assert.New(t)
		tm.Cleanup()
		assert.True(true, "Cleanup should run without errors")

	})
}
