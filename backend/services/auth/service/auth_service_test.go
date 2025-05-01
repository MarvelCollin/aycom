package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTokens(t *testing.T) {
	// Create a new AuthServiceImpl with default settings
	service := &AuthServiceImpl{
		jwtSecret: "test_secret_key",
	}

	// Generate tokens with a test user ID
	userID := uuid.New().String()
	ctx := context.Background()
	tokens, err := service.GenerateTokens(ctx, userID)

	// Verify tokens were generated successfully
	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Equal(t, userID, tokens.UserID)
	assert.Equal(t, "Bearer", tokens.TokenType)
	assert.Greater(t, tokens.ExpiresIn, int64(0))
}
