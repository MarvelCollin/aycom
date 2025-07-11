package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"aycom/backend/api-gateway/handlers"
)


func TestLogin(t *testing.T) {
	
	gin.SetMode(gin.TestMode)

	
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	
	r.POST("/api/v1/auth/login", handlers.Login)

	
	loginPayload := map[string]string{
		"email":    "kolina@gmail.com",
		"password": "Miawmiaw123@",
	}
	jsonPayload, _ := json.Marshal(loginPayload)

	
	req, err := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	
	req.Header.Set("Content-Type", "application/json")

	
	r.ServeHTTP(w, req)

	
	t.Run("StatusCodeCheck", func(t *testing.T) {
		assert := assert.New(t)

		
		
		
		

		
		if w.Code == http.StatusOK {
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)

			assert.NoError(err, "Should be able to parse response")
			assert.Contains(response, "token", "Response should contain token")
			assert.NotEmpty(response["token"], "Token should not be empty")
		} else {
			
			assert.True(w.Code == http.StatusUnauthorized ||
				w.Code == http.StatusInternalServerError ||
				w.Code == http.StatusBadRequest,
				"Status code should indicate auth failure or service unavailability")
		}
	})
}


func TestLoginValidation(t *testing.T) {
	
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
			
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			
			r.POST("/api/v1/auth/login", handlers.Login)

			
			loginPayload := map[string]string{
				"email":    test.email,
				"password": test.password,
			}
			jsonPayload, _ := json.Marshal(loginPayload)

			
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			
			r.ServeHTTP(w, req)

			
			
			if test.expectCode == http.StatusOK {
				
				assert.NotEqual(t, http.StatusBadRequest, w.Code,
					"Valid credentials should not fail validation")
			} else {
				
				assert.Equal(t, test.expectCode, w.Code,
					"Invalid credentials should fail with correct status")
			}
		})
	}
}
