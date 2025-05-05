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
