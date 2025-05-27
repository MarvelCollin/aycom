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

func GenerateJWT(userID string, expiryDuration time.Duration) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		secret = []byte("wompwomp123")
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

func GenerateVerificationCode() string {

	code := ""
	for i := 0; i < 6; i++ {

		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += fmt.Sprintf("%d", n)
	}
	return code
}

func GenerateUsername(name string) string {

	username := strings.ToLower(name)
	username = strings.ReplaceAll(username, " ", "")

	var result strings.Builder
	for _, r := range username {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		}
	}
	username = result.String()

	if len(username) > 10 {
		username = username[:10]
	}

	randomStr := generateRandomString(5)
	username = username + randomStr

	return username
}

func GenerateSecureRandomPassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		password[i] = chars[n.Int64()]
	}
	return string(password)
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return strings.ToLower(base64.URLEncoding.EncodeToString(b)[:length])
}
