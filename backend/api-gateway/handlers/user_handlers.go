package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"aycom/backend/services/user/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var userClient proto.UserServiceClient

func initUserServiceClient() error {
	if userClient != nil {
		return nil
	}

	// Get host and port from config (set by docker-compose env vars)
	userServiceHost := Config.Services.UserServiceHost // Should be "user_service"
	userServicePort := Config.Services.UserServicePort // Should be "9091"

	// Ensure both host and port are available
	if userServiceHost == "" || userServicePort == "" {
		log.Printf("Error: User service host (%s) or port (%s) is missing in config", userServiceHost, userServicePort)
		return fmt.Errorf("user service configuration is incomplete")
	}

	// Construct the target address for gRPC Dial
	target := fmt.Sprintf("%s:%s", userServiceHost, userServicePort) // Combine host and port

	log.Printf("Connecting to User service at %s", target) // Log the actual target address
	conn, err := grpc.Dial(
		target, // Use the combined host:port
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Printf("Failed to connect to User service: %v", err)
		return err
	}

	userClient = proto.NewUserServiceClient(conn)
	log.Printf("Connected to User service at %s", target)
	return nil
}

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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	// Initialize user service client if needed
	if err := initUserServiceClient(); err != nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Success: false,
			Message: "User service is currently unavailable",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	// Create request with user ID from token
	req := &proto.GetUserRequest{
		UserId: userID.(string),
	}

	// Set reasonable timeout for gRPC call
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call User service's GetUser method
	resp, err := userClient.GetUser(ctx, req)
	if err != nil {
		// Handle gRPC errors
		st, ok := status.FromError(err)
		if ok {
			// Map gRPC status code to HTTP status code
			switch st.Code() {
			case 5: // Not found
				c.JSON(http.StatusNotFound, ErrorResponse{
					Success: false,
					Message: "User profile not found",
					Code:    "NOT_FOUND",
				})
			case 16: // Unauthenticated
				c.JSON(http.StatusUnauthorized, ErrorResponse{
					Success: false,
					Message: "Unauthorized to access this profile",
					Code:    "UNAUTHORIZED",
				})
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Failed to retrieve user profile: " + st.Message(),
					Code:    "INTERNAL_ERROR",
				})
			}
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Internal server error while retrieving profile",
				Code:    "INTERNAL_ERROR",
			})
		}
		log.Printf("Error retrieving user profile: %v", err)
		return
	}

	// Format the date of birth for the response if present
	// The proto object should already have a properly formatted date

	// Return successful response with user profile
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":                  resp.User.Id,
			"name":                resp.User.Name,
			"username":            resp.User.Username,
			"email":               resp.User.Email,
			"gender":              resp.User.Gender,
			"date_of_birth":       resp.User.DateOfBirth,
			"profile_picture_url": resp.User.ProfilePictureUrl,
			"banner_url":          resp.User.BannerUrl,
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	// Initialize user service client if needed
	if err := initUserServiceClient(); err != nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Success: false,
			Message: "User service is currently unavailable",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	// Bind input from request body
	var input struct {
		Name              string `json:"name"`
		ProfilePictureURL string `json:"profile_picture_url"`
		BannerURL         string `json:"banner_url"`
		// Add other fields that should be updatable
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request payload: " + err.Error(),
			Code:    "BAD_REQUEST",
		})
		return
	}

	// Create update request
	req := &proto.UpdateUserRequest{
		UserId:            userID.(string),
		Name:              input.Name,
		ProfilePictureUrl: input.ProfilePictureURL,
		BannerUrl:         input.BannerURL,
		// Map other fields from input to request
	}

	// Set reasonable timeout for gRPC call
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call User service's UpdateUser method
	resp, err := userClient.UpdateUser(ctx, req)
	if err != nil {
		// Handle gRPC errors
		st, ok := status.FromError(err)
		if ok {
			// Map gRPC status code to HTTP status code
			switch st.Code() {
			case 5: // Not found
				c.JSON(http.StatusNotFound, ErrorResponse{
					Success: false,
					Message: "User profile not found",
					Code:    "NOT_FOUND",
				})
			case 16: // Unauthenticated
				c.JSON(http.StatusUnauthorized, ErrorResponse{
					Success: false,
					Message: "Unauthorized to update this profile",
					Code:    "UNAUTHORIZED",
				})
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Failed to update user profile: " + st.Message(),
					Code:    "INTERNAL_ERROR",
				})
			}
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Internal server error while updating profile",
				Code:    "INTERNAL_ERROR",
			})
		}
		log.Printf("Error updating user profile: %v", err)
		return
	}

	// Return successful response with updated user profile
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":                  resp.User.Id,
			"name":                resp.User.Name,
			"username":            resp.User.Username,
			"email":               resp.User.Email,
			"gender":              resp.User.Gender,
			"date_of_birth":       resp.User.DateOfBirth,
			"profile_picture_url": resp.User.ProfilePictureUrl,
			"banner_url":          resp.User.BannerUrl,
		},
	})
}

// Register registers a new user
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
	// Disable reCAPTCHA validation: always accept
	// if req.RecaptchaToken == "" { ... }
	if err := initUserServiceClient(); err != nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Success: false, Message: "User service unavailable"})
		return
	}
	user := &proto.User{
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
	createReq := &proto.CreateUserRequest{User: user}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	resp, err := userClient.CreateUser(ctx, createReq)
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

// Login authenticates a user (mock password check, returns mock JWT)
func LoginUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Success: false, Message: "Invalid request: " + err.Error()})
		return
	}
	if err := initUserServiceClient(); err != nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Success: false, Message: "User service unavailable"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	userResp, err := userClient.GetUser(ctx, &proto.GetUserRequest{UserId: req.Email})
	if err != nil || userResp.User == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Success: false, Message: "Invalid credentials"})
		return
	}
	// TODO: Replace with real password check
	if req.Password != "password" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Success: false, Message: "Invalid credentials"})
		return
	}
	// Return mock JWT
	token := "mock-jwt-token"
	c.JSON(http.StatusOK, gin.H{"success": true, "token": token, "user": userResp.User})
}
