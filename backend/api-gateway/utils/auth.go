package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID string, expiryDuration time.Duration) (string, error) {
	secret := GetJWTSecret()

	expiryTime := time.Now().Add(expiryDuration)

	claims := jwt.MapClaims{
		"sub":     userID,
		"user_id": userID,
		"exp":     expiryTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	log.Printf("Generating JWT for user %s with claims: sub=%s, user_id=%s, exp=%d",
		userID, userID, userID, expiryTime.Unix())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)

	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		return "", err
	}

	log.Printf("JWT generated successfully, length: %d chars", len(tokenString))
	return tokenString, nil
}

func GetJWTSecret() []byte {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		log.Println("Warning: JWT_SECRET environment variable not set or empty, using fallback value. This is not secure for production use.")
		jwtSecret = []byte("insecure_fallback_jwt_key")
	}
	return jwtSecret
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
