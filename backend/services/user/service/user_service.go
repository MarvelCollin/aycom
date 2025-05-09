package service

import (
	"context"
	"errors"
	"log"
	"time"

	"aycom/backend/proto/user"
	"aycom/backend/services/user/db"
	"aycom/backend/services/user/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// UserService defines the methods for user-related operations
type UserService interface {
	CreateUserProfile(ctx context.Context, req *user.CreateUserRequest) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUserProfile(ctx context.Context, req *user.UpdateUserRequest) (*model.User, error)
	UpdateUserVerificationStatus(ctx context.Context, req *user.UpdateUserVerificationStatusRequest) error
	DeleteUser(ctx context.Context, id string) error
	LoginUser(ctx context.Context, req *user.LoginUserRequest) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)

	// New social functions
	FollowUser(ctx context.Context, req *model.FollowUserRequest) error
	UnfollowUser(ctx context.Context, req *model.UnfollowUserRequest) error
	GetFollowers(ctx context.Context, req *model.GetFollowersRequest) ([]*model.User, int, error)
	GetFollowing(ctx context.Context, req *model.GetFollowingRequest) ([]*model.User, int, error)
	SearchUsers(ctx context.Context, req *model.SearchUsersRequest) ([]*model.User, int, error)
	GetRecommendedUsers(ctx context.Context, userID string, limit int) ([]*model.User, error)
}

// userService implements the UserService interface
type userService struct {
	repo db.UserRepository
}

// UserServiceImpl is used by main for seeding, can be removed if seeding is moved
// type UserServiceImpl struct {
// 	DB   *gorm.DB
// 	repo repository.UserRepository
// 	UserService
// }

// NewUserService creates a new user service
// Changed to accept repository and return the interface type
func NewUserService(repo db.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

// GetMigrationStatus is likely not needed here anymore, moved to main
// func (s *UserServiceImpl) GetMigrationStatus() error { ... }

// CreateUserProfile creates a new user profile in the system
// Renamed from CreateUser, accepts proto request
func (s *userService) CreateUserProfile(ctx context.Context, req *user.CreateUserRequest) (*model.User, error) {
	if req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "Missing user information")
	}
	userProto := req.User
	if userProto.Username == "" || userProto.Email == "" || userProto.Name == "" || userProto.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing required user profile information (incl. password)")
	}
	userID := uuid.New()

	existingUser, err := s.repo.FindUserByEmail(userProto.Email)
	if err == nil && existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "User with this email already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error checking email existence: %v", err)
		return nil, status.Error(codes.Internal, "Failed to check email existence")
	}

	existingUser, err = s.repo.FindUserByUsername(userProto.Username)
	if err == nil && existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "Username already taken")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error checking username existence: %v", err)
		return nil, status.Error(codes.Internal, "Failed to check username existence")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userProto.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, status.Error(codes.Internal, "Failed to process registration")
	}

	// Map proto User to model User, including new fields
	user := &model.User{
		ID:                    userID,
		Username:              userProto.Username,
		Email:                 userProto.Email,
		Name:                  userProto.Name,
		Gender:                userProto.Gender,
		ProfilePictureURL:     userProto.ProfilePictureUrl,
		BannerURL:             userProto.BannerUrl,
		PasswordHash:          string(hashedPassword), // Store hashed password
		SecurityQuestion:      userProto.SecurityQuestion,
		SecurityAnswer:        userProto.SecurityAnswer, // Consider hashing this too?
		SubscribeToNewsletter: userProto.SubscribeToNewsletter,
		IsVerified:            false, // Default to not verified
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		// DateOfBirth needs parsing if required by the model
	}

	// Optional: Parse DateOfBirth if your model.User expects time.Time
	if userProto.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", userProto.DateOfBirth) // Adjust format if needed
		if err == nil {
			user.DateOfBirth = &dob
		} else {
			log.Printf("Warning: Could not parse date of birth '%s': %v", userProto.DateOfBirth, err)
			// Decide if this should be a hard error or just a warning
			// return nil, status.Error(codes.InvalidArgument, "Invalid date of birth format")
		}
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		log.Printf("Error creating user in repository: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create user profile")
	}
	return user, nil
}

// GetUserByID gets a user by their ID
func (s *userService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		// Check if the error is 'record not found'
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "User with ID %s not found", id)
		}
		log.Printf("Error finding user by ID %s: %v", id, err)
		return nil, status.Error(codes.Internal, "Failed to retrieve user")
	}
	return user, nil
}

// GetUserByUsername gets a user by their username
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := s.repo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "User with username %s not found", username)
		}
		log.Printf("Error finding user by username %s: %v", username, err)
		return nil, status.Error(codes.Internal, "Failed to retrieve user")
	}
	return user, nil
}

// UpdateUserProfile updates a user's profile information
// Accepts proto request
func (s *userService) UpdateUserProfile(ctx context.Context, req *user.UpdateUserRequest) (*model.User, error) {
	// Validate User ID
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required for update")
	}

	// Get current user data
	user, err := s.repo.FindUserByID(req.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "User with ID %s not found for update", req.UserId)
		}
		log.Printf("Error finding user by ID %s for update: %v", req.UserId, err)
		return nil, status.Error(codes.Internal, "Failed to retrieve user for update")
	}

	// Apply updates from request (only update non-empty fields)
	updated := false
	if req.Name != "" {
		user.Name = req.Name
		updated = true
	}
	// Add email update if needed, but ensure uniqueness check if changed
	// if req.Email != "" && req.Email != user.Email { user.Email = req.Email; updated = true }
	if req.ProfilePictureUrl != "" {
		user.ProfilePictureURL = req.ProfilePictureUrl
		updated = true
	}
	if req.BannerUrl != "" {
		user.BannerURL = req.BannerUrl
		updated = true
	}

	if !updated {
		return user, nil // No changes detected
	}

	user.UpdatedAt = time.Now()

	// Save updated user
	err = s.repo.UpdateUser(user)
	if err != nil {
		log.Printf("Error updating user profile for ID %s: %v", req.UserId, err)
		return nil, status.Error(codes.Internal, "Failed to update user profile")
	}

	return user, nil
}

// UpdateUserVerificationStatus updates a user's verification status
// Accepts proto request
func (s *userService) UpdateUserVerificationStatus(ctx context.Context, req *user.UpdateUserVerificationStatusRequest) error {
	// Validate User ID
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "User ID is required for verification update")
	}
	err := s.repo.UpdateUserVerification(req.UserId, req.IsVerified)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "User with ID %s not found for verification update", req.UserId)
		}
		log.Printf("Error updating verification status for user ID %s: %v", req.UserId, err)
		return status.Error(codes.Internal, "Failed to update verification status")
	}
	return nil
}

// DeleteUser deletes a user by their ID
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	err := s.repo.DeleteUser(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "User with ID %s not found for deletion", id)
		}
		log.Printf("Error deleting user ID %s: %v", id, err)
		return status.Error(codes.Internal, "Failed to delete user")
	}
	return nil
}

// LoginUser authenticates a user based on email and password
func (s *userService) LoginUser(ctx context.Context, req *user.LoginUserRequest) (*model.User, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "Email and password are required")
	}

	// Find user by email
	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "User with email %s not found", req.Email)
		}
		log.Printf("Error finding user by email %s: %v", req.Email, err)
		return nil, status.Error(codes.Internal, "Failed to process login")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
	}

	// Log successful login to console (consider structured logging)
	log.Printf("User %s (%s) logged in successfully", user.Username, user.ID)

	// Return user data
	return user, nil
}

// GetUserByEmail gets a user by their email address
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "User with email %s not found", email)
		}
		log.Printf("Error finding user by email %s: %v", email, err)
		return nil, status.Error(codes.Internal, "Failed to retrieve user")
	}
	return user, nil
}

// FollowUser creates a follow relationship between two users
func (s *userService) FollowUser(ctx context.Context, req *model.FollowUserRequest) error {
	if req.FollowerID == "" || req.FollowedID == "" {
		return status.Error(codes.InvalidArgument, "Both follower and followed IDs are required")
	}

	// Validate follower exists
	followerUUID, err := uuid.Parse(req.FollowerID)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Invalid follower ID format")
	}
	_, err = s.repo.FindUserByID(req.FollowerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Error(codes.NotFound, "Follower user not found")
		}
		return status.Error(codes.Internal, "Failed to find follower user")
	}

	// Validate followed user exists
	followedUUID, err := uuid.Parse(req.FollowedID)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Invalid followed ID format")
	}
	_, err = s.repo.FindUserByID(req.FollowedID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Error(codes.NotFound, "Followed user not found")
		}
		return status.Error(codes.Internal, "Failed to find followed user")
	}

	// Check if already following
	exists, err := s.repo.CheckFollowExists(followerUUID, followedUUID)
	if err != nil {
		return status.Error(codes.Internal, "Failed to check follow relationship")
	}
	if exists {
		return status.Error(codes.AlreadyExists, "Already following this user")
	}

	// Create follow relationship
	follow := &model.Follow{
		ID:         uuid.New(),
		FollowerID: followerUUID,
		FollowedID: followedUUID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = s.repo.CreateFollow(follow)
	if err != nil {
		log.Printf("Error creating follow relationship: %v", err)
		return status.Error(codes.Internal, "Failed to create follow relationship")
	}

	return nil
}

// UnfollowUser removes a follow relationship between two users
func (s *userService) UnfollowUser(ctx context.Context, req *model.UnfollowUserRequest) error {
	if req.FollowerID == "" || req.FollowedID == "" {
		return status.Error(codes.InvalidArgument, "Both follower and followed IDs are required")
	}

	// Validate UUID formats
	followerUUID, err := uuid.Parse(req.FollowerID)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Invalid follower ID format")
	}

	followedUUID, err := uuid.Parse(req.FollowedID)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Invalid followed ID format")
	}

	// Check if follow relationship exists
	exists, err := s.repo.CheckFollowExists(followerUUID, followedUUID)
	if err != nil {
		return status.Error(codes.Internal, "Failed to check follow relationship")
	}
	if !exists {
		return status.Error(codes.NotFound, "Follow relationship does not exist")
	}

	// Delete follow relationship
	err = s.repo.DeleteFollow(followerUUID, followedUUID)
	if err != nil {
		log.Printf("Error deleting follow relationship: %v", err)
		return status.Error(codes.Internal, "Failed to delete follow relationship")
	}

	return nil
}

// GetFollowers gets a list of users who follow the specified user
func (s *userService) GetFollowers(ctx context.Context, req *model.GetFollowersRequest) ([]*model.User, int, error) {
	if req.UserID == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Validate user exists
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, 0, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}
	_, err = s.repo.FindUserByID(req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, status.Error(codes.NotFound, "User not found")
		}
		return nil, 0, status.Error(codes.Internal, "Failed to find user")
	}

	// Get followers with pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 10 // Default limit
	}

	followers, total, err := s.repo.GetFollowers(userUUID, page, limit)
	if err != nil {
		log.Printf("Error getting followers: %v", err)
		return nil, 0, status.Error(codes.Internal, "Failed to get followers")
	}

	return followers, total, nil
}

// GetFollowing gets a list of users followed by the specified user
func (s *userService) GetFollowing(ctx context.Context, req *model.GetFollowingRequest) ([]*model.User, int, error) {
	if req.UserID == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Validate user exists
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, 0, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}
	_, err = s.repo.FindUserByID(req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, status.Error(codes.NotFound, "User not found")
		}
		return nil, 0, status.Error(codes.Internal, "Failed to find user")
	}

	// Get followed users with pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 10 // Default limit
	}

	following, total, err := s.repo.GetFollowing(userUUID, page, limit)
	if err != nil {
		log.Printf("Error getting following users: %v", err)
		return nil, 0, status.Error(codes.Internal, "Failed to get following users")
	}

	return following, total, nil
}

// SearchUsers searches for users based on query and filters
func (s *userService) SearchUsers(ctx context.Context, req *model.SearchUsersRequest) ([]*model.User, int, error) {
	// Special case for popular filter - allows empty query to get recommended users
	if req.Filter == "popular" {
		// Get users with highest follower counts
		users, err := s.GetRecommendedUsers(ctx, "", req.Limit)
		if err != nil {
			return nil, 0, err
		}
		return users, len(users), nil
	}

	// Standard search requires a query
	if req.Query == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "Search query is required")
	}

	// Setup pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 10 // Default limit
	}

	// Search users with pagination and filters
	users, total, err := s.repo.SearchUsers(req.Query, req.Filter, page, limit)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		return nil, 0, status.Error(codes.Internal, "Failed to search users")
	}

	return users, total, nil
}

// GetRecommendedUsers returns users with the highest follower counts
func (s *userService) GetRecommendedUsers(ctx context.Context, userID string, limit int) ([]*model.User, error) {
	if limit <= 0 {
		limit = 3 // Default limit if not specified or invalid
	}

	// Call the repository to get recommended users
	users, err := s.repo.GetRecommendedUsers(limit, userID)
	if err != nil {
		log.Printf("Error fetching recommended users: %v", err)
		return nil, status.Error(codes.Internal, "Failed to fetch recommended users")
	}

	return users, nil
}
