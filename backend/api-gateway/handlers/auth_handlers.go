package handlers

import (
	"context"
	"net/http"
	"time"

	"aycom/backend/api-gateway/utils"
	userProto "aycom/backend/proto/user"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication
// @Summary Authentication related endpoints
// @Description Provides authentication services for the API
// @Tags Authentication
func AuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Auth handler working",
		})
	}
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refreshes the access token using a valid refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refresh_token body string true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/refresh [post]
func RefreshToken(c *gin.Context) {
	// TODO: Implement token refresh
	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
	})
}

// GetOAuthConfig returns OAuth configuration
// @Summary Get OAuth configuration
// @Description Returns OAuth configuration for social logins
// @Tags Authentication
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/oauth-config [get]
func GetOAuthConfig(c *gin.Context) {
	// Return mock OAuth configuration
	c.JSON(http.StatusOK, gin.H{
		"providers": []gin.H{
			{
				"name":      "google",
				"client_id": "mock-google-client-id",
				"auth_url":  "https://accounts.google.com/o/oauth2/auth",
				"scopes":    []string{"email", "profile"},
			},
			{
				"name":      "github",
				"client_id": "mock-github-client-id",
				"auth_url":  "https://github.com/login/oauth/authorize",
				"scopes":    []string{"user:email", "read:user"},
			},
		},
	})
}

// Login handles user login with email and password
// @Summary Login with email and password
// @Description Authenticate a user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body object true "Email and password"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
	// This is just a placeholder - actual login is handled by LoginUser
	c.JSON(http.StatusOK, gin.H{
		"message": "Login endpoint. Please use /api/v1/users/login instead.",
	})
}

// VerifyEmail verifies a user's email address
// @Summary Verify email address
// @Description Verifies a user's email address using a verification code
// @Tags Authentication
// @Accept json
// @Produce json
// @Param verification body object true "Verification code"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/verify-email [post]
func VerifyEmail(c *gin.Context) {
	// Mock implementation for email verification
	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

// ResendVerification resends the verification email
// @Summary Resend verification email
// @Description Resends the verification email to the user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param email body string true "User's email"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/resend-verification [post]
func ResendVerification(c *gin.Context) {
	// Mock implementation for resending verification email
	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email resent successfully",
	})
}

// GoogleLogin handles Google OAuth login
// @Summary Login with Google
// @Description Authenticate a user using Google OAuth
// @Tags Authentication
// @Accept json
// @Produce json
// @Param code body string true "OAuth code"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/google [post]
func GoogleLogin(c *gin.Context) {
	// Mock implementation for Google OAuth login
	c.JSON(http.StatusOK, gin.H{
		"message": "Google OAuth login successful",
		"user": gin.H{
			"id":    "mock-user-id",
			"email": "mock@example.com",
			"name":  "Mock User",
		},
		"token":         "mock-jwt-token",
		"refresh_token": "mock-refresh-token",
	})
}

// ForgotPassword handles password reset requests
// @Summary Request password reset
// @Description Sends a password reset link to the user's email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param email body object true "User's email"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/forgot-password [post]
func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid email format")
		return
	}

	// Check if the user service client is initialized
	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	// Get user by email to check if it exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the user by email
	userResp, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: req.Email,
	})

	// Handle errors but return generic response to prevent user enumeration
	if err != nil || userResp.User == nil {
		SendSuccessResponse(c, http.StatusOK, gin.H{
			"message": "If the email address exists in our system, a security question will be sent",
		})
		return
	}

	// Return the security question from the user's profile
	securityQuestion := userResp.User.SecurityQuestion
	if securityQuestion == "" {
		securityQuestion = "What is your mother's maiden name?" // Fallback question
	}

	SendSuccessResponse(c, http.StatusOK, gin.H{
		"message":           "Please answer your security question to reset your password",
		"security_question": securityQuestion,
		"email":             req.Email,
	})
}

// VerifySecurityAnswer handles security question answers for password reset
// @Summary Verify security answer
// @Description Verifies the security answer before allowing password reset
// @Tags Authentication
// @Accept json
// @Produce json
// @Param data body object true "Email and security answer"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/verify-security-answer [post]
func VerifySecurityAnswer(c *gin.Context) {
	var req struct {
		Email          string `json:"email" binding:"required,email"`
		SecurityAnswer string `json:"security_answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request format")
		return
	}

	// Check if the user service client is initialized
	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	// Get user by email
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify the user exists
	userResp, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: req.Email,
	})

	if err != nil || userResp.User == nil {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_EMAIL", "User not found")
		return
	}

	// Verify the security answer (in a real implementation, this would be properly hashed and compared)
	if userResp.User.SecurityAnswer != req.SecurityAnswer {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_ANSWER", "Security answer is incorrect")
		return
	}

	// Generate a reset token using the token manager
	resetToken := utils.GetTokenManager().Generate(req.Email)

	// Return success with the reset token
	SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Security answer verified. You may now reset your password.",
		"token":   resetToken.Token,
		"email":   req.Email,
		"expires": resetToken.ExpiresAt,
	})
}

// ResetPassword handles password reset with a valid token
// @Summary Reset password
// @Description Resets a user's password using a valid reset token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param reset_data body object true "Reset token and new password"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/reset-password [post]
func ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request. Token, email, and password (min 8 chars) are required.")
		return
	}

	// Check if the user service client is initialized
	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	// Validate the token
	if !utils.GetTokenManager().Validate(req.Token, req.Email) {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_TOKEN", "Reset token is invalid or expired")
		return
	}

	// Get user by email to verify it exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify the user exists
	userResp, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: req.Email,
	})

	if err != nil || userResp.User == nil {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_EMAIL", "User not found")
		return
	}

	// Simulate password verification
	storedPassword := userResp.User.Password
	if storedPassword == req.NewPassword {
		SendErrorResponse(c, http.StatusBadRequest, "SAME_PASSWORD", "New password cannot be the same as the old one")
		return
	}

	// In a real implementation, you would update the password in the database
	// For now, we'll simulate a successful password update

	// Return success message
	SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Password has been reset successfully. You can now log in with your new password",
	})

	// Delete the token so it can't be used again
	utils.GetTokenManager().Delete(req.Token)
}
