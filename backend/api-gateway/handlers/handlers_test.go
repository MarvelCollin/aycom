package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestHealthCheck tests the HealthCheck handler function
func TestHealthCheck(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Setup the router
	router := gin.New()
	router.GET("/health", HealthCheck)

	// Create a response recorder
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert the response content
	assert.Equal(t, "ok", response["status"])
}

// TestRegister tests the Register handler function
func TestRegister(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Setup the router
	router := gin.New()
	router.POST("/api/v1/auth/register", Register)

	// Create a request body
	requestBody := `{"email":"test@example.com","password":"password123"}`

	// Create a response recorder
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert the response content
	assert.Equal(t, "register endpoint", response["message"])
}

// TestGoogleAuth tests the GoogleAuth handler function
func TestGoogleAuth(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Setup the router
	router := gin.New()
	router.POST("/api/v1/auth/google", GoogleAuth)

	// Create a request body with a simulated Google token ID
	requestBody := `{"token_id":"mock-google-token-id"}`

	// Create a response recorder
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/google", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert the response content
	assert.Equal(t, true, response["success"])
	assert.NotEmpty(t, response["access_token"])
	assert.NotEmpty(t, response["refresh_token"])
}

// Example test for a handler that should validate input
func TestGoogleAuth_InvalidInput(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Setup the router
	router := gin.New()
	router.POST("/api/v1/auth/google", GoogleAuth)

	// Create an invalid request body (missing token_id)
	requestBody := `{}`

	// Create a response recorder
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/google", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response - should be Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert the response content
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["message"], "Invalid request")
}
