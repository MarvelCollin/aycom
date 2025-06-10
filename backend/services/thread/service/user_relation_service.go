package service

import (
	"context"
	"log"

	"aycom/backend/proto/user"
)

// userRelationService implements UserRelationService
type userRelationService struct {
	userClient     UserClient
	userGrpcClient user.UserServiceClient
}

// NewUserRelationService creates a new UserRelationService
func NewUserRelationService(userClient UserClient) UserRelationService {
	var userGrpcClient user.UserServiceClient

	// Get the gRPC client from the UserClient interface
	if userClient != nil {
		userGrpcClient = userClient.GetGrpcClient()
	}

	return &userRelationService{
		userClient:     userClient,
		userGrpcClient: userGrpcClient,
	}
}

// CheckUserFollows checks if follower follows followed
func (s *userRelationService) CheckUserFollows(ctx context.Context, followerID, followedID string) (bool, error) {
	if s.userGrpcClient == nil {
		log.Println("User gRPC client is not initialized")
		// Default to allowing the reply if we can't check
		return true, nil
	}

	// Direct call to user service gRPC method
	resp, err := s.userGrpcClient.IsFollowing(ctx, &user.IsFollowingRequest{
		FollowerId: followerID,
		FollowedId: followedID,
	})

	if err != nil {
		log.Printf("Failed to check follow relationship: %v", err)
		// Default to allowing the reply if we can't check
		return true, nil
	}

	return resp.IsFollowing, nil
}

// CheckUserVerified checks if a user is verified
func (s *userRelationService) CheckUserVerified(ctx context.Context, userID string) (bool, error) {
	if s.userClient == nil {
		log.Println("User client is not initialized")
		// Default to allowing the reply if we can't check
		return true, nil
	}

	userInfo, err := s.userClient.GetUserById(ctx, userID)
	if err != nil {
		log.Printf("Failed to get user information: %v", err)
		// Default to allowing the reply if we can't check
		return true, nil
	}

	return userInfo.IsVerified, nil
}
