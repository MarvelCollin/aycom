package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateJWT generates a new JWT token with the given user ID and expiration duration
func GenerateJWT(userID string, expiryDuration time.Duration) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		secret = []byte("wompwomp123") // Fallback secret from env_docs.md
	}

	expiryTime := time.Now().Add(expiryDuration)

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": expiryTime.Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// GenerateVerificationCode generates a 6-digit verification code
func GenerateVerificationCode() string {
	// Generate 6 random digits
	code := ""
	for i := 0; i < 6; i++ {
		// Generate a random number between 0-9
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += fmt.Sprintf("%d", n)
	}
	return code
}

// GenerateUsername creates a unique username from a name by adding random characters
func GenerateUsername(name string) string {
	// Remove spaces and special characters
	username := strings.ToLower(name)
	username = strings.ReplaceAll(username, " ", "")

	// Keep only alphanumeric characters
	var result strings.Builder
	for _, r := range username {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		}
	}
	username = result.String()

	// Ensure username is not too long (max 15 chars)
	if len(username) > 10 {
		username = username[:10]
	}

	// Add 5 random characters to make it unique
	randomStr := generateRandomString(5)
	username = username + randomStr

	return username
}

// GenerateSecureRandomPassword generates a secure random password of the given length
func GenerateSecureRandomPassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		password[i] = chars[n.Int64()]
	}
	return string(password)
}

// Helper function to generate random string
func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return strings.ToLower(base64.URLEncoding.EncodeToString(b)[:length])
}
