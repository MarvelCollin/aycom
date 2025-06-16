package handlers

import (
	userProto "aycom/backend/proto/user"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"aycom/backend/api-gateway/config"
	"aycom/backend/api-gateway/utils"
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

	SearchUsers(query string, filter string, page, limit int, enableFuzzy bool) ([]*User, int, error)
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

	UserExists(userID string) (bool, error)
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
	IsPrivate         bool
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

var (
	userServiceClient UserServiceClient
	healthCheckMutex  sync.Mutex
	healthCheckActive bool
)

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

		startBackgroundHealthCheck()
	} else {
		log.Println("WARNING: Cannot initialize user service client - no connection to User service")

		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			if i > 0 {
				log.Printf("Retry attempt %d/%d to connect to User service", i+1, maxRetries)
				time.Sleep(time.Duration(i+1) * time.Second)
			}

			conn, err := grpc.Dial(userServiceAddr,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)

			if err == nil {
				log.Println("Successfully connected to User service")
				UserClient = userProto.NewUserServiceClient(conn)
				userServiceClient = &GRPCUserServiceClient{
					client: UserClient,
				}

				startBackgroundHealthCheck()
				return
			}

			log.Printf("Failed to connect to User service on attempt %d: %v", i+1, err)
		}

		log.Println("Failed to establish connection with User service after multiple attempts")
	}
}

func startBackgroundHealthCheck() {
	healthCheckMutex.Lock()
	defer healthCheckMutex.Unlock()

	if healthCheckActive {
		log.Println("User service health check already active")
		return
	}

	healthCheckActive = true
	log.Println("Starting background health check for User service")

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			checkUserServiceHealth()
		}
	}()
}

func checkUserServiceHealth() {
	if UserClient == nil {
		log.Println("Health check: UserClient is nil, attempting to reconnect")
		InitGRPCServices()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := UserClient.GetUser(ctx, &userProto.GetUserRequest{
		UserId: "health-check",
	})

	if err != nil {
		log.Printf("Health check failed: %v - attempting to reconnect", err)
		InitGRPCServices()
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
		IsPrivate:         profile.IsPrivate,
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

func (c *GRPCUserServiceClient) SearchUsers(query string, filter string, page int, limit int, enableFuzzy bool) ([]*User, int, error) {
	if c.client == nil {
		return nil, 0, fmt.Errorf("user service client not initialized")
	}

	log.Printf("SearchUsers called with query=%s, filter=%s, page=%d, limit=%d, enableFuzzy=%t", query, filter, page, limit, enableFuzzy)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if filter == "" {
		filter = "all"
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "filter", filter)

	dbQuery := query

	fetchLimit := 100
	if query != "" && len(query) <= 4 {
		fetchLimit = 200
	}

	// For fuzzy search, we may need to do multiple attempts to get enough candidates
	var resp *userProto.SearchUsersResponse
	var err error

	// First attempt: search with original query
	req := &userProto.SearchUsersRequest{
		Query: dbQuery,
		Page:  1,
		Limit: int32(fetchLimit),
	}

	resp, err = c.client.SearchUsers(ctx, req)
	if err != nil {
		log.Printf("Error calling SearchUsers gRPC: %v", err)
		return nil, 0, err
	}

	log.Printf("User service returned %d users for query '%s'", len(resp.GetUsers()), query)

	// If fuzzy search is enabled and we got few/no results, try broader searches
	if enableFuzzy && query != "" && len(resp.GetUsers()) < 10 {
		log.Printf("Fuzzy search enabled with few results (%d), attempting broader search", len(resp.GetUsers()))

		// Try with substrings of the query
		broadQueries := []string{}
		if len(query) > 2 {
			// Try with first 3 characters
			broadQueries = append(broadQueries, query[:3])
		}
		if len(query) > 3 {
			// Try with first 4 characters
			broadQueries = append(broadQueries, query[:4])
		}

		for _, broadQuery := range broadQueries {
			log.Printf("Trying broader search with query '%s'", broadQuery)
			broadReq := &userProto.SearchUsersRequest{
				Query: broadQuery,
				Page:  1,
				Limit: int32(fetchLimit),
			}

			broadResp, broadErr := c.client.SearchUsers(ctx, broadReq)
			if broadErr != nil {
				log.Printf("Error in broader search with query '%s': %v", broadQuery, broadErr)
				continue
			}

			log.Printf("Broader search with '%s' returned %d users", broadQuery, len(broadResp.GetUsers()))

			// Merge results, avoiding duplicates
			existingUsers := make(map[string]bool)
			for _, user := range resp.GetUsers() {
				existingUsers[user.GetId()] = true
			}

			for _, user := range broadResp.GetUsers() {
				if !existingUsers[user.GetId()] {
					resp.Users = append(resp.Users, user)
					existingUsers[user.GetId()] = true
				}
			}

			// If we have enough candidates now, break
			if len(resp.GetUsers()) >= 20 {
				break
			}
		}

		log.Printf("After broader searches, have %d total users for fuzzy matching", len(resp.GetUsers()))
	}

	allUsers := make([]*User, 0)

	if query == "" {
		for _, protoUser := range resp.GetUsers() {
			user := convertProtoToUser(protoUser)
			if user != nil && (filter != "verified" || user.IsVerified) {
				allUsers = append(allUsers, user)
			}
		}
	} else {

		queryLower := strings.ToLower(query)

		var similarityThreshold float64
		switch {
		case len(query) <= 2:
			similarityThreshold = 0.8
		case len(query) <= 3:
			similarityThreshold = 0.7
		case len(query) <= 5:
			similarityThreshold = 0.5
		default:
			similarityThreshold = 0.4
		}

		requireSubstring := len(query) <= 2

		for _, protoUser := range resp.GetUsers() {
			user := convertProtoToUser(protoUser)
			if user == nil || (filter == "verified" && !user.IsVerified) {
				continue
			}

			usernameScore := utils.DamerauLevenshteinSimilarity(strings.ToLower(user.Username), queryLower)
			usernameMatch := usernameScore >= similarityThreshold

			usernameContains := strings.Contains(strings.ToLower(user.Username), queryLower)

			nameScore := utils.DamerauLevenshteinSimilarity(strings.ToLower(user.Name), queryLower)
			nameMatch := nameScore >= similarityThreshold

			nameContains := strings.Contains(strings.ToLower(user.Name), queryLower)

			emailMatch := strings.Contains(strings.ToLower(user.Email), queryLower)

			var shouldInclude bool
			if requireSubstring {
				shouldInclude = usernameContains || nameContains || emailMatch ||
					(usernameScore >= 0.9) || (nameScore >= 0.9)
			} else {

				shouldInclude = usernameMatch || nameMatch || emailMatch || usernameContains || nameContains
			}

			log.Printf("Fuzzy match for user '%s': username=%.2f (match=%v), name=%.2f (match=%v), query='%s', threshold=%.2f, include=%v",
				user.Username, usernameScore, usernameMatch, nameScore, nameMatch, queryLower, similarityThreshold, shouldInclude)

			if shouldInclude {
				log.Printf("---> MATCHED user '%s' with query '%s' (username=%.2f, name=%.2f, substring=%v)",
					user.Username, queryLower, usernameScore, nameScore, usernameContains || nameContains)
				allUsers = append(allUsers, user)
			}
		}
	}

	totalCount := len(allUsers)

	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	if startIndex >= len(allUsers) {
		return []*User{}, totalCount, nil
	}
	if endIndex > len(allUsers) {
		endIndex = len(allUsers)
	}

	paginatedUsers := allUsers[startIndex:endIndex]
	log.Printf("SearchUsers found %d users (total count: %d) with filter %s after fuzzy matching",
		len(paginatedUsers), totalCount, filter)

	return paginatedUsers, totalCount, nil
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
			user.IsFollowing = true
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
		IsPrivate:         u.IsPrivate,
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
		Page:     int32(page),
		Limit:    int32(limit),
		SortBy:   sortBy,
		SortDesc: !ascending,
	}

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

	log.Printf("Getting blocked users for user %s (page: %d, limit: %d)", userID, page, limit)

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

func (c *GRPCUserServiceClient) UserExists(userID string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.GetUser(ctx, &userProto.GetUserRequest{
		UserId: userID,
	})

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}

	return resp.User != nil, nil
}
