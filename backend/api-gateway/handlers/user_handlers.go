package handlers

import (
	"context"
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

	userServiceHost := Config.Services.UserServiceHost
	if userServiceHost == "" {
		// Use environment variables or defaults from config
		userServiceHost = "user_service:" + Config.Services.UserServicePort // Use Docker service name and configured port
	}

	log.Printf("Connecting to User service at %s", userServiceHost)
	conn, err := grpc.Dial(
		userServiceHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Printf("Failed to connect to User service: %v", err)
		return err
	}

	userClient = proto.NewUserServiceClient(conn)
	log.Printf("Connected to User service at %s", userServiceHost)
	return nil
}

// GetUserProfile retrieves the user's profile from the User service via gRPC
// @Summary Get user profile
// @Description Returns the profile of the authenticated user
// @Tags Users
// @Produce json
// @Success 200 {object} models.UserProfileResponse
// @Failure 401 {object} ErrorResponse
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
// @Param request body models.UpdateUserProfileRequest true "Update profile request"
// @Success 200 {object} models.UserProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
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
