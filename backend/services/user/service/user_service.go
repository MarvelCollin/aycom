package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	userpb "aycom/backend/proto/user"
	"aycom/backend/services/user/model"
	"aycom/backend/services/user/repository"
	"aycom/backend/services/user/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// ServiceFollowUserResponse represents the response for follow user operation within the service
type ServiceFollowUserResponse struct {
	Success             bool   `json:"success"`
	Message             string `json:"message"`
	WasAlreadyFollowing bool   `json:"was_already_following"`
	IsNowFollowing      bool   `json:"is_now_following"`
}

// ServiceUnfollowUserResponse represents the response for unfollow user operation within the service
type ServiceUnfollowUserResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	WasFollowing   bool   `json:"was_following"`
	IsNowFollowing bool   `json:"is_now_following"`
}

// UserService defines methods for user business logic
type UserService interface {
	CreateUserProfile(ctx context.Context, req *userpb.CreateUserRequest) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUserProfile(ctx context.Context, req *userpb.UpdateUserRequest) (*model.User, error)
	UpdateUserVerificationStatus(ctx context.Context, req *userpb.UpdateUserVerificationStatusRequest) error
	DeleteUser(ctx context.Context, id string) error
	LoginUser(ctx context.Context, req *userpb.LoginUserRequest) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	FollowUser(ctx context.Context, req *model.FollowUserRequest) error
	UnfollowUser(ctx context.Context, req *model.UnfollowUserRequest) error
	IsFollowing(ctx context.Context, followerID, followedID string) (bool, error)
	GetFollowers(ctx context.Context, req *model.GetFollowersRequest) ([]*model.User, int, error)
	GetFollowing(ctx context.Context, req *model.GetFollowingRequest) ([]*model.User, int, error)
	SearchUsers(ctx context.Context, req *model.SearchUsersRequest) ([]*model.User, int, error)
	GetRecommendedUsers(ctx context.Context, limit int) ([]*model.User, error)
	GetAllUsers(ctx context.Context, page, limit int, sortBy string, ascending bool) ([]*model.User, int, error)

	// Block and Report functionality
	BlockUser(ctx context.Context, blockerID, blockedID string) error
	UnblockUser(ctx context.Context, unblockerID, unblockedID string) error
	IsUserBlocked(ctx context.Context, userID, blockedByID string) (bool, error)
	ReportUser(ctx context.Context, reporterID, reportedID string, reason string) error
	GetBlockedUsers(ctx context.Context, userID string, page, limit int) ([]map[string]interface{}, int64, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) CreateUserProfile(ctx context.Context, req *userpb.CreateUserRequest) (*model.User, error) {
	if req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "Missing user information")
	}

	userProto := req.User

	// Validate all required fields with proper validation
	var validationErrors []string

	// 1. Validate name
	if err := utils.ValidateName(userProto.Name); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// 2. Validate username
	if err := utils.ValidateUsername(userProto.Username); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// 3. Validate email
	if err := utils.ValidateEmail(userProto.Email); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// 4. Validate password
	if passwordErrors := utils.ValidatePassword(userProto.Password); len(passwordErrors) > 0 {
		for _, err := range passwordErrors {
			validationErrors = append(validationErrors, err.Error())
		}
	}

	// 5. Validate gender
	if err := utils.ValidateGender(userProto.Gender); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// 6. Validate date of birth
	if err := utils.ValidateDateOfBirth(userProto.DateOfBirth); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// 7. Validate security question and answer
	if err := utils.ValidateSecurityQuestion(userProto.SecurityQuestion, userProto.SecurityAnswer); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// If there are validation errors, return them
	if len(validationErrors) > 0 {
		errorMsg := fmt.Sprintf("Validation failed: %s", strings.Join(validationErrors, "; "))
		return nil, status.Error(codes.InvalidArgument, errorMsg)
	}

	userID := uuid.New()

	// Check if email already exists
	existingUser, err := s.repo.FindUserByEmail(userProto.Email)
	if err == nil && existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "User with this email already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error checking email existence: %v", err)
		return nil, status.Error(codes.Internal, "Failed to check email existence")
	}

	// Check if username already exists
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

	// Create user object
	user := &model.User{
		ID:                    userID,
		Username:              userProto.Username,
		Email:                 userProto.Email,
		Name:                  userProto.Name,
		Gender:                userProto.Gender,
		ProfilePictureURL:     userProto.ProfilePictureUrl,
		BannerURL:             userProto.BannerUrl,
		PasswordHash:          string(hashedPassword),
		SecurityQuestion:      userProto.SecurityQuestion,
		SecurityAnswer:        userProto.SecurityAnswer,
		SubscribeToNewsletter: userProto.SubscribeToNewsletter,
		IsVerified:            userProto.IsVerified,
		IsAdmin:               userProto.IsAdmin,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Parse date of birth properly
	if userProto.DateOfBirth != "" {
		dob, err := utils.ParseCustomDateFormat(userProto.DateOfBirth)
		if err == nil {
			user.DateOfBirth = &dob
		} else {
			log.Printf("Warning: Could not parse date of birth '%s': %v", userProto.DateOfBirth, err)
		}
	}

	// Save the user to database
	err = s.repo.CreateUser(user)
	if err != nil {
		log.Printf("Error creating user in repository: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create user profile")
	}
	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "User with ID %s not found", id)
		}
		log.Printf("Error finding user by ID %s: %v", id, err)
		return nil, status.Error(codes.Internal, "Failed to retrieve user")
	}
	return user, nil
}

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

func (s *userService) UpdateUserProfile(ctx context.Context, req *userpb.UpdateUserRequest) (*model.User, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required for update")
	}

	user, err := s.repo.FindUserByID(req.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "User with ID %s not found for update", req.UserId)
		}
		log.Printf("Error finding user by ID %s for update: %v", req.UserId, err)
		return nil, status.Error(codes.Internal, "Failed to retrieve user for update")
	}

	updated := false

	// Handle direct fields
	if req.Name != "" {
		user.Name = req.Name
		updated = true
	}
	if req.ProfilePictureUrl != "" {
		user.ProfilePictureURL = req.ProfilePictureUrl
		updated = true
	}
	if req.BannerUrl != "" {
		user.BannerURL = req.BannerUrl
		updated = true
	}
	if req.Email != "" {
		user.Email = req.Email
		updated = true
	}

	// Handle fields from the User object
	if req.User != nil {
		if req.User.Bio != "" {
			user.Bio = req.User.Bio
			updated = true
		}
		if req.User.Gender != "" {
			user.Gender = req.User.Gender
			updated = true
		}
		if req.User.DateOfBirth != "" {
			// Convert string date to time.Time if needed
			if date, err := time.Parse("2006-01-02", req.User.DateOfBirth); err == nil {
				user.DateOfBirth = &date
				updated = true
			} else {
				log.Printf("Warning: Invalid date format for DateOfBirth: %s", req.User.DateOfBirth)
			}
		}

		// Handle admin status update explicitly instead of using reflection
		// Check if the admin status field is provided in the request
		if req.User.IsAdmin != user.IsAdmin {
			prevStatus := user.IsAdmin
			user.IsAdmin = req.User.IsAdmin
			updated = true
			log.Printf("User %s admin status updated from %v to %v", req.UserId, prevStatus, user.IsAdmin)
		}
	}

	if !updated {
		return user, nil
	}

	user.UpdatedAt = time.Now()

	err = s.repo.UpdateUser(user)
	if err != nil {
		log.Printf("Error updating user profile for ID %s: %v", req.UserId, err)
		return nil, status.Error(codes.Internal, "Failed to update user profile")
	}

	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	_, err := s.repo.FindUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "User with ID %s not found", id)
		}
		log.Printf("Error finding user by ID %s for deletion: %v", id, err)
		return status.Error(codes.Internal, "Failed to retrieve user for deletion")
	}

	err = s.repo.DeleteUser(id)
	if err != nil {
		log.Printf("Error deleting user with ID %s: %v", id, err)
		return status.Error(codes.Internal, "Failed to delete user")
	}

	return nil
}

func (s *userService) LoginUser(ctx context.Context, req *userpb.LoginUserRequest) (*model.User, error) {
	// Validate email format
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid email format")
	}

	// Password is required
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is required")
	}

	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Use generic error message to avoid user enumeration
			return nil, status.Error(codes.NotFound, "Invalid email or password")
		}
		log.Printf("Error finding user by email for login: %v", err)
		return nil, status.Error(codes.Internal, "Failed to process login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		// Use generic error message to avoid user enumeration
		return nil, status.Error(codes.Unauthenticated, "Invalid email or password")
	}

	return user, nil
}

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

func (s *userService) UpdateUserVerificationStatus(ctx context.Context, req *userpb.UpdateUserVerificationStatusRequest) error {
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "User ID is required")
	}

	err := s.repo.UpdateUserVerification(req.UserId, req.IsVerified)
	if err != nil {
		log.Printf("Error updating verification status for user %s: %v", req.UserId, err)
		return status.Error(codes.Internal, "Failed to update verification status")
	}

	return nil
}

func (s *userService) FollowUser(ctx context.Context, req *model.FollowUserRequest) error {
	if req.FollowerID == "" || req.FollowedID == "" {
		return status.Error(codes.InvalidArgument, "Follower ID and followed ID are required")
	}

	followerUUID, err := uuid.Parse(req.FollowerID)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Invalid follower ID format")
	}

	followedUUID, err := uuid.Parse(req.FollowedID)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Invalid followed ID format")
	}

	// Check if users exist
	_, err = s.repo.FindUserByID(req.FollowerID)
	if err != nil {
		return status.Errorf(codes.NotFound, "Follower user with ID %s not found", req.FollowerID)
	}

	_, err = s.repo.FindUserByID(req.FollowedID)
	if err != nil {
		return status.Errorf(codes.NotFound, "Followed user with ID %s not found", req.FollowedID)
	}

	// Check if already following
	exists, err := s.repo.CheckFollowExists(followerUUID, followedUUID)
	if err != nil {
		log.Printf("Error checking follow existence: %v", err)
		return status.Error(codes.Internal, "Failed to check follow relationship")
	}

	if exists {
		// Don't treat this as an error - just return success
		log.Printf("User %s is already following user %s", req.FollowerID, req.FollowedID)
		return nil
	}

	follow := &model.Follow{
		FollowerID: followerUUID,
		FollowedID: followedUUID,
		CreatedAt:  time.Now(),
	}

	// Use a transaction to ensure atomic updates
	err = s.repo.ExecuteInTransaction(func(tx repository.UserRepository) error {
		// Create follow relationship
		if err := tx.CreateFollow(follow); err != nil {
			return fmt.Errorf("failed to create follow relationship: %w", err)
		}

		// Increment followed user's follower count
		if err := tx.IncrementFollowerCount(req.FollowedID); err != nil {
			return fmt.Errorf("failed to increment follower count: %w", err)
		}

		// Increment follower's following count
		if err := tx.IncrementFollowingCount(req.FollowerID); err != nil {
			return fmt.Errorf("failed to increment following count: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("Transaction failed for follow operation: %v", err)
		return status.Error(codes.Internal, "Failed to follow user")
	}

	return nil
}

func (s *userService) UnfollowUser(ctx context.Context, req *model.UnfollowUserRequest) error {
	if req.FollowerID == "" || req.FollowedID == "" {
		return status.Error(codes.InvalidArgument, "Follower ID and followed ID are required")
	}

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
		log.Printf("Error checking follow existence: %v", err)
		return status.Error(codes.Internal, "Failed to check follow relationship")
	}
	if !exists {
		// Don't treat this as an error - just return success
		log.Printf("User %s is not following user %s - treating as successful unfollow", req.FollowerID, req.FollowedID)
		return nil
	}

	// Use a transaction to ensure atomic updates
	err = s.repo.ExecuteInTransaction(func(tx repository.UserRepository) error {
		// Delete follow relationship
		if err := tx.DeleteFollow(followerUUID, followedUUID); err != nil {
			return fmt.Errorf("failed to delete follow relationship: %w", err)
		}

		// Decrement followed user's follower count
		if err := tx.DecrementFollowerCount(req.FollowedID); err != nil {
			return fmt.Errorf("failed to decrement follower count: %w", err)
		}

		// Decrement follower's following count
		if err := tx.DecrementFollowingCount(req.FollowerID); err != nil {
			return fmt.Errorf("failed to decrement following count: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("Transaction failed for unfollow operation: %v", err)
		return status.Error(codes.Internal, "Failed to unfollow user")
	}

	return nil
}

func (s *userService) GetFollowers(ctx context.Context, req *model.GetFollowersRequest) ([]*model.User, int, error) {
	if req.UserID == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "User ID is required")
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, 0, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	followers, count, err := s.repo.GetFollowers(userUUID, req.Page, req.Limit)
	if err != nil {
		log.Printf("Error getting followers for user %s: %v", req.UserID, err)
		return nil, 0, status.Error(codes.Internal, "Failed to get followers")
	}

	return followers, count, nil
}

func (s *userService) GetFollowing(ctx context.Context, req *model.GetFollowingRequest) ([]*model.User, int, error) {
	if req.UserID == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "User ID is required")
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, 0, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	following, count, err := s.repo.GetFollowing(userUUID, req.Page, req.Limit)
	if err != nil {
		log.Printf("Error getting following for user %s: %v", req.UserID, err)
		return nil, 0, status.Error(codes.Internal, "Failed to get following")
	}

	return following, count, nil
}

func (s *userService) SearchUsers(ctx context.Context, req *model.SearchUsersRequest) ([]*model.User, int, error) {
	if req.Query == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "Search query is required")
	}

	users, count, err := s.repo.SearchUsers(req.Query, req.Filter, req.Page, req.Limit)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		return nil, 0, status.Error(codes.Internal, "Failed to search users")
	}

	return users, count, nil
}

func (s *userService) GetRecommendedUsers(ctx context.Context, limit int) ([]*model.User, error) {
	return s.repo.GetRecommendedUsers(limit, "")
}

func (s *userService) GetAllUsers(ctx context.Context, page, limit int, sortBy string, ascending bool) ([]*model.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10 // Default limit
	}

	users, total, err := s.repo.GetAllUsers(page, limit, sortBy, ascending)
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		return nil, 0, status.Error(codes.Internal, "Failed to retrieve users")
	}

	return users, total, nil
}

// BlockUser implements the UserService interface method for blocking users
func (s *userService) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	// Validate inputs
	if blockerID == "" || blockedID == "" {
		return status.Error(codes.InvalidArgument, "Both blocker ID and blocked ID are required")
	}

	// Check if users exist
	_, err := s.GetUserByID(ctx, blockerID)
	if err != nil {
		return status.Error(codes.NotFound, "Blocker user not found")
	}

	_, err = s.GetUserByID(ctx, blockedID)
	if err != nil {
		return status.Error(codes.NotFound, "User to be blocked not found")
	}

	// Use repository to block user
	if err := s.repo.BlockUser(blockerID, blockedID); err != nil {
		log.Printf("Error blocking user: %v", err)
		return status.Error(codes.Internal, "Failed to block user")
	}

	return nil
}

// UnblockUser implements the UserService interface method for unblocking users
func (s *userService) UnblockUser(ctx context.Context, unblockerID, unblockedID string) error {
	// Validate inputs
	if unblockerID == "" || unblockedID == "" {
		return status.Error(codes.InvalidArgument, "Both unblocker ID and unblocked ID are required")
	}

	// Check if users exist
	_, err := s.GetUserByID(ctx, unblockerID)
	if err != nil {
		return status.Error(codes.NotFound, "Unblocker user not found")
	}

	_, err = s.GetUserByID(ctx, unblockedID)
	if err != nil {
		return status.Error(codes.NotFound, "User to be unblocked not found")
	}

	// Use repository to unblock user
	if err := s.repo.UnblockUser(unblockerID, unblockedID); err != nil {
		log.Printf("Error unblocking user: %v", err)
		return status.Error(codes.Internal, "Failed to unblock user")
	}

	return nil
}

// IsUserBlocked checks if a user is blocked by another user
func (s *userService) IsUserBlocked(ctx context.Context, userID, blockedByID string) (bool, error) {
	// Validate inputs
	if userID == "" || blockedByID == "" {
		return false, status.Error(codes.InvalidArgument, "Both user ID and blocker ID are required")
	}

	// Use repository to check block status
	isBlocked, err := s.repo.IsUserBlocked(userID, blockedByID)
	if err != nil {
		log.Printf("Error checking if user is blocked: %v", err)
		return false, status.Error(codes.Internal, "Failed to check block status")
	}

	return isBlocked, nil
}

// ReportUser implements the UserService interface method for reporting users
func (s *userService) ReportUser(ctx context.Context, reporterID, reportedID string, reason string) error {
	// Validate inputs
	if reporterID == "" || reportedID == "" {
		return status.Error(codes.InvalidArgument, "Both reporter ID and reported ID are required")
	}

	// Check if users exist
	_, err := s.GetUserByID(ctx, reporterID)
	if err != nil {
		return status.Error(codes.NotFound, "Reporter user not found")
	}

	_, err = s.GetUserByID(ctx, reportedID)
	if err != nil {
		return status.Error(codes.NotFound, "User to be reported not found")
	}

	// Use repository to report user
	if err := s.repo.ReportUser(reporterID, reportedID, reason); err != nil {
		log.Printf("Error reporting user: %v", err)
		return status.Error(codes.Internal, "Failed to report user")
	}

	return nil
}

// GetBlockedUsers retrieves a list of blocked users
func (s *userService) GetBlockedUsers(ctx context.Context, userID string, page, limit int) ([]map[string]interface{}, int64, error) {
	// Validate inputs
	if userID == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "User ID is required")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10 // Default limit
	}

	// Use repository to get blocked users
	blockedUsers, total, err := s.repo.GetBlockedUsers(userID, page, limit)
	if err != nil {
		log.Printf("Error getting blocked users: %v", err)
		return nil, 0, status.Error(codes.Internal, "Failed to retrieve blocked users")
	}

	return blockedUsers, total, nil
}

// IsFollowing checks if a user is following another user
func (s *userService) IsFollowing(ctx context.Context, followerID, followedID string) (bool, error) {
	if followerID == "" || followedID == "" {
		return false, status.Error(codes.InvalidArgument, "Follower ID and followed ID are required")
	}

	followerUUID, err := uuid.Parse(followerID)
	if err != nil {
		return false, status.Error(codes.InvalidArgument, "Invalid follower ID format")
	}

	followedUUID, err := uuid.Parse(followedID)
	if err != nil {
		return false, status.Error(codes.InvalidArgument, "Invalid followed ID format")
	}

	exists, err := s.repo.CheckFollowExists(followerUUID, followedUUID)
	if err != nil {
		log.Printf("Error checking follow relationship: %v", err)
		return false, status.Error(codes.Internal, "Failed to check follow relationship")
	}

	return exists, nil
}
