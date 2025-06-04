package service

import (
	"context"
	"fmt"
	"log"

	"aycom/backend/proto/user"

	"google.golang.org/grpc"
)

type UserClient interface {
	GetUserById(ctx context.Context, userId string) (*UserInfo, error)
	UserExists(userId string) (bool, error)
	GetUserDetails(userId string) (map[string]interface{}, error)
}

type UserInfo struct {
	Id                string
	Username          string
	DisplayName       string
	Email             string
	ProfilePictureUrl string
	Bio               string
	IsVerified        bool
}

type realUserClient struct {
	client user.UserServiceClient
}

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

func (c *realUserClient) GetUserById(ctx context.Context, userId string) (*UserInfo, error) {
	log.Printf("Fetching real user data for user ID: %s", userId)

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

func (c *realUserClient) UserExists(userId string) (bool, error) {
	log.Printf("Checking if user exists: %s", userId)

	if c.client == nil {
		log.Printf("Warning: User service client is nil")
		return true, nil 
	}

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