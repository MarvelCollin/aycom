package utils

import (
	userProto "aycom/backend/proto/user"
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ResolveUserIdentifier attempts to resolve a user identifier (which could be a UUID or username)
// to a valid user ID.
func ResolveUserIdentifier(ctx context.Context, userClient userProto.UserServiceClient, identifier string) (string, error) {
	// First check if it's a valid UUID
	if _, err := uuid.Parse(identifier); err == nil {
		// It's a valid UUID, verify the user exists
		_, err := userClient.GetUser(ctx, &userProto.GetUserRequest{
			UserId: identifier,
		})
		if err != nil {
			return "", fmt.Errorf("user with ID %s not found: %v", identifier, err)
		}
		return identifier, nil
	}

	// Try to resolve as username
	resp, err := userClient.GetUserByUsername(ctx, &userProto.GetUserByUsernameRequest{
		Username: identifier,
	})
	if err != nil {
		return "", fmt.Errorf("user with username %s not found: %v", identifier, err)
	}
	if resp.User == nil {
		return "", fmt.Errorf("user with username %s not found", identifier)
	}

	return resp.User.Id, nil
}

// CheckFollowStatus checks if one user is following another
func CheckFollowStatus(ctx context.Context, userClient userProto.UserServiceClient, followerID, followedID string) (bool, error) {
	isFollowingReq := &userProto.IsFollowingRequest{
		FollowerId: followerID,
		FollowedId: followedID,
	}

	isFollowingResp, err := userClient.IsFollowing(ctx, isFollowingReq)
	if err != nil {
		return false, err
	}

	return isFollowingResp.IsFollowing, nil
}

// EnrichUsersWithFollowStatus adds the is_following field to a list of users
func EnrichUsersWithFollowStatus(ctx context.Context, userClient userProto.UserServiceClient, currentUserID string, users []gin.H) error {
	for i, user := range users {
		userID, ok := user["id"].(string)
		if !ok {
			continue
		}

		isFollowing, err := CheckFollowStatus(ctx, userClient, currentUserID, userID)
		if err != nil {
			log.Printf("Error checking follow status for user %s: %v", userID, err)
			continue
		}

		users[i]["is_following"] = isFollowing
	}
	return nil
}
