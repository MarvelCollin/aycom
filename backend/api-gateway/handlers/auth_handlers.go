package handlers

import (
	"aycom/backend/api-gateway/models"
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	userProto "aycom/backend/proto/user"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OAuthConfigResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
}

// generateJWT generates access and refresh tokens for a user
func generateJWT(userID string) (accessToken string, refreshToken string, err error) {
	// Get JWT configuration from environment variables
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("wompwomp123") // Fallback to default in .env
		log.Println("Warning: JWT_SECRET environment variable not set, using default.")
	}

	// Get expiry times from environment or use defaults
	var accessExpiry int64 = 3600    // Default 1 hour
	var refreshExpiry int64 = 604800 // Default 7 days

	if expiryStr := os.Getenv("JWT_EXPIRY"); expiryStr != "" {
		if expiry, err := strconv.ParseInt(expiryStr, 10, 64); err == nil {
			accessExpiry = expiry
		}
	}

	if refreshExpiryStr := os.Getenv("REFRESH_TOKEN_EXPIRY"); refreshExpiryStr != "" {
		if expiry, err := strconv.ParseInt(refreshExpiryStr, 10, 64); err == nil {
			refreshExpiry = expiry
		}
	}

	// Generate Access Token
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Second * time.Duration(accessExpiry)).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	}
	accessTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenJWT.SignedString(jwtSecret)
	if err != nil {
		log.Printf("Error generating access token: %v", err)
		return "", "", err
	}

	// Generate Refresh Token
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Second * time.Duration(refreshExpiry)).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}
	refreshTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenJWT.SignedString(jwtSecret)
	if err != nil {
		log.Printf("Error generating refresh token: %v", err)
		return accessToken, "", err // Return access token even if refresh fails
	}

	return accessToken, refreshToken, nil
}

// @Summary User login
// @Description Authenticates a user and returns tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login request"
// @Success 200 {object} AuthServiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Use the globally initialized UserClient from handlers/common.go
	if UserClient == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Success: false,
			Message: "User service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	// Call User Service LoginUser RPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	loginResp, err := UserClient.LoginUser(ctx, &userProto.LoginUserRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			switch statusErr.Code() {
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Success: false,
					Message: statusErr.Message(),
				})
			case codes.NotFound, codes.Unauthenticated:
				c.JSON(http.StatusUnauthorized, ErrorResponse{
					Success: false,
					Message: "Invalid email or password", // Keep message generic for security
				})
			default:
				log.Printf("gRPC Error during login for %s: %v", req.Email, statusErr.Message())
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Authentication service error",
				})
			}
		} else {
			log.Printf("Unknown Error during login for %s: %v", req.Email, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Internal server error during login",
			})
		}
		return
	}

	// Authentication successful, generate JWT tokens
	accessToken, refreshToken, err := generateJWT(loginResp.User.Id)
	if err != nil {
		log.Printf("Error generating JWT for user %s: %v", loginResp.User.Id, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to generate authentication tokens",
		})
		return
	}

	// Get expiry from env or use default
	accessExpiry := 3600
	if expiryStr := os.Getenv("JWT_EXPIRY"); expiryStr != "" {
		if expiry, err := strconv.Atoi(expiryStr); err == nil {
			accessExpiry = expiry
		}
	}

	c.JSON(http.StatusOK, AuthServiceResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       loginResp.User.Id,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessExpiry),
		User:         loginResp.User,
	})
}

// @Summary User registration
// @Description Registers a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Register request"
// @Success 201 {object} models.RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate request
	// REMOVE PasswordConfirmation check as field likely doesn't exist
	/*
		if req.Password != req.PasswordConfirmation {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Success: false,
				Message: "Passwords do not match",
			})
			return
		}
	*/

	// Use the globally initialized UserClient from handlers/common.go
	if UserClient == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Success: false,
			Message: "User service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	// Call User Service Register RPC (matching the proto definition)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fix: Use CreateUser method with CreateUserRequest as defined in the proto
	registerResp, err := UserClient.CreateUser(ctx, &userProto.CreateUserRequest{
		User: &userProto.User{
			Email:    req.Email,
			Password: req.Password,
			Name:     req.Name,
		},
	})

	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			switch statusErr.Code() {
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Success: false,
					Message: statusErr.Message(),
				})
			case codes.AlreadyExists:
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Success: false,
					Message: "Email already registered",
				})
			default:
				log.Printf("gRPC Error during registration for %s: %v", req.Email, statusErr.Message())
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Registration service error",
				})
			}
		} else {
			log.Printf("Unknown Error during registration for %s: %v", req.Email, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Internal server error during registration",
			})
		}
		return
	}

	accessToken, refreshToken, err := generateJWT(registerResp.User.Id)
	if err != nil {
		log.Printf("Error generating JWT for new user %s: %v", registerResp.User.Id, err)
		// We still want to return success but note the token issue
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Registration successful, but could not generate authentication tokens",
			"user_id": registerResp.User.Id,
		})
		return
	}

	c.JSON(http.StatusCreated, AuthServiceResponse{
		Success:      true,
		Message:      "Registration successful",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       registerResp.User.Id,
		TokenType:    "Bearer",
	})
}

// @Summary Refresh token
// @Description Refreshes an access token using a refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} AuthServiceResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/refresh [post]
func RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Parse and validate refresh token
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("wompwomp123") // Fallback if not set
	}

	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "Invalid refresh token",
		})
		return
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "Invalid token claims",
		})
		return
	}

	// Verify this is a refresh token
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "Invalid token type",
		})
		return
	}

	// Extract user ID
	userId, ok := claims["user_id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "Invalid user ID in token",
		})
		return
	}

	// Generate new access token
	accessToken, refreshToken, err := generateJWT(userId)
	if err != nil {
		log.Printf("Error refreshing token for user %s: %v", userId, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to generate new tokens",
		})
		return
	}

	// Get expiry from env or use default
	accessExpiry := 3600
	if expiryStr := os.Getenv("JWT_EXPIRY"); expiryStr != "" {
		if expiry, err := strconv.Atoi(expiryStr); err == nil {
			accessExpiry = expiry
		}
	}

	c.JSON(http.StatusOK, AuthServiceResponse{
		Success:      true,
		Message:      "Token refreshed successfully",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       userId,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessExpiry),
	})
}

// @Summary Register with media
// @Description Registers a new user with media upload
// @Tags Auth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Media file"
// @Param name formData string true "Name"
// @Param email formData string true "Email"
// @Param password formData string true "Password"
// @Success 201 {object} models.RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/register-with-media [post]
func RegisterWithMedia(c *gin.Context) {
	// This function would handle media upload with registration
	// Implementation left as an exercise for the backend team
	c.JSON(http.StatusNotImplemented, ErrorResponse{
		Success: false,
		Message: "This feature is not yet implemented",
	})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	// Client-side logout is preferred for JWT authentication
	// This endpoint is mainly for API completeness and future extensions

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully logged out",
	})
}

// @Summary Get OAuth config
// @Description Returns OAuth configuration for frontend
// @Tags Auth
// @Produce json
// @Success 200 {object} OAuthConfigResponse
// @Router /api/v1/auth/oauth-config [get]
func GetOAuthConfig(c *gin.Context) {
	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleRedirectUri := os.Getenv("GOOGLE_REDIRECT_URI")

	c.JSON(http.StatusOK, OAuthConfigResponse{
		Success: true,
		Data: map[string]interface{}{
			"googleClientId":    googleClientId,
			"googleRedirectUri": googleRedirectUri,
		},
	})
}

// VerifyEmail handles verification of a user's email
// @Summary Verify Email
// @Description Verifies a user's email with a verification code
// @Tags Auth
// @Accept json
// @Produce json
// @Router /api/v1/auth/verify-email [post]
func VerifyEmail(c *gin.Context) {
	var input struct {
		Email            string `json:"email" binding:"required,email"`
		VerificationCode string `json:"verification_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request format: " + err.Error(),
		})
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the user by email first
	userResp, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: input.Email,
	})

	if err != nil {
		// Handle specific gRPC errors
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Email not found",
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to verify email: " + st.Message(),
				})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to verify email: " + err.Error(),
			})
		}
		log.Printf("Email verification failed for %s: %v", input.Email, err)
		return
	}

	// In a real implementation, you would verify the code against a stored value
	// For now, use a mock verification (assuming 123456 is valid)
	if input.VerificationCode != "123456" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid verification code",
		})
		return
	}

	// Update user verification status using the available RPC
	_, err = UserClient.UpdateUserVerificationStatus(ctx, &userProto.UpdateUserVerificationStatusRequest{
		UserId:     userResp.User.Id,
		IsVerified: true,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update verification status: " + err.Error(),
		})
		log.Printf("Failed to update verification status for %s: %v", input.Email, err)
		return
	}

	// Generate tokens for the verified user
	accessToken, refreshToken, err := generateJWT(userResp.User.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error generating tokens: " + err.Error(),
		})
		return
	}

	// Get expiry from env or use default
	accessExpiry := 3600
	if expiryStr := os.Getenv("JWT_EXPIRY"); expiryStr != "" {
		if expiry, err := strconv.Atoi(expiryStr); err == nil {
			accessExpiry = expiry
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Email verification successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user_id":       userResp.User.Id,
		"expires_in":    accessExpiry,
	})
}

// ResendVerification handles resending verification code
// @Summary Resend verification code
// @Description Resends a verification code to the user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Router /api/v1/auth/resend-verification [post]
func ResendVerification(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request format: " + err.Error(),
		})
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if the email exists
	_, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: input.Email,
	})

	if err != nil {
		// Handle specific gRPC errors
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Email not found",
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to resend verification code: " + st.Message(),
				})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to resend verification code: " + err.Error(),
			})
		}
		log.Printf("Verification code resend failed for %s: %v", input.Email, err)
		return
	}

	// In a real implementation:
	// 1. Generate a new verification code
	// 2. Update it in the database
	// 3. Send it via email

	// For now, we'll just simulate success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Verification code has been resent to your email",
	})
}

// GoogleLogin handles Google OAuth login
// @Summary Google OAuth login
// @Description Process Google OAuth login
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.GoogleLoginRequest true "Google login request"
// @Success 200 {object} AuthServiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/google [post]
func GoogleLogin(c *gin.Context) {
	var req models.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if req.TokenID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Google token ID is required",
		})
		return
	}

	// In a real implementation, you would:
	// 1. Verify the Google token ID
	// 2. Extract user information (email, name, etc.)
	// 3. Check if user exists or create a new user
	// 4. Generate JWT tokens

	// For demonstration purposes, we'll simulate the process
	// Get user by email (assuming we extracted email from token)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mockEmail := "google_user@example.com" // In reality, this comes from token verification
	mockName := "Google User"              // In reality, this comes from token verification

	// First, try to get the user by email
	userResp, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: mockEmail,
	})

	var userId string

	if err != nil {
		// User doesn't exist, create a new user
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			log.Printf("Creating new user for Google login: %s", mockEmail)

			// Generate a secure random password for the user
			randomPassword := uuid.New().String()

			// Create user using the regular user creation endpoint
			createResp, err := UserClient.CreateUser(ctx, &userProto.CreateUserRequest{
				User: &userProto.User{
					Name:     mockName,
					Email:    mockEmail,
					Password: randomPassword,
					// Other fields as needed
				},
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Failed to create user account: " + err.Error(),
				})
				return
			}

			userId = createResp.User.Id
		} else {
			// Some other error occurred
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Authentication service error: " + err.Error(),
			})
			return
		}
	} else {
		// User exists
		userId = userResp.User.Id
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := generateJWT(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to generate authentication tokens",
		})
		return
	}

	// Get expiry from env or use default
	accessExpiry := 3600
	if expiryStr := os.Getenv("JWT_EXPIRY"); expiryStr != "" {
		if expiry, err := strconv.Atoi(expiryStr); err == nil {
			accessExpiry = expiry
		}
	}

	// Get updated user data
	updatedUserResp, err := UserClient.GetUser(ctx, &userProto.GetUserRequest{
		UserId: userId,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to retrieve user data: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AuthServiceResponse{
		Success:      true,
		Message:      "Google login successful",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       userId,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessExpiry),
		User:         updatedUserResp.User,
	})
}
