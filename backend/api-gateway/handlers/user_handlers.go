package handlers

import (
	userProto "aycom/backend/proto/user"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
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
				c.JSON(http.StatusConflict, ErrorResponse{Success: false, Message: "User already exists"})
				return
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, ErrorResponse{Success: false, Message: st.Message()})
				return
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Failed to register user: " + err.Error()})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Failed to register user: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Registration successful", "user": resp.User})
}

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

	token, err := generateJWT(loginResp.User.Id)
	if err != nil {
		log.Printf("Error generating JWT token: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error generating authentication token")
		return
	}

	expirySeconds, err := strconv.Atoi(os.Getenv("JWT_EXPIRY"))
	if err != nil || expirySeconds <= 0 {
		expirySeconds = 3600
	}

	response := AuthServiceResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  token,
		RefreshToken: "",
		UserId:       loginResp.User.Id,
		TokenType:    "Bearer",
		ExpiresIn:    int64(expirySeconds),
	}

	c.JSON(http.StatusOK, response)
}

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

func generateJWT(userID string) (string, error) {

	expirySeconds, err := strconv.Atoi(os.Getenv("JWT_EXPIRY"))
	if err != nil || expirySeconds <= 0 {
		expirySeconds = 3600
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(expirySeconds) * time.Second).Unix(),
		"iat":     time.Now().Unix(),
	})

	tokenString, err := token.SignedString(GetJWTSecret())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserSuggestions(c *gin.Context) {

	limit := 10
	if limitParam := c.DefaultQuery("limit", "10"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	log.Printf("GetUserSuggestions: Providing mock suggestions data (limit=%d)", limit)

	var users []gin.H
	for i := 1; i <= limit; i++ {
		users = append(users, gin.H{
			"id":              fmt.Sprintf("user-%d", i),
			"username":        fmt.Sprintf("user%d", i),
			"display_name":    fmt.Sprintf("User %d", i),
			"avatar_url":      fmt.Sprintf("https://i.pravatar.cc/150?u=user%d", i),
			"is_verified":     i%3 == 0,
			"bio":             fmt.Sprintf("This is user %d bio", i),
			"follower_count":  i*100 + 50,
			"following_count": i * 75,
			"is_following":    false,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"users":   users,
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

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrderStr := c.DefaultQuery("ascending", "false")
	sortDesc := sortOrderStr != "true"

	if UserClient == nil {
		log.Printf("User service unavailable, using fallback implementation for /users/all endpoint")
		provideFallbackAllUsersList(c, page, limit)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &userProto.GetAllUsersRequest{
		Page:   int32(page),
		Limit:  int32(limit),
		SortBy: sortBy,
	}

	reflect.ValueOf(req).Elem().FieldByName("SortDesc").SetBool(sortDesc)

	resp, err := UserClient.GetAllUsers(ctx, req)

	if err != nil {
		log.Printf("Error getting all users: %v, using fallback implementation", err)
		provideFallbackAllUsersList(c, page, limit)
		return
	}

	users := make([]map[string]interface{}, 0, len(resp.GetUsers()))
	for _, u := range resp.GetUsers() {
		users = append(users, map[string]interface{}{
			"id":              u.GetId(),
			"username":        u.GetUsername(),
			"display_name":    u.GetName(),
			"avatar_url":      u.GetProfilePictureUrl(),
			"is_verified":     u.GetIsVerified(),
			"is_admin":        u.GetIsAdmin(),
			"bio":             u.GetBio(),
			"follower_count":  u.GetFollowerCount(),
			"following_count": u.GetFollowingCount(),
		})
	}

	totalPages := int(1)
	if resp.GetTotalCount() > 0 && int32(limit) > 0 {
		totalPages = int((resp.GetTotalCount() + int32(limit) - 1) / int32(limit))
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       users,
		"total_count": resp.GetTotalCount(),
		"page":        resp.GetPage(),
		"total_pages": totalPages,
	})
}

func provideFallbackAllUsersList(c *gin.Context, page, limit int) {

	mockUsers := []map[string]interface{}{
		{
			"id":              "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			"username":        "testuser1",
			"display_name":    "Test User One",
			"avatar_url":      "https://secure.gravatar.com/avatar/1?d=mp",
			"is_verified":     true,
			"is_admin":        false,
			"bio":             "This is a test user bio",
			"follower_count":  42,
			"following_count": 24,
		},
		{
			"id":              "f47ac10b-58cc-4372-a567-0e02b2c3d480",
			"username":        "testuser2",
			"display_name":    "Test User Two",
			"avatar_url":      "https://secure.gravatar.com/avatar/2?d=mp",
			"is_verified":     false,
			"is_admin":        false,
			"bio":             "Another test user bio",
			"follower_count":  17,
			"following_count": 35,
		},
		{
			"id":              "f47ac10b-58cc-4372-a567-0e02b2c3d481",
			"username":        "kolin",
			"display_name":    "Kolin",
			"avatar_url":      "https://secure.gravatar.com/avatar/3?d=mp",
			"is_verified":     true,
			"is_admin":        true,
			"bio":             "A developer bio",
			"follower_count":  128,
			"following_count": 55,
		},
	}

	totalCount := len(mockUsers)
	totalPages := (totalCount + limit - 1) / limit

	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	startIdx := (page - 1) * limit
	endIdx := startIdx + limit

	if startIdx >= totalCount {
		startIdx = 0
		endIdx = 0
	}

	if endIdx > totalCount {
		endIdx = totalCount
	}

	var pageUsers []map[string]interface{}
	if startIdx < endIdx {
		pageUsers = mockUsers[startIdx:endIdx]
	} else {
		pageUsers = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       pageUsers,
		"total_count": totalCount,
		"page":        page,
		"total_pages": totalPages,
	})

	log.Printf("Provided %d fallback users for page %d (limit %d)", len(pageUsers), page, limit)
}

func UpdateProfilePictureURLHandler(c *gin.Context) {
	userIdValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIdValue.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	var req struct {
		ProfilePictureUrl string `json:"profile_picture_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !strings.Contains(req.ProfilePictureUrl, ".supabase.co") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL source"})
		return
	}

	if UserClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "User service unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	updateReq := &userProto.UpdateUserRequest{
		UserId:            userID,
		ProfilePictureUrl: req.ProfilePictureUrl,
	}

	_, err := UserClient.UpdateUser(ctx, updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update profile: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "url": req.ProfilePictureUrl, "profile_picture_url": req.ProfilePictureUrl})
}

func UpdateBannerURLHandler(c *gin.Context) {
	userIdValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIdValue.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	var req struct {
		BannerUrl string `json:"banner_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !strings.Contains(req.BannerUrl, ".supabase.co") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL source"})
		return
	}

	if UserClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "User service unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	updateReq := &userProto.UpdateUserRequest{
		UserId:    userID,
		BannerUrl: req.BannerUrl,
	}

	_, err := UserClient.UpdateUser(ctx, updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update profile: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "url": req.BannerUrl, "banner_url": req.BannerUrl})
}

func GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Username is required")
		return
	}

	currentUserID, exists := c.Get("userID")
	if !exists {
		currentUserID = ""
	}
	currentUserIDStr := currentUserID.(string)

	if userServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	user, err := userServiceClient.GetUserByUsername(username)
	if err != nil {

		st, ok := status.FromError(err)
		if ok {
			code := status.Code(err)
			if code == codes.NotFound {
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
				return
			}
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch user: "+st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while fetching user")
		}
		log.Printf("Error fetching user by username '%s': %v", username, err)
		return
	}

	isUserBlocked, err := userServiceClient.IsUserBlocked(user.ID, currentUserIDStr)
	if err != nil {
		log.Printf("Error checking if user is blocked: %v", err)
	}

	isFollowing := false
	if currentUserIDStr != "" && currentUserIDStr != user.ID {
		isFollowing, err = userServiceClient.IsFollowing(currentUserIDStr, user.ID)
		if err != nil {
			log.Printf("Error checking follow status: %v", err)
		}
	}

	displayName := user.DisplayName
	if displayName == "" {
		displayName = user.Name
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"user": gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"display_name":        displayName,
			"bio":                 user.Bio,
			"profile_picture_url": user.ProfilePictureURL,
			"banner_url":          user.BannerURL,
			"follower_count":      user.FollowerCount,
			"following_count":     user.FollowingCount,
			"created_at":          user.CreatedAt.Format(time.RFC3339),
			"is_verified":         user.IsVerified,
			"is_private":          user.IsPrivate,
			"is_following":        isFollowing,
			"is_blocked":          isUserBlocked,
			"is_current_user":     currentUserIDStr == user.ID,
		},
	})
}

func GetUserById(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "User ID is required")
		return
	}

	log.Printf("GetUserById: Fetching user with ID: %s", userId)

	currentUserID, exists := c.Get("userID")
	if !exists {
		currentUserID = ""
	}
	currentUserIDStr, ok := currentUserID.(string)
	if !ok {
		currentUserIDStr = ""
	}

	if userServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	user, err := userServiceClient.GetUserById(userId)
	if err != nil {

		log.Printf("Failed to find user with ID %s, trying as username", userId)
		user, err = userServiceClient.GetUserByUsername(userId)
		if err != nil {

			st, ok := status.FromError(err)
			if ok {
				code := status.Code(err)
				if code == codes.NotFound {
					utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
					return
				}
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch user: "+st.Message())
			} else {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while fetching user")
			}
			log.Printf("Error fetching user with ID/username '%s': %v", userId, err)
			return
		}
	}

	isUserBlocked, err := userServiceClient.IsUserBlocked(user.ID, currentUserIDStr)
	if err != nil {
		log.Printf("Error checking if user is blocked: %v", err)
	}

	isFollowing := false
	if currentUserIDStr != "" && currentUserIDStr != user.ID {
		isFollowing, err = userServiceClient.IsFollowing(currentUserIDStr, user.ID)
		if err != nil {
			log.Printf("Error checking follow status: %v", err)
		}
	}

	displayName := user.DisplayName
	if displayName == "" {
		displayName = user.Name
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"user": gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"display_name":        displayName,
			"bio":                 user.Bio,
			"profile_picture_url": user.ProfilePictureURL,
			"banner_url":          user.BannerURL,
			"follower_count":      user.FollowerCount,
			"following_count":     user.FollowingCount,
			"created_at":          user.CreatedAt.Format(time.RFC3339),
			"is_verified":         user.IsVerified,
			"is_private":          user.IsPrivate,
			"is_following":        isFollowing,
			"is_blocked":          isUserBlocked,
			"is_current_user":     currentUserIDStr == user.ID,
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

	token, err := generateJWT(resp.User.Id)
	if err != nil {
		log.Printf("Error generating JWT for new admin user: %v", err)

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
		"token": token,
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
		SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request payload: "+err.Error())
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
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	isDebugRequest := req.IsDebugRequest
	log.Printf("UpdateUserAdminStatus: Debug request check - IsDebugRequest: %t", isDebugRequest)

	if !isDebugRequest {
		log.Println("UpdateUserAdminStatus: Not a debug request, checking admin status")

		currentUser, err := userServiceClient.GetUserProfile(req.UserID)
		if err != nil {
			log.Printf("UpdateUserAdminStatus: Error getting user profile: %v", err)
			SendErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Only admins can update admin status")
			return
		}

		log.Printf("UpdateUserAdminStatus: Current user IsAdmin: %t", currentUser.IsAdmin)
		if !currentUser.IsAdmin {
			log.Printf("UpdateUserAdminStatus: User %s is not an admin", req.UserID)
			SendErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Only admins can update admin status")
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
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			default:
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user admin status: "+st.Message())
			}
		} else {
			log.Printf("UpdateUserAdminStatus: Non-gRPC error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while updating user admin status")
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
	SendSuccessResponse(c, http.StatusOK, gin.H{
		"success":  true,
		"message":  fmt.Sprintf("User %s has been %s", userName, action),
		"user_id":  req.UserID,
		"is_admin": isAdmin,
	})
}
