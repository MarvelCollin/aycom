package service

import (
	"aycom/backend/proto/user"
	"context"
	"log"
)

type userRelationService struct {
	userClient     UserClient
	userGrpcClient user.UserServiceClient
}

func NewUserRelationService(userClient UserClient) UserRelationService {
	var userGrpcClient user.UserServiceClient

	if userClient != nil {
		userGrpcClient = userClient.GetGrpcClient()
	}

	return &userRelationService{
		userClient:     userClient,
		userGrpcClient: userGrpcClient,
	}
}

func (s *userRelationService) CheckUserFollows(ctx context.Context, followerID, followedID string) (bool, error) {
	if s.userGrpcClient == nil {
		log.Println("User gRPC client is not initialized")

		return true, nil
	}

	resp, err := s.userGrpcClient.IsFollowing(ctx, &user.IsFollowingRequest{
		FollowerId: followerID,
		FollowedId: followedID,
	})

	if err != nil {
		log.Printf("Failed to check follow relationship: %v", err)

		return true, nil
	}

	return resp.IsFollowing, nil
}

func (s *userRelationService) CheckUserVerified(ctx context.Context, userID string) (bool, error) {
	if s.userClient == nil {
		log.Println("User client is not initialized")

		return true, nil
	}

	userInfo, err := s.userClient.GetUserById(ctx, userID)
	if err != nil {
		log.Printf("Failed to get user information: %v", err)

		return true, nil
	}

	return userInfo.IsVerified, nil
}
