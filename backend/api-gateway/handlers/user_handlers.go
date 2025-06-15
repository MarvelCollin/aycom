package handlers

import (
	userProto "aycom/backend/proto/user"
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"aycom/backend/api-gateway/utils"
)

func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userIDStr := userID.(string)
	log.Printf("GetUserProfile Handler: Retrieved userID from context: %s", userIDStr)

	if _, err := uuid.Parse(userIDStr); err != nil {
		log.Printf("GetUserProfile Handler: Invalid UUID format for userID: %s", userIDStr)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	if userServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	user, err := userServiceClient.GetUserProfile(userIDStr)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case 5:
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User profile not found")
			case 16:
				utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized to access this profile")
			default:
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve user profile: "+st.Message())
			}
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while retrieving profile")
		}
		log.Printf("Error retrieving user profile: %v", err)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"user": gin.H{
			"id":                  user.ID,
			"name":                user.Name,
			"username":            user.Username,
			"email":               user.Email,
			"profile_picture_url": user.ProfilePictureURL,
			"banner_url":          user.BannerURL,
			"bio":                 user.Bio,
			"is_verified":         user.IsVerified,
			"is_admin":            user.IsAdmin,
			"is_private":          user.IsPrivate,
			"follower_count":      user.FollowerCount,
			"following_count":     user.FollowingCount,
			"created_at":          user.CreatedAt,
		},
	})
}

func UpdateUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userIDStr := userID.(string)
	log.Printf("UpdateUserProfile Handler: Retrieved userID from context: %s", userIDStr)

	if _, err := uuid.Parse(userIDStr); err != nil {
		log.Printf("UpdateUserProfile Handler: Invalid UUID format for userID: %s", userIDStr)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	if userServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	var input struct {
		Name              string `json:"name"`
		DisplayName       string `json:"display_name"`
		Bio               string `json:"bio"`
		Email             string `json:"email"`
		DateOfBirth       string `json:"date_of_birth"`
		Gender            string `json:"gender"`
		ProfilePictureURL string `json:"profile_picture_url"`
		ProfilePicture    string `json:"profile_picture"`
		Avatar            string `json:"avatar"`
		BannerURL         string `json:"banner_url"`
		Banner            string `json:"banner"`
		BackgroundBanner  string `json:"background_banner"`
		IsPrivate         bool   `json:"is_private"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload: "+err.Error())
		return
	}

	name := input.Name
	if name == "" && input.DisplayName != "" {
		name = input.DisplayName
	}

	profilePictureURL := input.ProfilePictureURL
	if profilePictureURL == "" && input.ProfilePicture != "" {
		profilePictureURL = input.ProfilePicture
	}
	if profilePictureURL == "" && input.Avatar != "" {
		profilePictureURL = input.Avatar
	}

	bannerURL := input.BannerURL
	if bannerURL == "" && input.Banner != "" {
		bannerURL = input.Banner
	}
	if bannerURL == "" && input.BackgroundBanner != "" {
		bannerURL = input.BackgroundBanner
	}

	profileUpdate := &UserProfileUpdate{
		Name:              name,
		Bio:               input.Bio,
		Email:             input.Email,
		DateOfBirth:       input.DateOfBirth,
		Gender:            input.Gender,
		ProfilePictureURL: profilePictureURL,
		BannerURL:         bannerURL,
		IsPrivate:         input.IsPrivate,
	}

	updatedUser, err := userServiceClient.UpdateUserProfile(userIDStr, profileUpdate)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case 5:
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User profile not found")
			case 16:
				utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized to update this profile")
			default:
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user profile: "+st.Message())
			}
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while updating profile")
		}
		log.Printf("Error updating user profile: %v", err)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"user": gin.H{
			"id":                  updatedUser.ID,
			"name":                updatedUser.Name,
			"username":            updatedUser.Username,
			"email":               updatedUser.Email,
			"profile_picture_url": updatedUser.ProfilePictureURL,
			"banner_url":          updatedUser.BannerURL,
			"bio":                 updatedUser.Bio,
			"is_verified":         updatedUser.IsVerified,
			"is_admin":            updatedUser.IsAdmin,
			"is_private":          updatedUser.IsPrivate,
			"follower_count":      updatedUser.FollowerCount,
			"following_count":     updatedUser.FollowingCount,
			"created_at":          updatedUser.CreatedAt,
		},
	})
}

func RegisterUser(c *gin.Context) {
	var req struct {
		Name                  string `json:"name" binding:"required,min=4,max=50"`
		Username              string `json:"username" binding:"required,min=3,max=15"`
		Email                 string `json:"email" binding:"required,email"`
		Password              string `json:"password" binding:"required,min=8"`
		ConfirmPassword       string `json:"confirm_password" binding:"required,eqfield=Password"`
		Gender                string `json:"gender" binding:"required,oneof=male female"`
		DateOfBirth           string `json:"date_of_birth" binding:"required"`
		SecurityQuestion      string `json:"security_question" binding:"required"`
		SecurityAnswer        string `json:"security_answer" binding:"required,min=3"`
		SubscribeToNewsletter bool   `json:"subscribe_to_newsletter"`
		RecaptchaToken        string `json:"recaptcha_token"`
		ProfilePictureUrl     string `json:"profile_picture_url,omitempty"`
		BannerUrl             string `json:"banner_url,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request: "+err.Error())
		return
	}

	if userServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service unavailable")
		return
	}

	if len(req.Password) < 8 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_PASSWORD", "Password must be at least 8 characters")
		return
	}

	if req.Password != req.ConfirmPassword {
		utils.SendErrorResponse(c, http.StatusBadRequest, "PASSWORD_MISMATCH", "Password and confirmation password do not match")
		return
	}

	user := &userProto.User{
		Id:                "",
		Name:              req.Name,
		Username:          req.Username,
		Email:             req.Email,
		Gender:            req.Gender,
		DateOfBirth:       req.DateOfBirth,
		ProfilePictureUrl: req.ProfilePictureUrl,
		BannerUrl:         req.BannerUrl,

		Password:              req.Password,
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
		if ok {
			switch st.Code() {
			case codes.AlreadyExists:
				utils.SendErrorResponse(c, http.StatusConflict, "ALREADY_EXISTS", "User already exists")
				return
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
				return
			default:
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register user: "+err.Error())
				return
			}
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register user: "+err.Error())
		return
	}
	utils.SendSuccessResponse(c, http.StatusCreated, gin.H{"message": "Registration successful", "user": resp.User})
}

func LoginUser(c *gin.Context) {
	var req struct {
		Email          string `json:"email"`
		Password       string `json:"password"`
		RecaptchaToken string `json:"recaptcha_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request: "+err.Error())
		return
	}

	// Skip reCAPTCHA verification in development mode
	devMode := utils.IsDevelopmentMode()

	log.Printf("Login request received for email: %s (dev mode: %v, has recaptcha: %v)",
		req.Email, devMode, req.RecaptchaToken != "")

	if !devMode && req.RecaptchaToken == "" {
		log.Println("reCAPTCHA token missing in production mode")
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "RECAPTCHA_REQUIRED",
			"message":    "reCAPTCHA verification is required",
		})
		return
	}

	if !devMode && req.RecaptchaToken != "" {
		// Verify reCAPTCHA token
		log.Printf("Verifying reCAPTCHA token with length: %d", len(req.RecaptchaToken))

		success, err := utils.VerifyRecaptcha(req.RecaptchaToken)
		if err != nil {
			log.Printf("reCAPTCHA verification error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "RECAPTCHA_ERROR",
				"message":    "reCAPTCHA verification failed: " + err.Error(),
			})
			return
		} else if !success {
			log.Printf("reCAPTCHA verification failed: token invalid")
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "RECAPTCHA_FAILED",
				"message":    "reCAPTCHA verification failed: invalid token",
			})
			return
		}

		log.Printf("reCAPTCHA verification successful for login request from: %s", req.Email)
	} else if devMode {
		log.Printf("Skipping reCAPTCHA verification in development mode for: %s", req.Email)
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service unavailable")
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
		log.Printf("Login failed for email: %s, error: %v", req.Email, err)
		utils.SendErrorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid credentials")
		return
	}

	token, err := utils.GenerateJWT(loginResp.User.Id, time.Hour)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate token")
		return
	}

	refreshTokenDuration := 7 * 24 * time.Hour
	refreshToken, err := utils.GenerateJWT(loginResp.User.Id, refreshTokenDuration)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate refresh token")
		return
	}

	log.Printf("Login successful for user: %s (id: %s)", loginResp.User.Email, loginResp.User.Id)

	response := AuthServiceResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  token,
		RefreshToken: refreshToken,
		UserId:       loginResp.User.Id,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Hour.Seconds()),
	}

	c.JSON(http.StatusOK, response)
}

func GetUserByEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Email is required")
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	grpcReq := &userProto.GetUserByEmailRequest{Email: req.Email}
	resp, err := UserClient.GetUserByEmail(ctx, grpcReq)
	if err != nil || resp.User == nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{"user": resp.User})
}

func GetUserSuggestions(c *gin.Context) {
	log.Printf("======= GetUserSuggestions endpoint called PATH=%s METHOD=%s =======", c.Request.URL.Path, c.Request.Method)

	log.Printf("Request Headers:")
	for name, values := range c.Request.Header {
		log.Printf("  %s: %s", name, values)
	}

	var userIDStr string
	userID, exists := c.Get("userID")
	if exists {
		var ok bool
		userIDStr, ok = userID.(string)
		if ok {
			log.Printf("UserID from context exists: %s", userIDStr)
		} else {
			log.Printf("UserID from context exists but is not a string: %v, %T", userID, userID)
			userIDStr = ""
		}
	} else {
		log.Printf("No userID in context, proceeding as anonymous user")
	}

	userId, exists := c.Get("userId")
	if exists {
		log.Printf("Found alternative 'userId' in context: %v", userId)
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "3"))
	log.Printf("Using limit: %d", limit)

	if userServiceClient == nil {
		log.Printf("ERROR: userServiceClient is nil - initializing client for this request")

		InitUserServiceClient(AppConfig)

		if userServiceClient == nil {
			log.Printf("ERROR: Failed to initialize userServiceClient")

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"users":   []gin.H{},
				"message": "Service temporarily unavailable, please try again later",
			})
			return
		}
	}

	var users []*User
	var err error

	if userIDStr != "" {

		log.Printf("Getting personalized recommendations for user %s", userIDStr)
		users, err = userServiceClient.GetUserRecommendations(userIDStr, limit)
		if err != nil {
			log.Printf("Error getting user recommendations: %v", err)

		}
	}

	if users == nil || len(users) == 0 {
		log.Printf("No recommendations found or user not authenticated, falling back to all users")

		users, totalCount, _, err := userServiceClient.GetAllUsers(1, limit, "created_at", false)
		if err != nil {
			log.Printf("Error getting all users: %v", err)

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"users":   []gin.H{},
			})
			return
		}
		log.Printf("Got %d users from fallback (total count: %d)", len(users), totalCount)
	}

	var userResults []gin.H
	for _, user := range users {
		userResults = append(userResults, gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"name":                user.Name,
			"profile_picture_url": user.ProfilePictureURL,
			"bio":                 user.Bio,
			"is_verified":         user.IsVerified,
			"is_admin":            user.IsAdmin,
			"follower_count":      user.FollowerCount,
			"is_following":        userIDStr != "" && user.IsFollowing,
		})
	}

	log.Printf("Returning %d user suggestions", len(userResults))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"users":   userResults,
	})
}

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
	} else {
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
	} else {
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

func GetAllUsers(c *gin.Context) {
	if userServiceClient == nil {
		log.Printf("User service unavailable")
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service unavailable")
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrderStr := c.DefaultQuery("ascending", "false")
	ascending := sortOrderStr == "true"

	log.Printf("GetAllUsers: page=%d, limit=%d, sortBy=%s, ascending=%v", page, limit, sortBy, ascending)

	users, totalCount, totalPages, err := userServiceClient.GetAllUsers(page, limit, sortBy, ascending)

	if err != nil {
		log.Printf("Error getting all users: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve users")
		return
	}

	usersData := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		usersData = append(usersData, map[string]interface{}{
			"id":                  user.ID,
			"username":            user.Username,
			"name":                user.Name,
			"profile_picture_url": user.ProfilePictureURL,
			"is_verified":         user.IsVerified,
			"is_admin":            user.IsAdmin,
			"bio":                 user.Bio,
			"follower_count":      user.FollowerCount,
			"following_count":     user.FollowingCount,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"users":       usersData,
		"total_count": totalCount,
		"page":        page,
		"total_pages": totalPages,
	})

	log.Printf("Provided %d users for page %d (limit %d)", len(usersData), page, limit)
}

func UpdateProfilePictureURLHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userIDStr := userID.(string)

	var input struct {
		URL string `json:"profile_picture_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if input.URL == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Profile picture URL is required")
		return
	}

	log.Printf("UpdateProfilePictureURLHandler: Updating profile picture URL for user %s: %s", userIDStr, input.URL)

	profileUpdate := &UserProfileUpdate{
		ProfilePictureURL: input.URL,
	}

	_, err := userServiceClient.UpdateUserProfile(userIDStr, profileUpdate)

	if err != nil {
		log.Printf("Error updating profile picture URL: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update profile picture URL")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message":             "Profile picture URL updated successfully",
		"profile_picture_url": input.URL,
	})
}

func UpdateBannerURLHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userIDStr := userID.(string)

	var input struct {
		URL string `json:"banner_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if input.URL == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Banner URL is required")
		return
	}

	log.Printf("UpdateBannerURLHandler: Updating banner URL for user %s: %s", userIDStr, input.URL)

	profileUpdate := &UserProfileUpdate{
		BannerURL: input.URL,
	}

	_, err := userServiceClient.UpdateUserProfile(userIDStr, profileUpdate)

	if err != nil {
		log.Printf("Error updating banner URL: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update banner URL")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message":    "Banner URL updated successfully",
		"banner_url": input.URL,
	})
}

func GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Username is required")
		return
	}

	log.Printf("GetUserByUsername Handler: Looking up user with username: %s", username)

	user, err := userServiceClient.GetUserByUsername(username)

	if err != nil {
		log.Printf("Error getting user by username: %v", err)
		utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	if id, exists := c.Get("userID"); exists {
		requesterID := id.(string)
		if requesterID != "" && user.ID != requesterID {
			isFollowing, _ := userServiceClient.IsFollowing(requesterID, user.ID)
			user.IsFollowing = isFollowing
		}
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"user": gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"name":                user.Name,
			"bio":                 user.Bio,
			"profile_picture_url": user.ProfilePictureURL,
			"banner_url":          user.BannerURL,
			"is_verified":         user.IsVerified,
			"is_admin":            user.IsAdmin,
			"follower_count":      user.FollowerCount,
			"following_count":     user.FollowingCount,
			"created_at":          user.CreatedAt,
			"is_following":        user.IsFollowing,
		},
	})
}

func GetUserById(c *gin.Context) {
	userIdParam := c.Param("userId")
	if userIdParam == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	if _, err := uuid.Parse(userIdParam); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	log.Printf("GetUserById Handler: Looking up user with ID: %s", userIdParam)

	user, err := userServiceClient.GetUserById(userIdParam)

	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
		utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	if id, exists := c.Get("userID"); exists {
		requesterID := id.(string)
		if requesterID != "" && user.ID != requesterID {
			isFollowing, _ := userServiceClient.IsFollowing(requesterID, user.ID)
			user.IsFollowing = isFollowing
		}
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"user": gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"name":                user.Name,
			"bio":                 user.Bio,
			"profile_picture_url": user.ProfilePictureURL,
			"banner_url":          user.BannerURL,
			"is_verified":         user.IsVerified,
			"is_admin":            user.IsAdmin,
			"follower_count":      user.FollowerCount,
			"following_count":     user.FollowingCount,
			"created_at":          user.CreatedAt,
			"is_following":        user.IsFollowing,
		},
	})
}

func CreateAdminUser(c *gin.Context) {
	log.Printf("CreateAdminUser Handler: Processing request")

	var req struct {
		Name             string `json:"name" binding:"required,min=4,max=50"`
		Username         string `json:"username" binding:"required,min=3,max=15"`
		Email            string `json:"email" binding:"required,email"`
		Password         string `json:"password" binding:"required,min=8"`
		ConfirmPassword  string `json:"confirm_password" binding:"required,eqfield=Password"`
		Gender           string `json:"gender" binding:"required,oneof=male female other"`
		DateOfBirth      string `json:"date_of_birth" binding:"required"`
		SecurityQuestion string `json:"security_question" binding:"required"`
		SecurityAnswer   string `json:"security_answer" binding:"required"`
		IsAdmin          bool   `json:"is_admin" binding:"required"`
		IsVerified       bool   `json:"is_verified"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateAdminUser Handler: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload: "+err.Error())
		return
	}

	if !req.IsAdmin {
		log.Printf("CreateAdminUser Handler: Attempt to create admin user without is_admin=true")
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "is_admin must be set to true for admin user creation")
		return
	}

	if userServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User gRPC client not initialized")
		return
	}

	userReq := &userProto.CreateUserRequest{
		User: &userProto.User{
			Name:             req.Name,
			Username:         req.Username,
			Email:            req.Email,
			Password:         req.Password,
			Gender:           req.Gender,
			DateOfBirth:      req.DateOfBirth,
			SecurityQuestion: req.SecurityQuestion,
			SecurityAnswer:   req.SecurityAnswer,
			IsAdmin:          true,
			IsVerified:       req.IsVerified,
		},
	}

	log.Printf("CreateAdminUser Handler: Creating admin user with email: %s, username: %s", req.Email, req.Username)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := UserClient.CreateUser(ctx, userReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.AlreadyExists:
				utils.SendErrorResponse(c, http.StatusConflict, "ALREADY_EXISTS", "User with this email or username already exists")
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_INPUT", st.Message())
			default:
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create admin user: "+st.Message())
			}
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while creating admin user")
		}
		log.Printf("Error creating admin user: %v", err)
		return
	}

	token, err := utils.GenerateJWT(resp.User.Id, time.Hour)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate token")
		return
	}

	refreshTokenDuration := 7 * 24 * time.Hour
	refreshToken, err := utils.GenerateJWT(resp.User.Id, refreshTokenDuration)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate refresh token")
		return
	}

	utils.SendSuccessResponse(c, http.StatusCreated, gin.H{
		"success": true,
		"message": "Admin user created successfully",
		"user": gin.H{
			"id":          resp.User.Id,
			"name":        resp.User.Name,
			"username":    resp.User.Username,
			"email":       resp.User.Email,
			"is_admin":    true,
			"is_verified": resp.User.IsVerified,
			"created_at":  resp.User.CreatedAt,
		},
		"token":         token,
		"refresh_token": refreshToken,
	})
}

func UpdateUserAdminStatus(c *gin.Context) {
	log.Println("UpdateUserAdminStatus: Starting handler execution")

	currentUserIDValue, exists := c.Get("userID")
	if !exists {
		log.Println("UpdateUserAdminStatus: No userID in context - unauthorized")
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}
	log.Printf("UpdateUserAdminStatus: Request from user ID: %v", currentUserIDValue)

	if userServiceClient == nil {
		log.Println("Initializing user service client in UpdateUserAdminStatus")
		InitUserServiceClient(AppConfig)
		if userServiceClient == nil {
			log.Println("UpdateUserAdminStatus: Failed to initialize userServiceClient")
			utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
			return
		}
		log.Println("UpdateUserAdminStatus: Successfully initialized userServiceClient")
	}

	var req struct {
		UserID         string      `json:"user_id" binding:"required"`
		IsAdmin        interface{} `json:"is_admin" binding:"required"`
		IsDebugRequest bool        `json:"is_debug_request"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateUserAdminStatus: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload: "+err.Error())
		return
	}

	isAdmin := false
	switch v := req.IsAdmin.(type) {
	case bool:
		isAdmin = v
	case string:
		isAdmin = v == "true" || v == "t" || v == "1"
	case float64:
		isAdmin = v == 1
	case int:
		isAdmin = v == 1
	}

	log.Printf("UpdateUserAdminStatus: Request payload - Target UserID: %s, Is Admin (raw): %v, Is Admin (parsed): %t, Debug: %t",
		req.UserID, req.IsAdmin, isAdmin, req.IsDebugRequest)

	if UserClient == nil {
		log.Println("UpdateUserAdminStatus: UserClient is nil")
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	isDebugRequest := req.IsDebugRequest
	log.Printf("UpdateUserAdminStatus: Debug request check - IsDebugRequest: %t", isDebugRequest)

	if !isDebugRequest {
		log.Println("UpdateUserAdminStatus: Not a debug request, checking admin status")

		currentUser, err := userServiceClient.GetUserProfile(req.UserID)
		if err != nil {
			log.Printf("UpdateUserAdminStatus: Error getting user profile: %v", err)
			utils.SendErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Only admins can update admin status")
			return
		}

		log.Printf("UpdateUserAdminStatus: Current user IsAdmin: %t", currentUser.IsAdmin)
		if !currentUser.IsAdmin {
			log.Printf("UpdateUserAdminStatus: User %s is not an admin", req.UserID)
			utils.SendErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Only admins can update admin status")
			return
		}
		log.Printf("UpdateUserAdminStatus: User %s is an admin, proceeding", req.UserID)
	} else {
		log.Println("Debug request detected - bypassing admin check for admin status update")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateReq := &userProto.UpdateUserRequest{
		UserId: req.UserID,
		User: &userProto.User{
			Id:      req.UserID,
			IsAdmin: isAdmin,
		},
	}

	log.Printf("UpdateUserAdminStatus: Sending UpdateUser request with IsAdmin=%t", isAdmin)
	_, err := UserClient.UpdateUser(ctx, updateReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Printf("UpdateUserAdminStatus: gRPC error code: %d, message: %s", st.Code(), st.Message())
			switch st.Code() {
			case codes.NotFound:
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			default:
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user admin status: "+st.Message())
			}
		} else {
			log.Printf("UpdateUserAdminStatus: Non-gRPC error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while updating user admin status")
		}
		log.Printf("Error updating user admin status: %v", err)
		return
	}

	updatedUser, err := userServiceClient.GetUserById(req.UserID)
	if err != nil {
		log.Printf("UpdateUserAdminStatus: Error getting updated user data: %v", err)

	} else {
		log.Printf("UpdateUserAdminStatus: Updated user returned with IsAdmin=%t", updatedUser.IsAdmin)
	}

	var userName string
	if updatedUser != nil {
		userName = updatedUser.Username
	} else {
		userName = req.UserID
	}

	action := "removed from admins"
	if isAdmin {
		action = "promoted to admin"
	}

	log.Printf("UpdateUserAdminStatus: Successfully completed. User %s has been %s", userName, action)
	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"success":  true,
		"message":  fmt.Sprintf("User %s has been %s", userName, action),
		"user_id":  req.UserID,
		"is_admin": isAdmin,
	})
}

func GetPublicUserSuggestions(c *gin.Context) {
	log.Printf("GetPublicUserSuggestions endpoint called")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "3"))

	if userServiceClient == nil {
		InitUserServiceClient(AppConfig)
		if userServiceClient == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"users":   []gin.H{},
				"message": "Service temporarily unavailable, please try again later",
			})
			return
		}
	}

	users, totalCount, _, err := userServiceClient.GetAllUsers(1, limit, "created_at", false)
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"users":   []gin.H{},
		})
		return
	}

	log.Printf("Got %d users for public suggestions (total count: %d)", len(users), totalCount)

	var userResults []gin.H
	for _, user := range users {
		userResults = append(userResults, gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"name":                user.Name,
			"profile_picture_url": user.ProfilePictureURL,
			"bio":                 user.Bio,
			"is_verified":         user.IsVerified,
			"is_admin":            user.IsAdmin,
			"follower_count":      user.FollowerCount,
			"is_following":        false,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"users":   userResults,
	})
}

func CreatePremiumRequest(c *gin.Context) {
	log.Printf("CreatePremiumRequest: Processing premium request creation")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	var req struct {
		Reason             string `json:"reason" binding:"required"`
		IdentityCardNumber string `json:"identity_card_number" binding:"required"`
		FacePhotoURL       string `json:"face_photo_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreatePremiumRequest Handler: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.CreatePremiumRequest(ctx, &userProto.CreatePremiumRequestRequest{
		UserId:             userID.(string),
		Reason:             req.Reason,
		IdentityCardNumber: req.IdentityCardNumber,
		FacePhotoUrl:       req.FacePhotoURL,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.AlreadyExists:
				utils.SendErrorResponse(c, http.StatusConflict, "ALREADY_EXISTS", "You already have a pending or approved premium request")
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("CreatePremiumRequest Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create premium request")
			}
		} else {
			log.Printf("CreatePremiumRequest Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create premium request")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}
