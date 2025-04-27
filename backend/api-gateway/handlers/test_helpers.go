package handlers

import (
	"github.com/gin-gonic/gin"
)

// This file contains helper functions and structures for testing

// MockServiceClients holds mock implementations of service clients
// that can be used in tests to avoid real gRPC calls
type MockServiceClients struct {
	AuthClient MockAuthClient
	UserClient MockUserClient
}

// MockAuthClient represents a mock implementation of the auth service client
type MockAuthClient struct {
	// Functions that can be overridden in tests
	RegisterFunc     func(email, password string) (bool, error)
	LoginFunc        func(email, password string) (string, string, error)
	RefreshTokenFunc func(refreshToken string) (string, string, error)
	GoogleAuthFunc   func(tokenID string) (string, string, string, error)
	VerifyEmailFunc  func(email, code string) (bool, error)
	ResendVerifyFunc func(email string) (bool, error)
}

// MockUserClient represents a mock implementation of the user service client
type MockUserClient struct {
	// Functions that can be overridden in tests
	GetUserProfileFunc    func(userID string) (map[string]interface{}, error)
	UpdateUserProfileFunc func(userID string, data map[string]interface{}) (bool, error)
}

// SetupTestRouter creates a Gin router for testing with provided mock clients
func SetupTestRouter(mocks *MockServiceClients) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Setup routes that use your mock clients
	// When implementing the gRPC clients, you'll inject these mocks

	return r
}
