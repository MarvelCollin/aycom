package service

import (
	"context"
	"fmt"
	"log"

	"aycom/backend/proto/user"

	"google.golang.org/grpc"
)

// UserClient defines the interface for user operations
type UserClient interface {
	GetUserById(ctx context.Context, userId string) (*UserInfo, error)
	UserExists(userId string) (bool, error)
	GetUserDetails(userId string) (map[string]interface{}, error)
}

// UserInfo represents user information returned by the user service
type UserInfo struct {
	Id                string
	Username          string
	DisplayName       string
	Email             string
	ProfilePictureUrl string
	Bio               string
	IsVerified        bool
}

// realUserClient implements the real UserClient interface using gRPC
type realUserClient struct {
	client user.UserServiceClient
}

// NewUserClient creates a new user client
func NewUserClient(conn *grpc.ClientConn) UserClient {
	if conn == nil {
		log.Println("ERROR: No connection provided to User Service client")
		return nil
	}

	log.Println("Creating real User Service client with gRPC connection")
	return &realUserClient{
		client: user.NewUserServiceClient(conn),
	}
}

// GetUserById retrieves user information by ID from the user service
func (c *realUserClient) GetUserById(ctx context.Context, userId string) (*UserInfo, error) {
	log.Printf("Fetching real user data for user ID: %s", userId)

	// Make the actual gRPC call to the user service
	response, err := c.client.GetUser(ctx, &user.GetUserRequest{
		UserId: userId,
	})

	if err != nil {
		log.Printf("Error calling user service: %v", err)
		return nil, fmt.Errorf("failed to get user data: %w", err)
	}

	if response == nil || response.User == nil {
		log.Printf("No user data returned for ID: %s", userId)
		return nil, fmt.Errorf("no user data found for ID: %s", userId)
	}

	// Map the response to UserInfo
	user := &UserInfo{
		Id:                response.User.Id,
		Username:          response.User.Username,
		DisplayName:       response.User.Name,
		Email:             response.User.Email,
		ProfilePictureUrl: response.User.ProfilePictureUrl,
		Bio:               response.User.Bio,
		IsVerified:        false,
	}

	log.Printf("Successfully retrieved real user data for %s (username: %s)", user.Id, user.Username)
	return user, nil
}

// UserExists checks if a user exists in the user service
func (c *realUserClient) UserExists(userId string) (bool, error) {
	log.Printf("Checking if user exists: %s", userId)

	if c.client == nil {
		log.Printf("Warning: User service client is nil")
		return true, nil // Assume user exists if we can't check
	}

	// Try to get the user - if successful, the user exists
	ctx := context.Background()
	_, err := c.GetUserById(ctx, userId)
	if err != nil {
		if err.Error() == fmt.Sprintf("no user data found for ID: %s", userId) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetUserDetails retrieves user details from the user service
func (c *realUserClient) GetUserDetails(userId string) (map[string]interface{}, error) {
	log.Printf("Getting user details for: %s", userId)

	if c.client == nil {
		log.Printf("Warning: User service client is nil")
		return map[string]interface{}{}, nil
	}

	ctx := context.Background()
	user, err := c.GetUserById(ctx, userId)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"id":                  user.Id,
		"username":            user.Username,
		"display_name":        user.DisplayName,
		"email":               user.Email,
		"profile_picture_url": user.ProfilePictureUrl,
		"bio":                 user.Bio,
		"is_verified":         user.IsVerified,
	}, nil
}
