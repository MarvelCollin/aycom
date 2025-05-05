package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	userProto "aycom/backend/proto/user"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

// GetUserProfile retrieves the user's profile from the User service via gRPC
// @Summary Get user profile
// @Description Returns the profile of the authenticated user
// @Tags Users
// @Produce json
// @Router /api/v1/users/profile [get]
func GetUserProfile(c *gin.Context) {
	// Get user ID from JWT token (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userIDStr := userID.(string)
	log.Printf("GetUserProfile Handler: Retrieved userID from context: %s", userIDStr)

	// Validate UUID
	if _, err := uuid.Parse(userIDStr); err != nil {
		log.Printf("GetUserProfile Handler: Invalid UUID format for userID: %s", userIDStr)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	// Use our service client
	if userServiceClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	// Call the service client method
	user, err := userServiceClient.GetUserProfile(userIDStr)
	if err != nil {
		// Handle errors
		st, ok := status.FromError(err)
		if ok {
			// Map gRPC status code to HTTP status code
			switch st.Code() {
			case 5: // Not found
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User profile not found")
			case 16: // Unauthenticated
				SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized to access this profile")
			default:
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve user profile: "+st.Message())
			}
		} else {
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while retrieving profile")
		}
		log.Printf("Error retrieving user profile: %v", err)
		return
	}

	// Return successful response with user profile
	SendSuccessResponse(c, http.StatusOK, gin.H{
		"user": gin.H{
			"id":                  user.ID,
			"name":                user.Name,
			"username":            user.Username,
			"email":               user.Email,
			"profile_picture_url": user.ProfilePictureURL,
			"bio":                 user.Bio,
			"is_verified":         user.IsVerified,
			"follower_count":      user.FollowerCount,
			"following_count":     user.FollowingCount,
		},
	})
}

// UpdateUserProfile updates the user's profile in the User service via gRPC
// @Summary Update user profile
// @Description Updates the profile of the authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Router /api/v1/users/profile [put]
func UpdateUserProfile(c *gin.Context) {
	// Get user ID from JWT token (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userIDStr := userID.(string)
	log.Printf("UpdateUserProfile Handler: Retrieved userID from context: %s", userIDStr)

	// Validate UUID
	if _, err := uuid.Parse(userIDStr); err != nil {
		log.Printf("UpdateUserProfile Handler: Invalid UUID format for userID: %s", userIDStr)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	// Check if the service client is initialized
	if userServiceClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	// Bind input from request body
	var input struct {
		Name              string `json:"name"`
		Bio               string `json:"bio"`
		ProfilePictureURL string `json:"profile_picture_url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload: "+err.Error())
		return
	}

	// Create profile update object
	profileUpdate := &UserProfileUpdate{
		Name:              input.Name,
		Bio:               input.Bio,
		ProfilePictureURL: input.ProfilePictureURL,
	}

	// Call the service client method
	updatedUser, err := userServiceClient.UpdateUserProfile(userIDStr, profileUpdate)
	if err != nil {
		// Handle errors
		st, ok := status.FromError(err)
		if ok {
			// Map gRPC status code to HTTP status code
			switch st.Code() {
			case 5: // Not found
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User profile not found")
			case 16: // Unauthenticated
				SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized to update this profile")
			default:
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user profile: "+st.Message())
			}
		} else {
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while updating profile")
		}
		log.Printf("Error updating user profile: %v", err)
		return
	}

	// Return successful response with updated user profile
	SendSuccessResponse(c, http.StatusOK, gin.H{
		"user": gin.H{
			"id":                  updatedUser.ID,
			"name":                updatedUser.Name,
			"username":            updatedUser.Username,
			"email":               updatedUser.Email,
			"profile_picture_url": updatedUser.ProfilePictureURL,
			"bio":                 updatedUser.Bio,
			"is_verified":         updatedUser.IsVerified,
			"follower_count":      updatedUser.FollowerCount,
			"following_count":     updatedUser.FollowingCount,
		},
	})
}

// RegisterUser handles user registration
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User registration details"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse "Email already in use"
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/auth/register [post]
func RegisterUser(c *gin.Context) {
	var req struct {
		Name                  string `json:"name"`
		Username              string `json:"username"`
		Email                 string `json:"email"`
		Password              string `json:"password"`
		ConfirmPassword       string `json:"confirm_password"`
		Gender                string `json:"gender"`
		DateOfBirth           string `json:"date_of_birth"`
		SecurityQuestion      string `json:"securityQuestion"`
		SecurityAnswer        string `json:"securityAnswer"`
		SubscribeToNewsletter bool   `json:"subscribeToNewsletter"`
		RecaptchaToken        string `json:"recaptcha_token"`
		ProfilePictureUrl     string `json:"profile_picture_url,omitempty"`
		BannerUrl             string `json:"banner_url,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Success: false, Message: "Invalid request: " + err.Error()})
		return
	}

	// Use the globally initialized UserClient from handlers/common.go
	if UserClient == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Success: false, Message: "User service unavailable"})
		return
	}

	user := &userProto.User{
		Id:                "", // Let backend generate UUID
		Name:              req.Name,
		Username:          req.Username,
		Email:             req.Email,
		Gender:            req.Gender,
		DateOfBirth:       req.DateOfBirth,
		ProfilePictureUrl: req.ProfilePictureUrl,
		BannerUrl:         req.BannerUrl,
		// Map new fields
		Password:              req.Password, // Send raw password
		SecurityQuestion:      req.SecurityQuestion,
		SecurityAnswer:        req.SecurityAnswer,
		SubscribeToNewsletter: req.SubscribeToNewsletter,
	}
	createReq := &userProto.CreateUserRequest{User: user}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	resp, err := UserClient.CreateUser(ctx, createReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == 6 {
			c.JSON(http.StatusConflict, ErrorResponse{Success: false, Message: "User already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Failed to register user: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Registration successful", "user": resp.User})
}

// LoginUser handles user login
// @Summary Login a user
// @Description Authenticate a user and return tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "User credentials"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse "Invalid credentials"
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/auth/login [post]
func LoginUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Success: false, Message: "Invalid request: " + err.Error()})
		return
	}

	if UserClient == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Success: false, Message: "User service unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	loginReq := &userProto.LoginUserRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	loginResp, err := UserClient.LoginUser(ctx, loginReq)
	if err != nil || loginResp.User == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Success: false, Message: "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := generateJWT(loginResp.User.Id)
	if err != nil {
		log.Printf("Error generating JWT token: %v", err)
		SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error generating authentication token")
		return
	}

	// Construct and send the response
	response := AuthServiceResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  token,
		RefreshToken: "",
		UserId:       loginResp.User.Id,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour in seconds
	}

	c.JSON(http.StatusOK, response)
}

// GetUserByEmail retrieves a user by email from the User service via gRPC
func GetUserByEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Email is required",
		})
		return
	}

	if UserClient == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Success: false,
			Message: "User service unavailable",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	grpcReq := &userProto.GetUserByEmailRequest{Email: req.Email}
	resp, err := UserClient.GetUserByEmail(ctx, grpcReq)
	if err != nil || resp.User == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Success: false, Message: "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "user": resp.User})
}

// generateJWT generates a JSON Web Token for the given userID
func generateJWT(userID string) (string, error) {
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		"iat":     time.Now().Unix(),
	})

	// Sign the token with our secret
	tokenString, err := token.SignedString(GetJWTSecret())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
