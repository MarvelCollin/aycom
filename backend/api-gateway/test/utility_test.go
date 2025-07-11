package test

import (
	"testing"
	"aycom/backend/api-gateway/utils"
	"github.com/stretchr/testify/assert"
)


func TestUtilityFunctions(t *testing.T) {
	t.Run("StringSimilarityBasic", func(t *testing.T) {
		
		distance := utils.DamerauLevenshteinDistance("hello", "hallo")
		assert.Equal(t, 1, distance, "Distance should be 1 for hello->hallo")
		
		
		distance2 := utils.DamerauLevenshteinDistance("test", "test")
		assert.Equal(t, 0, distance2, "Distance should be 0 for identical strings")
	})
	
	t.Run("TokenManagerBasic", func(t *testing.T) {
		tm := utils.NewTokenManager()
		
		
		email := "test@example.com"
		token := tm.Generate(email)
		assert.NotEmpty(t, token.Token, "Token should not be empty")
		assert.Equal(t, email, token.Email, "Email should match")
		
		
		assert.True(t, tm.Validate(token.Token, email), "Token should validate with correct email")
		assert.False(t, tm.Validate(token.Token, "wrong@example.com"), "Token should not validate with wrong email")
	})
}
