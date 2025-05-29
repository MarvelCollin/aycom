package handlers

import (
	userProto "aycom/backend/proto/user"
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"aycom/backend/api-gateway/config"
)

type UserServiceClient interface {
	Register(username, email, password, name string) (string, error)
	Login(emailOrUsername, password string) (*UserAuthResponse, error)
	GetUserProfile(userID string) (*User, error)
	UpdateUserProfile(userID string, profile *UserProfileUpdate) (*User, error)
	CheckUsernameAvailability(username string) (bool, error)
	GetUserByEmail(email string) (*User, error)
	GetUserById(userId string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	IsUserBlocked(userId, blockedById string) (bool, error)
	IsFollowing(followerId, followeeId string) (bool, error)

	FollowUser(followerID, followedID string) error
	UnfollowUser(followerID, followedID string) error
	GetFollowers(userID string, page, limit int) ([]*User, error)
	GetFollowing(userID string, page, limit int) ([]*User, error)

	SearchUsers(query string, filter string, page, limit int) ([]*User, int, error)
	GetUserRecommendations(userID string, limit int) ([]*User, error)
	GetAllUsers(page, limit int, sortBy string, ascending bool) ([]*User, int, int, error)

	RequestPasswordReset(email string) (bool, string, string, error)
	VerifySecurityAnswer(email, securityAnswer string) (bool, string, string, error)
	VerifyResetToken(token string) (bool, string, string, error)
	ResetPassword(token, newPassword, email string) (bool, string, error)

	BlockUser(blockerID, blockedID string) error
	UnblockUser(blockerID, blockedID string) error
	GetBlockedUsers(userID string, page, limit int) ([]*User, error)
	ReportUser(reporterID, reportedID, reason string) error
}

type User struct {
	ID                string
	Username          string
	Email             string
	Name              string
	DisplayName       string
	ProfilePictureURL string
	BannerURL         string
	Bio               string
	IsVerified        bool
	IsAdmin           bool
	IsPrivate         bool
	FollowerCount     int
	FollowingCount    int
	IsFollowing       bool
	CreatedAt         time.Time
}

type UserProfileUpdate struct {
	Name              string
	Bio               string
	Email             string
	DateOfBirth       string
	Gender            string
	ProfilePictureURL string
	BannerURL         string
}

type UserAuthResponse struct {
	UserID       string
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         *User
}

type GRPCUserServiceClient struct {
	client userProto.UserServiceClient
}

var userServiceClient UserServiceClient

func InitUserServiceClient(cfg *config.Config) {
	if userServiceClient != nil {
		log.Println("User service client already initialized")
		return
	}

	userServiceAddr := cfg.Services.UserService
	log.Printf("Initializing User service client at %s", userServiceAddr)

	if UserClient != nil {
		log.Println("Creating user service client wrapper using existing gRPC connection")
		userServiceClient = &GRPCUserServiceClient{
			client: UserClient,
		}
		log.Println("User service client initialized successfully")
	} else {
		log.Println("WARNING: Cannot initialize user service client - no connection to User service")
	}
}

func (c *GRPCUserServiceClient) Register(username, email, password, name string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := &userProto.User{
		Username: username,
		Email:    email,
		Password: password,
		Name:     name,
	}

	resp, err := c.client.CreateUser(ctx, &userProto.CreateUserRequest{
		User: user,
	})
	if err != nil {
		return "", err
	}

	return resp.User.Id, nil
}

func (c *GRPCUserServiceClient) Login(emailOrUsername, password string) (*UserAuthResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.LoginUser(ctx, &userProto.LoginUserRequest{
		Email:    emailOrUsername,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	return &UserAuthResponse{
		UserID:       resp.User.Id,
		AccessToken:  "",
		RefreshToken: "",
		ExpiresIn:    3600,
		User:         convertProtoToUser(resp.User),
	}, nil
}

func (c *GRPCUserServiceClient) GetUserProfile(userID string) (*User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.GetUser(ctx, &userProto.GetUserRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	return convertProtoToUser(resp.User), nil
}

func (c *GRPCUserServiceClient) UpdateUserProfile(userID string, profile *UserProfileUpdate) (*User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &userProto.UpdateUserRequest{
		UserId:            userID,
		Name:              profile.Name,
		Email:             profile.Email,
		ProfilePictureUrl: profile.ProfilePictureURL,
		BannerUrl:         profile.BannerURL,
	}

	req.User = &userProto.User{
		Bio:         profile.Bio,
		Gender:      profile.Gender,
		DateOfBirth: profile.DateOfBirth,
	}

	resp, err := c.client.UpdateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return convertProtoToUser(resp.User), nil
}

func (c *GRPCUserServiceClient) CheckUsernameAvailability(username string) (bool, error) {

	if c.client == nil {
		return false, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: username,
	})

	if err != nil {

		return true, nil
	}

	return resp.User == nil, nil
}

func (c *GRPCUserServiceClient) GetUserByEmail(email string) (*User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: email,
	})
	if err != nil {
		return nil, err
	}

	return convertProtoToUser(resp.User), nil
}

func (c *GRPCUserServiceClient) GetUserById(userId string) (*User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.GetUser(ctx, &userProto.GetUserRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	return convertProtoToUser(resp.User), nil
}

func (c *GRPCUserServiceClient) GetUserByUsername(username string) (*User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.SearchUsers(ctx, &userProto.SearchUsersRequest{
		Query: username,
		Limit: 1,
	})
	if err != nil {
		log.Printf("Failed to search for user by username '%s': %v", username, err)
		return nil, fmt.Errorf("failed to find user with username: %s", username)
	}

	if len(resp.Users) == 0 {
		return nil, fmt.Errorf("user with username '%s' not found", username)
	}

	for _, user := range resp.Users {
		if strings.EqualFold(user.Username, username) {
			return convertProtoToUser(user), nil
		}
	}

	return convertProtoToUser(resp.Users[0]), nil
}

func (c *GRPCUserServiceClient) IsUserBlocked(userId, blockedById string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("user service client not initialized")
	}

	if userId == "" || blockedById == "" {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &userProto.IsUserBlockedRequest{
		UserId:      userId,
		BlockedById: blockedById,
	}

	resp, err := c.client.IsUserBlocked(ctx, req)
	if err != nil {
		log.Printf("Error checking if user %s is blocked by %s: %v", userId, blockedById, err)
		return false, nil
	}

	return resp.IsBlocked, nil
}

func (c *GRPCUserServiceClient) IsFollowing(followerId, followeeId string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("user service client not initialized")
	}

	if followerId == "" || followeeId == "" {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use the proper IsFollowing gRPC method instead of GetFollowing
	resp, err := c.client.IsFollowing(ctx, &userProto.IsFollowingRequest{
		FollowerId: followerId,
		FollowedId: followeeId,
	})
	if err != nil {
		log.Printf("Error checking if user %s is following %s: %v", followerId, followeeId, err)
		return false, nil
	}

	return resp.IsFollowing, nil
}

func (c *GRPCUserServiceClient) SearchUsers(query string, filter string, page int, limit int) ([]*User, int, error) {
	if c.client == nil {
		return nil, 0, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &userProto.SearchUsersRequest{
		Query: query,
		Page:  int32(page),
		Limit: int32(limit),
	}

	resp, err := c.client.SearchUsers(ctx, req)
	if err != nil {
		log.Printf("Error calling SearchUsers gRPC: %v", err)
		return nil, 0, err
	}

	users := make([]*User, 0, len(resp.GetUsers()))
	for _, protoUser := range resp.GetUsers() {
		user := convertProtoToUser(protoUser)
		if user != nil {
			users = append(users, user)
		}
	}

	totalCount := int(resp.GetTotalCount())

	return users, totalCount, nil
}

func (c *GRPCUserServiceClient) GetUserRecommendations(userID string, limit int) ([]*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &userProto.GetRecommendedUsersRequest{
		Limit: int32(limit),
	}

	resp, err := c.client.GetRecommendedUsers(ctx, req)
	if err != nil {
		return nil, err
	}

	var users []*User
	for _, u := range resp.GetUsers() {
		user := convertProtoToUser(u)
		if user != nil {
			users = append(users, user)
		}
	}

	return users, nil
}

func (c *GRPCUserServiceClient) FollowUser(followerID string, followedID string) error {
	if c.client == nil {
		return fmt.Errorf("user service client not initialized")
	}

	log.Printf("User %s following user %s", followerID, followedID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &userProto.FollowUserRequest{
		FollowerId: followerID,
		FollowedId: followedID,
	}

	resp, err := c.client.FollowUser(ctx, req)
	if err != nil {
		log.Printf("Error in FollowUser gRPC call: %v", err)
		return err
	}

	log.Printf("Follow response: success=%v, message=%s", resp.Success, resp.Message)
	return nil
}

func (c *GRPCUserServiceClient) UnfollowUser(followerID string, followedID string) error {
	if c.client == nil {
		return fmt.Errorf("user service client not initialized")
	}

	log.Printf("User %s unfollowing user %s", followerID, followedID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &userProto.UnfollowUserRequest{
		FollowerId: followerID,
		FollowedId: followedID,
	}

	resp, err := c.client.UnfollowUser(ctx, req)
	if err != nil {
		log.Printf("Error in UnfollowUser gRPC call: %v", err)
		return err
	}

	log.Printf("Unfollow response: success=%v, message=%s", resp.Success, resp.Message)
	return nil
}

func (c *GRPCUserServiceClient) GetFollowers(userID string, page int, limit int) ([]*User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	log.Printf("Getting followers for user %s, page %d, limit %d", userID, page, limit)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &userProto.GetFollowersRequest{
		UserId: userID,
		Page:   int32(page),
		Limit:  int32(limit),
	}

	resp, err := c.client.GetFollowers(ctx, req)
	if err != nil {
		log.Printf("Error in GetFollowers gRPC call: %v", err)
		return nil, err
	}

	followers := make([]*User, 0, len(resp.GetFollowers()))
	for _, protoUser := range resp.GetFollowers() {
		follower := convertProtoToUser(protoUser)
		if follower != nil {
			followers = append(followers, follower)
		}
	}

	return followers, nil
}

func (c *GRPCUserServiceClient) GetFollowing(userID string, page int, limit int) ([]*User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	log.Printf("Getting following for user %s, page %d, limit %d", userID, page, limit)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &userProto.GetFollowingRequest{
		UserId: userID,
		Page:   int32(page),
		Limit:  int32(limit),
	}

	resp, err := c.client.GetFollowing(ctx, req)
	if err != nil {
		log.Printf("Error in GetFollowing gRPC call: %v", err)
		return nil, err
	}

	following := make([]*User, 0, len(resp.GetFollowing()))
	for _, protoUser := range resp.GetFollowing() {
		user := convertProtoToUser(protoUser)
		if user != nil {
			user.IsFollowing = true // We're getting users we already follow
			following = append(following, user)
		}
	}

	return following, nil
}

func convertProtoToUser(u *userProto.User) *User {
	if u == nil {
		return nil
	}

	createdAt := time.Time{}
	if u.CreatedAt != "" {
		parsed, err := time.Parse(time.RFC3339, u.CreatedAt)
		if err == nil {
			createdAt = parsed
		}
	}

	result := &User{
		ID:                u.Id,
		Username:          u.Username,
		Email:             u.Email,
		Name:              u.Name,
		DisplayName:       u.Name,
		ProfilePictureURL: u.ProfilePictureUrl,
		BannerURL:         u.BannerUrl,
		Bio:               u.Bio,
		IsVerified:        u.IsVerified,
		IsAdmin:           u.IsAdmin,
		IsPrivate:         false,
		FollowerCount:     int(u.FollowerCount),
		FollowingCount:    int(u.FollowingCount),
		IsFollowing:       u.IsFollowing,
		CreatedAt:         createdAt,
	}

	return result
}

func (c *GRPCUserServiceClient) GetAllUsers(page, limit int, sortBy string, ascending bool) ([]*User, int, int, error) {
	if c.client == nil {
		return nil, 0, 0, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &userProto.GetAllUsersRequest{
		Page:   int32(page),
		Limit:  int32(limit),
		SortBy: sortBy,
	}

	reflect.ValueOf(req).Elem().FieldByName("SortDesc").SetBool(!ascending)

	resp, err := c.client.GetAllUsers(ctx, req)
	if err != nil {
		log.Printf("Error calling GetAllUsers gRPC: %v", err)
		return nil, 0, 0, err
	}

	users := make([]*User, 0, len(resp.GetUsers()))
	for _, protoUser := range resp.GetUsers() {
		user := convertProtoToUser(protoUser)
		if user != nil {
			users = append(users, user)
		}
	}

	totalPages := 1
	if resp.GetTotalCount() > 0 && limit > 0 {
		totalPages = int((resp.GetTotalCount() + int32(limit) - 1) / int32(limit))
	}

	return users, int(resp.GetTotalCount()), totalPages, nil
}

func (c *GRPCUserServiceClient) RequestPasswordReset(email string) (bool, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.RequestPasswordReset(ctx, &userProto.RequestPasswordResetRequest{
		Email: email,
	})

	if err != nil {
		return false, "", "", err
	}

	return resp.GetSuccess(), resp.GetMessage(), "security_question", nil
}

func (c *GRPCUserServiceClient) VerifySecurityAnswer(email, securityAnswer string) (bool, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.VerifySecurityAnswer(ctx, &userProto.VerifySecurityAnswerRequest{
		Email:  email,
		Answer: securityAnswer,
	})

	if err != nil {
		return false, "", "", err
	}

	return resp.Valid, resp.Message, resp.Token, nil
}

func (c *GRPCUserServiceClient) VerifyResetToken(token string) (bool, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.VerifyResetToken(ctx, &userProto.VerifyResetTokenRequest{
		Token: token,
	})

	if err != nil {
		return false, "", "", err
	}

	return resp.Valid, resp.UserId, resp.Message, nil
}

func (c *GRPCUserServiceClient) ResetPassword(token, newPassword, email string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.ResetPassword(ctx, &userProto.ResetPasswordRequest{
		Token:       token,
		NewPassword: newPassword,
		Email:       email,
	})

	if err != nil {
		return false, "", err
	}

	return resp.Success, resp.Message, nil
}

func (c *GRPCUserServiceClient) BlockUser(blockerID, blockedID string) error {
	if c.client == nil {
		return fmt.Errorf("user service client not initialized")
	}

	if blockerID == "" || blockedID == "" {
		return fmt.Errorf("both blocker ID and blocked ID are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &userProto.BlockUserRequest{
		UserId:      blockedID,
		BlockedById: blockerID,
	}

	_, err := c.client.BlockUser(ctx, req)
	if err != nil {
		log.Printf("Error blocking user %s by user %s: %v", blockedID, blockerID, err)
		return err
	}

	return nil
}

func (c *GRPCUserServiceClient) UnblockUser(blockerID, blockedID string) error {
	if c.client == nil {
		return fmt.Errorf("user service client not initialized")
	}

	if blockerID == "" || blockedID == "" {
		return fmt.Errorf("both blocker ID and blocked ID are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &userProto.UnblockUserRequest{
		UserId:        blockedID,
		UnblockedById: blockerID,
	}

	_, err := c.client.UnblockUser(ctx, req)
	if err != nil {
		log.Printf("Error unblocking user %s by user %s: %v", blockedID, blockerID, err)
		return err
	}

	return nil
}

func (c *GRPCUserServiceClient) GetBlockedUsers(userID string, page, limit int) ([]*User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Since GetBlockedUsers is not defined in the proto file, we implement a mock
	log.Printf("Getting blocked users for user %s (page: %d, limit: %d)", userID, page, limit)

	// Return mock data for now
	return []*User{}, nil
}

func (c *GRPCUserServiceClient) ReportUser(reporterID, reportedID, reason string) error {
	if c.client == nil {
		return fmt.Errorf("user service client not initialized")
	}

	if reporterID == "" || reportedID == "" {
		return fmt.Errorf("both reporter ID and reported ID are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &userProto.ReportUserRequest{
		UserId:       reportedID,
		ReportedById: reporterID,
		Reason:       reason,
	}

	_, err := c.client.ReportUser(ctx, req)
	if err != nil {
		log.Printf("Error reporting user %s by user %s: %v", reportedID, reporterID, err)
		return err
	}

	return nil
}
