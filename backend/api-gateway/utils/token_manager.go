package utils

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type ResetToken struct {
	Token     string
	Email     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type TokenManager struct {
	tokens map[string]ResetToken
	mutex  sync.RWMutex
}

func NewTokenManager() *TokenManager {
	return &TokenManager{
		tokens: make(map[string]ResetToken),
	}
}

func (tm *TokenManager) Generate(email string) ResetToken {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	token := ResetToken{
		Token:     uuid.New().String(),
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	tm.tokens[token.Token] = token

	return token
}

func (tm *TokenManager) Validate(tokenString, email string) bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	token, exists := tm.tokens[tokenString]
	if !exists {
		return false
	}

	if token.ExpiresAt.Before(time.Now()) {
		return false
	}

	return token.Email == email
}

func (tm *TokenManager) Delete(tokenString string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	delete(tm.tokens, tokenString)
}

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

var resetTokenManager *TokenManager

func GetTokenManager() *TokenManager {
	if resetTokenManager == nil {
		resetTokenManager = NewTokenManager()

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
