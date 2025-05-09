package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"aycom/backend/api-gateway/utils"
	userProto "aycom/backend/proto/user"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userIDStr := userID.(string)
	log.Printf("GetUserProfile Handler: Retrieved userID from context: %s", userIDStr)

	if _, err := uuid.Parse(userIDStr); err != nil {
		log.Printf("GetUserProfile Handler: Invalid UUID format for userID: %s", userIDStr)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	if userServiceClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	user, err := userServiceClient.GetUserProfile(userIDStr)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
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

func UpdateUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userIDStr := userID.(string)
	log.Printf("UpdateUserProfile Handler: Retrieved userID from context: %s", userIDStr)

	if _, err := uuid.Parse(userIDStr); err != nil {
		log.Printf("UpdateUserProfile Handler: Invalid UUID format for userID: %s", userIDStr)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	if userServiceClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	var input struct {
		Name              string `json:"name"`
		Bio               string `json:"bio"`
		ProfilePictureURL string `json:"profile_picture_url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload: "+err.Error())
		return
	}

	profileUpdate := &UserProfileUpdate{
		Name:              input.Name,
		Bio:               input.Bio,
		ProfilePictureURL: input.ProfilePictureURL,
	}

	updatedUser, err := userServiceClient.UpdateUserProfile(userIDStr, profileUpdate)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
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

// GetUserSuggestions returns suggested users for the current user to follow
// @Summary Get user suggestions
// @Description Get a list of suggested users to follow
// @Tags Users
// @Accept json
// @Produce json
// @Param limit query int false "Number of suggestions to fetch"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/suggestions [get]
func GetUserSuggestions(c *gin.Context) {
	userID := ""
	userIDAny, exists := c.Get("userId")
	if exists {
		userID, _ = userIDAny.(string)
	}

	limitStr := c.DefaultQuery("limit", "3")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 3 // Default limit
	}

	log.Printf("Fetching user suggestions for user %s, limit: %d", userID, limit)

	if UserClient == nil {
		log.Printf("User service unavailable, returning mock suggestions")
		returnMockSuggestions(c, limit)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := UserClient.SearchUsers(ctx, &userProto.SearchUsersRequest{
		Query:  "",        // Empty query to find all users
		Filter: "popular", // This filter tells the service to sort by follower count
		Page:   1,
		Limit:  int32(limit + 1), // Request one more to account for filtering out current user
	})

	if err != nil {
		log.Printf("Failed to get user suggestions: %v", err)
		returnMockSuggestions(c, limit)
		return
	}

	suggestedUsers := make([]gin.H, 0, limit)
	for _, u := range resp.GetUsers() {
		if u.GetId() == userID {
			continue
		}

		suggestedUsers = append(suggestedUsers, gin.H{
			"id":             u.GetId(),
			"username":       u.GetUsername(),
			"display_name":   u.GetName(),
			"avatar_url":     u.GetProfilePictureUrl(),
			"verified":       u.GetIsVerified(),
			"follower_count": u.GetFollowerCount(),
			"is_following":   false, // Default to false since we're showing recommendations
		})

		if len(suggestedUsers) >= limit {
			break
		}
	}

	if len(suggestedUsers) < limit {
		log.Printf("Not enough users found, adding mock suggestions")
		mockCount := limit - len(suggestedUsers)

		for i := 1; i <= mockCount; i++ {
			mockId := fmt.Sprintf("mock_%d", i)
			exists := false
			for _, su := range suggestedUsers {
				if su["id"] == mockId {
					exists = true
					break
				}
			}

			if !exists {
				suggestedUsers = append(suggestedUsers, gin.H{
					"id":             mockId,
					"username":       fmt.Sprintf("suggested_user%d", i),
					"display_name":   fmt.Sprintf("Suggested User %d", i),
					"avatar_url":     fmt.Sprintf("https://example.com/avatar%d.jpg", i),
					"verified":       i%3 == 0, // Every third user is verified
					"follower_count": 100 + i*10,
					"is_following":   false,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": suggestedUsers,
	})
}

// Helper function to return mock user suggestions
func returnMockSuggestions(c *gin.Context, limit int) {
	var suggestedUsers []gin.H

	for i := 1; i <= limit; i++ {
		suggestedUsers = append(suggestedUsers, gin.H{
			"id":             fmt.Sprintf("user_%d", i),
			"username":       fmt.Sprintf("suggested_user%d", i),
			"display_name":   fmt.Sprintf("Suggested User %d", i),
			"avatar_url":     fmt.Sprintf("https://example.com/avatar%d.jpg", i),
			"verified":       i%3 == 0, // Every third user is verified
			"follower_count": 100 + i*10,
			"is_following":   false,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users": suggestedUsers,
	})
}

// CheckUsernameAvailability checks if a username is available
// @Summary Check username availability
// @Description Checks if a username is available for registration
// @Tags Users
// @Produce json
// @Param username query string true "Username to check"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/users/check-username [get]
func CheckUsernameAvailability(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Username parameter is required",
			"code":    "INVALID_REQUEST",
		})
		return
	}

	if userServiceClient == nil {
		available := true
		if username == "admin" || username == "test" || username == "user" {
			available = false
		}

		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"available": available,
		})
		return
	}

	available, err := userServiceClient.CheckUsernameAvailability(username)
	if err != nil {
		mockAvailable := true
		if username == "admin" || username == "test" || username == "user" {
			mockAvailable = false
		}

		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"available": mockAvailable,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"available": available,
	})
}

// UploadProfileMedia handles uploading a profile picture or banner image
// @Summary Upload profile media
// @Description Upload a profile picture or banner image
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Media file to upload"
// @Param type formData string true "Media type (profile_picture or banner)" Enums(profile_picture, banner)
// @Success 200 {object} models.MediaUploadResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/users/media [post]
func UploadProfileMedia(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid user ID format"})
		return
	}

	mediaType := c.PostForm("type")
	if mediaType != "profile_picture" && mediaType != "banner" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid media type. Must be 'profile_picture' or 'banner'",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No file provided"})
		return
	}

	fileExt := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	if !allowedExts[fileExt] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "File type not allowed. Only jpg, jpeg, png, and gif are allowed.",
		})
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to open file"})
		return
	}
	defer fileContent.Close()

	var url string

	if mediaType == "profile_picture" {
		url, err = utils.UploadProfilePicture(fileContent, file.Filename, userIDStr)
	} else { // banner
		url, err = utils.UploadBanner(fileContent, file.Filename, userIDStr)
	}

	if err != nil {
		log.Printf("Failed to upload %s to Supabase: %v", mediaType, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to upload file: %v", err),
		})
		return
	}

	var updateRequest userProto.UpdateUserRequest
	if mediaType == "profile_picture" {
		updateRequest = userProto.UpdateUserRequest{
			UserId:            userIDStr,
			ProfilePictureUrl: url,
		}
	} else { // banner
		updateRequest = userProto.UpdateUserRequest{
			UserId:    userIDStr,
			BannerUrl: url,
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err = UserClient.UpdateUser(ctx, &updateRequest)
	if err != nil {
		log.Printf("Failed to update user with new %s URL: %v", mediaType, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("File uploaded but failed to update user profile: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"type":    mediaType,
		"url":     url,
	})
}
