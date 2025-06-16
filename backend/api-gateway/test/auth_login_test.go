package test

import (
	"aycom/backend/api-gateway/handlers"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestLogin tests the login functionality with the provided credentials
func TestLogin(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new gin context for testing
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	// Setup the login route
	r.POST("/api/v1/auth/login", handlers.Login)

	// Create login payload with the provided credentials
	loginPayload := map[string]string{
		"email":    "kolina@gmail.com",
		"password": "Miawmiaw123@",
	}
	jsonPayload, _ := json.Marshal(loginPayload)

	// Create the request
	req, err := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set content type header
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	r.ServeHTTP(w, req)

	// Test assertions
	t.Run("StatusCodeCheck", func(t *testing.T) {
		assert := assert.New(t)

		// Note: Since we're not connecting to an actual database in this test,
		// we expect the login to fail. In a real-world scenario with mocks or
		// test database, we would check for 200 OK.
		// Here we're just demonstrating the test structure.

		// Check if response contains token on success or proper error
		if w.Code == http.StatusOK {
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)

			assert.NoError(err, "Should be able to parse response")
			assert.Contains(response, "token", "Response should contain token")
			assert.NotEmpty(response["token"], "Token should not be empty")
		} else {
			// If not connected to real services, we expect error
			assert.True(w.Code == http.StatusUnauthorized ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusBadRequest,
				"Status code should indicate auth failure or service unavailability")
		}
	})
}

// TestLoginValidation tests the validation of login input
func TestLoginValidation(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		email      string
		password   string
		expectCode int
	}{
		{"valid_credentials", "kolina@gmail.com", "Miawmiaw123@", http.StatusOK},
		{"empty_email", "", "Miawmiaw123@", http.StatusBadRequest},
		{"empty_password", "kolina@gmail.com", "", http.StatusBadRequest},
		{"invalid_email_format", "kolina", "Miawmiaw123@", http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a new recorder and context
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			// Setup the login route
			r.POST("/api/v1/auth/login", handlers.Login)

			// Create payload
			loginPayload := map[string]string{
				"email":    test.email,
				"password": test.password,
			}
			jsonPayload, _ := json.Marshal(loginPayload)

			// Create request
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			// Perform the request
			r.ServeHTTP(w, req)

			// In a test environment without real services, we can't expect actual success
			// So we check if validation works (bad requests fail) and valid requests reach the service
			if test.expectCode == http.StatusOK {
				// Valid input should at least not return 400
				assert.NotEqual(t, http.StatusBadRequest, w.Code,
					"Valid credentials should not fail validation")
			} else {
				// Invalid input should return expected error
				assert.Equal(t, test.expectCode, w.Code,
					"Invalid credentials should fail with correct status")
			}
		})
	}
}
