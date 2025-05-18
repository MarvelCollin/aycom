package handlers

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"aycom/backend/api-gateway/config"
	userProto "aycom/backend/proto/user"
)

// UserServiceClient provides methods to interact with the User service
type UserServiceClient interface {
	// User authentication & profile operations
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

	// Social operations
	FollowUser(followerID, followedID string) error
	UnfollowUser(followerID, followedID string) error
	GetFollowers(userID string, page, limit int) ([]*User, error)
	GetFollowing(userID string, page, limit int) ([]*User, error)

	// Search operations
	SearchUsers(query string, filter string, page, limit int) ([]*User, int, error)
	GetUserRecommendations(userID string, limit int) ([]*User, error)
	GetAllUsers(page, limit int, sortBy string, ascending bool) ([]*User, int, int, error)

	// Password reset operations
	RequestPasswordReset(email string) (bool, string, string, error)
	VerifySecurityAnswer(email, securityAnswer string) (bool, string, string, error)
	VerifyResetToken(token string) (bool, string, string, error)
	ResetPassword(token, newPassword, email string) (bool, string, error)
}

// User represents a user with profile information
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
	IsPrivate         bool
	FollowerCount     int
	FollowingCount    int
	IsFollowing       bool
	CreatedAt         time.Time
}

// UserProfileUpdate contains fields that can be updated in a user profile
type UserProfileUpdate struct {
	Name              string
	Bio               string
	Email             string
	DateOfBirth       string
	Gender            string
	ProfilePictureURL string
	BannerURL         string
}

// UserAuthResponse contains authentication response data
type UserAuthResponse struct {
	UserID       string
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         *User
}

// GRPCUserServiceClient is an implementation of UserServiceClient
// that communicates with the User service via gRPC
type GRPCUserServiceClient struct {
	client userProto.UserServiceClient
}

// Global instance of the user service client
var userServiceClient UserServiceClient

// InitUserServiceClient initializes the user service client
func InitUserServiceClient(cfg *config.Config) {
	if userServiceClient != nil {
		log.Println("User service client already initialized")
		return
	}

	userServiceAddr := cfg.Services.UserService
	log.Printf("Initializing User service client at %s", userServiceAddr)

	// Only attempt to initialize if we have a valid connection to the service
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

// Register implements UserServiceClient
func (c *GRPCUserServiceClient) Register(username, email, password, name string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a User object with the registration information
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

// Login implements UserServiceClient
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

	// Since we don't have tokens in the LoginUserResponse per proto,
	// we'll need to generate them at the API gateway level
	return &UserAuthResponse{
		UserID:       resp.User.Id,
		AccessToken:  "",   // Will need to be generated by the API gateway
		RefreshToken: "",   // Will need to be generated by the API gateway
		ExpiresIn:    3600, // 1 hour default
		User:         convertProtoToUser(resp.User),
	}, nil
}

// GetUserProfile implements UserServiceClient
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

// UpdateUserProfile implements UserServiceClient
func (c *GRPCUserServiceClient) UpdateUserProfile(userID string, profile *UserProfileUpdate) (*User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Only include fields that exist in the UpdateUserRequest proto message
	req := &userProto.UpdateUserRequest{
		UserId:            userID,
		Name:              profile.Name,
		Email:             profile.Email,
		ProfilePictureUrl: profile.ProfilePictureURL,
		BannerUrl:         profile.BannerURL,
	}

	// Create a user object with all the fields that need to be updated
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

// CheckUsernameAvailability implements UserServiceClient
func (c *GRPCUserServiceClient) CheckUsernameAvailability(username string) (bool, error) {
	// Since this method isn't directly available in the proto, we can implement it
	// by trying to get a user with the given username
	if c.client == nil {
		return false, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// We can use GetUserByEmail as a workaround, even though it's not ideal
	resp, err := c.client.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: username, // Using username in place of email to check availability
	})

	if err != nil {
		// If we get an error, it likely means the username doesn't exist, so it's available
		return true, nil
	}

	// If we get a response with a non-nil user, the username is taken
	return resp.User == nil, nil
}

// GetUserByEmail implements UserServiceClient
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

// GetUserById retrieves a user by ID
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

// GetUserByUsername retrieves a user by username
func (c *GRPCUserServiceClient) GetUserByUsername(username string) (*User, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use SearchUsers as a workaround since direct username lookup isn't implemented yet
	resp, err := c.client.SearchUsers(ctx, &userProto.SearchUsersRequest{
		Query: username,
		Limit: 1,
	})
	if err != nil {
		log.Printf("Failed to search for user by username '%s': %v", username, err)
		return nil, fmt.Errorf("failed to find user with username: %s", username)
	}

	// Check if we found any users
	if len(resp.Users) == 0 {
		return nil, fmt.Errorf("user with username '%s' not found", username)
	}

	// Find the exact username match
	for _, user := range resp.Users {
		if strings.EqualFold(user.Username, username) {
			return convertProtoToUser(user), nil
		}
	}

	// If no exact match, return the first user as fallback
	return convertProtoToUser(resp.Users[0]), nil
}

// IsUserBlocked checks if a user is blocked by another user
func (c *GRPCUserServiceClient) IsUserBlocked(userId, blockedById string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("user service client not initialized")
	}

	// If either ID is empty, return false (not blocked)
	if userId == "" || blockedById == "" {
		return false, nil
	}

	// Since this method isn't implemented in the proto yet, return false
	log.Printf("IsUserBlocked check not implemented in proto, returning false for userId=%s, blockedById=%s", userId, blockedById)
	return false, nil
}

// IsFollowing checks if a user is following another user
func (c *GRPCUserServiceClient) IsFollowing(followerId, followeeId string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("user service client not initialized")
	}

	// If either ID is empty, return false (not following)
	if followerId == "" || followeeId == "" {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get following list as workaround
	resp, err := c.client.GetFollowing(ctx, &userProto.GetFollowingRequest{
		UserId: followerId,
		Limit:  100, // Get a reasonable number to check
		Page:   1,
	})
	if err != nil {
		log.Printf("Error checking if user %s is following %s: %v", followerId, followeeId, err)
		return false, nil
	}

	// Check if followeeId is in the list
	for _, user := range resp.Following {
		if user.Id == followeeId {
			return true, nil
		}
	}

	return false, nil
}

// SearchUsers searches for users based on query
func (c *GRPCUserServiceClient) SearchUsers(query string, filter string, page int, limit int) ([]*User, int, error) {
	if c.client == nil {
		return nil, 0, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create the request
	req := &userProto.SearchUsersRequest{
		Query:  query,
		Filter: filter,
		Page:   int32(page),
		Limit:  int32(limit),
	}

	// Call the gRPC service
	resp, err := c.client.SearchUsers(ctx, req)
	if err != nil {
		log.Printf("Error calling SearchUsers gRPC: %v", err)
		return nil, 0, err
	}

	// Convert proto users to our User type
	users := make([]*User, 0, len(resp.GetUsers()))
	for _, protoUser := range resp.GetUsers() {
		user := convertProtoToUser(protoUser)
		if user != nil {
			users = append(users, user)
		}
	}

	// Get the total count from the response
	totalCount := int(resp.GetTotalCount())

	return users, totalCount, nil
}

// GetUserRecommendations implements UserServiceClient
func (c *GRPCUserServiceClient) GetUserRecommendations(userID string, limit int) ([]*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &userProto.GetRecommendedUsersRequest{
		UserId: userID,
		Limit:  int32(limit),
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

// FollowUser creates a follow relationship between two users
// TODO: Replace with real gRPC implementation once proto files are regenerated
func (c *GRPCUserServiceClient) FollowUser(followerID string, followedID string) error {
	// This is a temporary mock implementation
	log.Printf("[MOCK] User %s follows user %s", followerID, followedID)

	// In a real implementation, this would call the gRPC service
	// Return success for now - we'll properly implement this after proto regeneration
	return nil
}

// UnfollowUser removes a follow relationship between two users
// TODO: Replace with real gRPC implementation once proto files are regenerated
func (c *GRPCUserServiceClient) UnfollowUser(followerID string, followedID string) error {
	// This is a temporary mock implementation
	log.Printf("[MOCK] User %s unfollows user %s", followerID, followedID)

	// In a real implementation, this would call the gRPC service
	// Return success for now - we'll properly implement this after proto regeneration
	return nil
}

// GetFollowers gets the followers of a user
// TODO: Replace with real gRPC implementation once proto files are regenerated
func (c *GRPCUserServiceClient) GetFollowers(userID string, page int, limit int) ([]*User, error) {
	// This is a temporary mock implementation
	log.Printf("[MOCK] Getting followers for user %s, page %d, limit %d", userID, page, limit)

	// Return mock data
	mockUsers := []*User{
		{
			ID:                "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			Username:          "follower1",
			Name:              "Follower One",
			DisplayName:       "Follower One",
			ProfilePictureURL: "https://example.com/avatar1.jpg",
			IsVerified:        true,
			IsPrivate:         false,
			Email:             "follower1@example.com",
			FollowerCount:     10,
			FollowingCount:    20,
			IsFollowing:       false,
			CreatedAt:         time.Now(),
		},
		{
			ID:                "f47ac10b-58cc-4372-a567-0e02b2c3d480",
			Username:          "follower2",
			Name:              "Follower Two",
			DisplayName:       "Follower Two",
			ProfilePictureURL: "https://example.com/avatar2.jpg",
			IsVerified:        false,
			IsPrivate:         true,
			Email:             "follower2@example.com",
			FollowerCount:     5,
			FollowingCount:    15,
			IsFollowing:       true,
			CreatedAt:         time.Now(),
		},
	}

	return mockUsers, nil
}

// GetFollowing gets the users followed by a user
// TODO: Replace with real gRPC implementation once proto files are regenerated
func (c *GRPCUserServiceClient) GetFollowing(userID string, page int, limit int) ([]*User, error) {
	// This is a temporary mock implementation
	log.Printf("[MOCK] Getting following for user %s, page %d, limit %d", userID, page, limit)

	// Return mock data
	mockUsers := []*User{
		{
			ID:                "f47ac10b-58cc-4372-a567-0e02b2c3d481",
			Username:          "following1",
			Name:              "Following One",
			DisplayName:       "Following One",
			ProfilePictureURL: "https://example.com/avatar3.jpg",
			IsVerified:        true,
			IsPrivate:         false,
			Email:             "following1@example.com",
			FollowerCount:     25,
			FollowingCount:    5,
			IsFollowing:       true,
			CreatedAt:         time.Now(),
		},
		{
			ID:                "f47ac10b-58cc-4372-a567-0e02b2c3d482",
			Username:          "following2",
			Name:              "Following Two",
			DisplayName:       "Following Two",
			ProfilePictureURL: "https://example.com/avatar4.jpg",
			IsVerified:        false,
			IsPrivate:         true,
			Email:             "following2@example.com",
			FollowerCount:     30,
			FollowingCount:    3,
			IsFollowing:       true,
			CreatedAt:         time.Now(),
		},
	}

	return mockUsers, nil
}

// Helper to convert proto User to internal User type
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

	// Create user with the fields we know exist in the proto
	result := &User{
		ID:                u.Id,
		Username:          u.Username,
		Email:             u.Email,
		Name:              u.Name,
		DisplayName:       u.Name, // If display_name is empty, fallback to name
		ProfilePictureURL: u.ProfilePictureUrl,
		BannerURL:         u.BannerUrl,
		Bio:               u.Bio,
		IsVerified:        u.IsVerified,
		IsPrivate:         false, // Default to false since it's not in proto yet
		FollowerCount:     int(u.FollowerCount),
		FollowingCount:    int(u.FollowingCount),
		IsFollowing:       u.IsFollowing,
		CreatedAt:         createdAt,
	}

	return result
}

// GetAllUsers gets a paginated list of all users
func (c *GRPCUserServiceClient) GetAllUsers(page, limit int, sortBy string, ascending bool) ([]*User, int, int, error) {
	if c.client == nil {
		return nil, 0, 0, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create the request with the fields as defined in the proto
	req := &userProto.GetAllUsersRequest{
		Page:   int32(page),
		Limit:  int32(limit),
		SortBy: sortBy,
	}
	// Manually set the field using reflection
	reflect.ValueOf(req).Elem().FieldByName("SortDesc").SetBool(!ascending)

	// Call the gRPC service
	resp, err := c.client.GetAllUsers(ctx, req)
	if err != nil {
		log.Printf("Error calling GetAllUsers gRPC: %v", err)
		return nil, 0, 0, err
	}

	// Convert proto users to our User type
	users := make([]*User, 0, len(resp.GetUsers()))
	for _, protoUser := range resp.GetUsers() {
		user := convertProtoToUser(protoUser)
		if user != nil {
			users = append(users, user)
		}
	}

	// Calculate total pages if needed
	totalPages := 1
	if resp.GetTotalCount() > 0 && limit > 0 {
		totalPages = int((resp.GetTotalCount() + int32(limit) - 1) / int32(limit))
	}

	// Return users, total count, and total pages
	return users, int(resp.GetTotalCount()), totalPages, nil
}

// Implementation for the password reset methods
func (c *GRPCUserServiceClient) RequestPasswordReset(email string) (bool, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the actual gRPC method
	resp, err := c.client.RequestPasswordReset(ctx, &userProto.RequestPasswordResetRequest{
		Email: email,
	})

	if err != nil {
		return false, "", "", err
	}

	// We need to use the actual field names from the generated Go code
	// Looking at the proto definition, it has a 'token' field
	return resp.GetSuccess(), resp.GetMessage(), "security_question", nil // Returning a placeholder since token isn't accessible
}

func (c *GRPCUserServiceClient) VerifySecurityAnswer(email, securityAnswer string) (bool, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the actual gRPC method
	resp, err := c.client.VerifySecurityAnswer(ctx, &userProto.VerifySecurityAnswerRequest{
		Email:          email,
		SecurityAnswer: securityAnswer,
	})

	if err != nil {
		return false, "", "", err
	}

	// Return using the getters, not direct field access
	return resp.GetSuccess(), resp.GetMessage(), "token", nil // Returning a placeholder since token isn't accessible
}

func (c *GRPCUserServiceClient) VerifyResetToken(token string) (bool, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the actual gRPC method
	resp, err := c.client.VerifyResetToken(ctx, &userProto.VerifyResetTokenRequest{
		Token: token,
	})

	if err != nil {
		return false, "", "", err
	}

	return resp.Valid, resp.Email, resp.Message, nil
}

func (c *GRPCUserServiceClient) ResetPassword(token, newPassword, email string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the actual gRPC method
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
