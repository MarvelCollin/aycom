package service

import (
	userpb "aycom/backend/proto/user"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"aycom/backend/services/user/model"
	"aycom/backend/services/user/repository"
	"aycom/backend/services/user/utils"
)

type ServiceFollowUserResponse struct {
	Success             bool   `json:"success"`
	Message             string `json:"message"`
	WasAlreadyFollowing bool   `json:"was_already_following"`
	IsNowFollowing      bool   `json:"is_now_following"`
}

type ServiceUnfollowUserResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	WasFollowing   bool   `json:"was_following"`
	IsNowFollowing bool   `json:"is_now_following"`
}

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
	GetAllUsers(ctx context.Context, page, limit int, sortBy string, ascending bool, searchQuery string, newsletterOnly bool) ([]*model.User, int, error)

	BlockUser(ctx context.Context, blockerID, blockedID string) error
	UnblockUser(ctx context.Context, unblockerID, unblockedID string) error
	IsUserBlocked(ctx context.Context, userID, blockedByID string) (bool, error)
	ReportUser(ctx context.Context, reporterID, reportedID string, reason string) error
	GetBlockedUsers(ctx context.Context, userID string, page, limit int) ([]map[string]interface{}, int64, error)
	CreatePremiumRequest(ctx context.Context, req *userpb.CreatePremiumRequestRequest) (*userpb.CreatePremiumRequestResponse, error)
	GetNewsletterSubscribers(ctx context.Context, page, limit int) ([]*model.User, int, error)
}

type userService struct {
	repo repository.UserRepository
	db   *gorm.DB
}

func NewUserService(repo repository.UserRepository) UserService {

	var db *gorm.DB
	if repoWithDB, ok := repo.(*repository.PostgresUserRepository); ok {
		db = repoWithDB.GetDB()
	}

	return &userService{
		repo: repo,
		db:   db,
	}
}

func (s *userService) CreateUserProfile(ctx context.Context, req *userpb.CreateUserRequest) (*model.User, error) {
	if req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "Missing user information")
	}

	userProto := req.User

	var validationErrors []string

	if err := utils.ValidateName(userProto.Name); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if err := utils.ValidateUsername(userProto.Username); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if err := utils.ValidateEmail(userProto.Email); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if passwordErrors := utils.ValidatePassword(userProto.Password); len(passwordErrors) > 0 {
		for _, err := range passwordErrors {
			validationErrors = append(validationErrors, err.Error())
		}
	}

	if err := utils.ValidateGender(userProto.Gender); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if err := utils.ValidateDateOfBirth(userProto.DateOfBirth); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if err := utils.ValidateSecurityQuestion(userProto.SecurityQuestion, userProto.SecurityAnswer); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		errorMsg := fmt.Sprintf("Validation failed: %s", strings.Join(validationErrors, "; "))
		return nil, status.Error(codes.InvalidArgument, errorMsg)
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userProto.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, status.Error(codes.Internal, "Failed to process registration")
	}

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

	if userProto.DateOfBirth != "" {
		dob, err := utils.ParseCustomDateFormat(userProto.DateOfBirth)
		if err == nil {
			user.DateOfBirth = &dob
		} else {
			log.Printf("Warning: Could not parse date of birth '%s': %v", userProto.DateOfBirth, err)
		}
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		log.Printf("Error creating user in repository: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create user profile")
	}
	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	log.Printf("GetUserByID called with ID: %s", id)
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		log.Printf("Error finding user by ID %s: %v", id, err)
		return nil, status.Error(codes.NotFound, "User not found")
	}
	log.Printf("Found user with ID %s: %s", id, user.Name)
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

	if req.GetIsPrivate() != user.IsPrivate {
		user.IsPrivate = req.GetIsPrivate()
		updated = true
	}

	if req.User != nil {
		if req.User.Bio != "" {
			user.Bio = req.User.Bio
			updated = true
		}
		if req.User.Gender != "" {
			user.Gender = req.User.Gender
			updated = true
		}
		if req.User.Location != "" {
			user.Location = req.User.Location
			updated = true
		}
		if req.User.Website != "" {
			user.Website = req.User.Website
			updated = true
		}
		if req.User.SecurityQuestion != "" && req.User.SecurityAnswer != "" {
			user.SecurityQuestion = req.User.SecurityQuestion
			user.SecurityAnswer = req.User.SecurityAnswer
			updated = true
		}
		if req.User.DateOfBirth != "" {
			if date, err := time.Parse("2006-01-02", req.User.DateOfBirth); err == nil {
				user.DateOfBirth = &date
				updated = true
			} else {
				log.Printf("Warning: Invalid date format for DateOfBirth: %s", req.User.DateOfBirth)
			}
		}

		if req.User.IsAdmin != user.IsAdmin {
			prevStatus := user.IsAdmin
			user.IsAdmin = req.User.IsAdmin
			updated = true
			log.Printf("User %s admin status updated from %v to %v", req.UserId, prevStatus, user.IsAdmin)
		}

		if req.User.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.User.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Error hashing password: %v", err)
				return nil, status.Error(codes.Internal, "Failed to process password update")
			}
			user.PasswordHash = string(hashedPassword)
			updated = true
			log.Printf("Password updated for user %s", req.UserId)
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

	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid email format")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is required")
	}

	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return nil, status.Error(codes.NotFound, "Invalid email or password")
		}
		log.Printf("Error finding user by email for login: %v", err)
		return nil, status.Error(codes.Internal, "Failed to process login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {

		return nil, status.Error(codes.Unauthenticated, "Invalid email or password")
	}

	
	if user.IsBanned {
		log.Printf("Banned user %s attempted to login", user.Email)
		return nil, status.Error(codes.PermissionDenied, "Your account has been banned. Please contact support.")
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

	log.Printf("FollowUser: Processing request for follower: %s to follow: %s", req.FollowerID, req.FollowedID)

	followerUUID, err := uuid.Parse(req.FollowerID)
	if err != nil {
		log.Printf("FollowUser: Invalid follower ID format: %s - %v", req.FollowerID, err)
		return status.Errorf(codes.InvalidArgument, "Invalid follower ID format: %v", err)
	}

	followedUUID, err := uuid.Parse(req.FollowedID)
	if err != nil {
		log.Printf("FollowUser: Invalid followed ID format: %s - %v", req.FollowedID, err)
		return status.Errorf(codes.InvalidArgument, "Invalid followed ID format: %v", err)
	}

	follower, err := s.repo.FindUserByID(req.FollowerID)
	if err != nil {
		log.Printf("FollowUser: Follower user with ID %s not found: %v", req.FollowerID, err)
		return status.Errorf(codes.NotFound, "Follower user with ID %s not found: %v", req.FollowerID, err)
	}

	followed, err := s.repo.FindUserByID(req.FollowedID)
	if err != nil {
		log.Printf("FollowUser: Followed user with ID %s not found: %v", req.FollowedID, err)
		return status.Errorf(codes.NotFound, "Followed user with ID %s not found: %v", req.FollowedID, err)
	}

	log.Printf("FollowUser: User %s (%s) is attempting to follow user %s (%s)",
		follower.Username, req.FollowerID, followed.Username, req.FollowedID)

	exists, err := s.repo.CheckFollowExists(followerUUID, followedUUID)
	if err != nil {
		log.Printf("FollowUser: Error checking follow existence: %v", err)
		return status.Errorf(codes.Internal, "Failed to check follow relationship: %v", err)
	}

	if exists {
		log.Printf("FollowUser: User %s is already following user %s", req.FollowerID, req.FollowedID)
		return nil
	}

	follow := &model.Follow{
		FollowerID: followerUUID,
		FollowedID: followedUUID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	var userExists bool
	userExists, err = s.repo.UserExists(req.FollowerID)
	if err != nil {
		log.Printf("FollowUser: Error checking follower existence: %v", err)
		return status.Errorf(codes.Internal, "Failed to verify follower existence: %v", err)
	}
	if !userExists {
		log.Printf("FollowUser: Follower user with ID %s does not exist", req.FollowerID)
		return status.Errorf(codes.NotFound, "Follower user with ID %s not found", req.FollowerID)
	}

	userExists, err = s.repo.UserExists(req.FollowedID)
	if err != nil {
		log.Printf("FollowUser: Error checking followed user existence: %v", err)
		return status.Errorf(codes.Internal, "Failed to verify followed user existence: %v", err)
	}
	if !userExists {
		log.Printf("FollowUser: Followed user with ID %s does not exist", req.FollowedID)
		return status.Errorf(codes.NotFound, "Followed user with ID %s not found", req.FollowedID)
	}

	err = s.repo.ExecuteInTransaction(func(tx repository.UserRepository) error {

		if err := tx.CreateFollow(follow); err != nil {
			log.Printf("FollowUser: Failed to create follow relationship: %v", err)
			return fmt.Errorf("failed to create follow relationship: %w", err)
		}

		if err := tx.IncrementFollowerCount(req.FollowedID); err != nil {
			log.Printf("FollowUser: Failed to increment follower count: %v", err)
			return fmt.Errorf("failed to increment follower count: %w", err)
		}

		if err := tx.IncrementFollowingCount(req.FollowerID); err != nil {
			log.Printf("FollowUser: Failed to increment following count: %v", err)
			return fmt.Errorf("failed to increment following count: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("FollowUser: Transaction failed for follow operation: %v", err)
		return status.Errorf(codes.Internal, "Failed to follow user: %v", err)
	}

	log.Printf("FollowUser: Successfully created follow relationship between %s and %s", req.FollowerID, req.FollowedID)
	return nil
}

func (s *userService) UnfollowUser(ctx context.Context, req *model.UnfollowUserRequest) error {
	if req.FollowerID == "" || req.FollowedID == "" {
		return status.Error(codes.InvalidArgument, "Follower ID and followed ID are required")
	}

	log.Printf("UnfollowUser: Processing request for follower: %s to unfollow: %s", req.FollowerID, req.FollowedID)

	followerUUID, err := uuid.Parse(req.FollowerID)
	if err != nil {
		log.Printf("UnfollowUser: Invalid follower ID format: %s - %v", req.FollowerID, err)
		return status.Errorf(codes.InvalidArgument, "Invalid follower ID format: %v", err)
	}

	followedUUID, err := uuid.Parse(req.FollowedID)
	if err != nil {
		log.Printf("UnfollowUser: Invalid followed ID format: %s - %v", req.FollowedID, err)
		return status.Errorf(codes.InvalidArgument, "Invalid followed ID format: %v", err)
	}

	follower, err := s.repo.FindUserByID(req.FollowerID)
	if err != nil {
		log.Printf("UnfollowUser: Follower user with ID %s not found: %v", req.FollowerID, err)
		return status.Errorf(codes.NotFound, "Follower user with ID %s not found: %v", req.FollowerID, err)
	}

	followed, err := s.repo.FindUserByID(req.FollowedID)
	if err != nil {
		log.Printf("UnfollowUser: Followed user with ID %s not found: %v", req.FollowedID, err)
		return status.Errorf(codes.NotFound, "Followed user with ID %s not found: %v", req.FollowedID, err)
	}

	log.Printf("UnfollowUser: User %s (%s) is attempting to unfollow user %s (%s)",
		follower.Username, req.FollowerID, followed.Username, req.FollowedID)

	exists, err := s.repo.CheckFollowExists(followerUUID, followedUUID)
	if err != nil {
		log.Printf("UnfollowUser: Error checking follow existence: %v", err)
		return status.Errorf(codes.Internal, "Failed to check follow relationship: %v", err)
	}

	if !exists {
		log.Printf("UnfollowUser: User %s is not following user %s - treating as successful unfollow", req.FollowerID, req.FollowedID)
		return nil
	}

	var userExists bool
	userExists, err = s.repo.UserExists(req.FollowerID)
	if err != nil {
		log.Printf("UnfollowUser: Error checking follower existence: %v", err)
		return status.Errorf(codes.Internal, "Failed to verify follower existence: %v", err)
	}
	if !userExists {
		log.Printf("UnfollowUser: Follower user with ID %s does not exist", req.FollowerID)
		return status.Errorf(codes.NotFound, "Follower user with ID %s not found", req.FollowerID)
	}

	userExists, err = s.repo.UserExists(req.FollowedID)
	if err != nil {
		log.Printf("UnfollowUser: Error checking followed user existence: %v", err)
		return status.Errorf(codes.Internal, "Failed to verify followed user existence: %v", err)
	}
	if !userExists {
		log.Printf("UnfollowUser: Followed user with ID %s does not exist", req.FollowedID)
		return status.Errorf(codes.NotFound, "Followed user with ID %s not found", req.FollowedID)
	}

	err = s.repo.ExecuteInTransaction(func(tx repository.UserRepository) error {

		if err := tx.DeleteFollow(followerUUID, followedUUID); err != nil {
			log.Printf("UnfollowUser: Failed to delete follow relationship: %v", err)
			return fmt.Errorf("failed to delete follow relationship: %w", err)
		}

		if err := tx.DecrementFollowerCount(req.FollowedID); err != nil {
			log.Printf("UnfollowUser: Failed to decrement follower count: %v", err)
			return fmt.Errorf("failed to decrement follower count: %w", err)
		}

		if err := tx.DecrementFollowingCount(req.FollowerID); err != nil {
			log.Printf("UnfollowUser: Failed to decrement following count: %v", err)
			return fmt.Errorf("failed to decrement following count: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("UnfollowUser: Transaction failed for unfollow operation: %v", err)
		return status.Errorf(codes.Internal, "Failed to unfollow user: %v", err)
	}

	log.Printf("UnfollowUser: Successfully removed follow relationship between %s and %s", req.FollowerID, req.FollowedID)
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

	md, ok := metadata.FromIncomingContext(ctx)
	if ok && len(md["filter"]) > 0 {
		req.Filter = md["filter"][0]
		log.Printf("Got filter from metadata: %s", req.Filter)
	}

	if req.Query == "" && req.Filter == "" {
		log.Printf("SearchUsers error: Both query and filter are empty")
		return nil, 0, status.Error(codes.InvalidArgument, "Search query or filter is required")
	}

	if req.Filter == "" {
		req.Filter = "all"
		log.Printf("Using default filter: %s", req.Filter)
	}

	if req.Query != "" {
		const MAX_QUERY_LENGTH = 50
		if len(req.Query) > MAX_QUERY_LENGTH {
			log.Printf("Search query in user service too long (%d chars), truncating to %d characters", len(req.Query), MAX_QUERY_LENGTH)
			req.Query = req.Query[:MAX_QUERY_LENGTH]
		}
	}

	log.Printf("Searching users with query: '%s', filter: '%s', page: %d, limit: %d",
		req.Query, req.Filter, req.Page, req.Limit)

	users, count, err := s.repo.SearchUsers(req.Query, req.Filter, req.Page, req.Limit)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		return nil, 0, status.Error(codes.Internal, "Failed to search users")
	}

	log.Printf("Search found %d users (total count: %d)", len(users), count)
	return users, count, nil
}

func (s *userService) GetRecommendedUsers(ctx context.Context, limit int) ([]*model.User, error) {
	return s.repo.GetRecommendedUsers(limit, "")
}

func (s *userService) GetAllUsers(ctx context.Context, page, limit int, sortBy string, ascending bool, searchQuery string, newsletterOnly bool) ([]*model.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var users []*model.User
	var total int
	var err error

	if newsletterOnly {
		users, total, err = s.repo.GetNewsletterSubscribers(page, limit)
	} else if searchQuery != "" {
		users, total, err = s.repo.SearchUsers(searchQuery, "", page, limit)
	} else {
		users, total, err = s.repo.GetAllUsers(page, limit, sortBy, ascending)
	}

	if err != nil {
		log.Printf("Error getting users: %v", err)
		return nil, 0, status.Error(codes.Internal, "Failed to retrieve users")
	}

	return users, total, nil
}

func (s *userService) GetNewsletterSubscribers(ctx context.Context, page, limit int) ([]*model.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := s.repo.GetNewsletterSubscribers(page, limit)
	if err != nil {
		log.Printf("Error getting newsletter subscribers: %v", err)
		return nil, 0, status.Error(codes.Internal, "Failed to retrieve newsletter subscribers")
	}

	return users, total, nil
}

func (s *userService) BlockUser(ctx context.Context, blockerID, blockedID string) error {

	if blockerID == "" || blockedID == "" {
		return status.Error(codes.InvalidArgument, "Both blocker ID and blocked ID are required")
	}

	_, err := s.GetUserByID(ctx, blockerID)
	if err != nil {
		return status.Error(codes.NotFound, "Blocker user not found")
	}

	_, err = s.GetUserByID(ctx, blockedID)
	if err != nil {
		return status.Error(codes.NotFound, "User to be blocked not found")
	}

	if err := s.repo.BlockUser(blockerID, blockedID); err != nil {
		log.Printf("Error blocking user: %v", err)
		return status.Error(codes.Internal, "Failed to block user")
	}

	return nil
}

func (s *userService) UnblockUser(ctx context.Context, unblockerID, unblockedID string) error {

	if unblockerID == "" || unblockedID == "" {
		return status.Error(codes.InvalidArgument, "Both unblocker ID and unblocked ID are required")
	}

	_, err := s.GetUserByID(ctx, unblockerID)
	if err != nil {
		return status.Error(codes.NotFound, "Unblocker user not found")
	}

	_, err = s.GetUserByID(ctx, unblockedID)
	if err != nil {
		return status.Error(codes.NotFound, "User to be unblocked not found")
	}

	if err := s.repo.UnblockUser(unblockerID, unblockedID); err != nil {
		log.Printf("Error unblocking user: %v", err)
		return status.Error(codes.Internal, "Failed to unblock user")
	}

	return nil
}

func (s *userService) IsUserBlocked(ctx context.Context, userID, blockedByID string) (bool, error) {

	if userID == "" || blockedByID == "" {
		return false, status.Error(codes.InvalidArgument, "Both user ID and blocker ID are required")
	}

	isBlocked, err := s.repo.IsUserBlocked(userID, blockedByID)
	if err != nil {
		log.Printf("Error checking if user is blocked: %v", err)
		return false, status.Error(codes.Internal, "Failed to check block status")
	}

	return isBlocked, nil
}

func (s *userService) ReportUser(ctx context.Context, reporterID, reportedID string, reason string) error {

	if reporterID == "" || reportedID == "" {
		return status.Error(codes.InvalidArgument, "Both reporter ID and reported ID are required")
	}

	_, err := s.GetUserByID(ctx, reporterID)
	if err != nil {
		return status.Error(codes.NotFound, "Reporter user not found")
	}

	_, err = s.GetUserByID(ctx, reportedID)
	if err != nil {
		return status.Error(codes.NotFound, "User to be reported not found")
	}

	if err := s.repo.ReportUser(reporterID, reportedID, reason); err != nil {
		log.Printf("Error reporting user: %v", err)
		return status.Error(codes.Internal, "Failed to report user")
	}

	return nil
}

func (s *userService) GetBlockedUsers(ctx context.Context, userID string, page, limit int) ([]map[string]interface{}, int64, error) {

	if userID == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "User ID is required")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	blockedUsers, total, err := s.repo.GetBlockedUsers(userID, page, limit)
	if err != nil {
		log.Printf("Error getting blocked users: %v", err)
		return nil, 0, status.Error(codes.Internal, "Failed to retrieve blocked users")
	}

	return blockedUsers, total, nil
}

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

func (s *userService) CreatePremiumRequest(ctx context.Context, req *userpb.CreatePremiumRequestRequest) (*userpb.CreatePremiumRequestResponse, error) {
	log.Printf("CreatePremiumRequest: Starting to process request for user: %s", req.UserId)

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	if req.Reason == "" {
		return nil, status.Error(codes.InvalidArgument, "Reason is required")
	}

	if req.IdentityCardNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "Identity card number is required")
	}

	if req.FacePhotoUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "Face photo URL is required")
	}

	log.Printf("CreatePremiumRequest: FacePhotoURL length: %d", len(req.FacePhotoUrl))

	userIdUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("CreatePremiumRequest: Invalid user ID format: %v", err)
		return nil, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	user, err := s.repo.FindUserByID(req.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("CreatePremiumRequest: User not found: %s", req.UserId)
			return nil, status.Error(codes.NotFound, "User not found")
		}
		log.Printf("CreatePremiumRequest: Failed to verify user: %v", err)
		return nil, status.Error(codes.Internal, "Failed to verify user")
	}

	log.Printf("CreatePremiumRequest: Found user with ID %s and username %s", user.ID, user.Username)

	var existingRequestCount int64
	err = s.db.Model(&model.PremiumRequest{}).
		Where("user_id = ? AND (status = 'pending' OR status = 'approved')", userIdUUID).
		Count(&existingRequestCount).Error

	if err != nil {
		log.Printf("CreatePremiumRequest: Error checking for existing premium requests: %v", err)
		return nil, status.Error(codes.Internal, "Failed to check for existing requests")
	}

	if existingRequestCount > 0 {
		log.Printf("CreatePremiumRequest: User already has a pending or approved premium request")
		return nil, status.Error(codes.AlreadyExists, "User already has a pending or approved premium request")
	}

	encryptedIDNumber, err := encryptSensitiveData(req.IdentityCardNumber)
	if err != nil {
		log.Printf("CreatePremiumRequest: Error encrypting identity card number: %v", err)
		return nil, status.Error(codes.Internal, "Failed to secure sensitive data")
	}

	facePhotoURL := req.FacePhotoUrl
	if len(facePhotoURL) > 1000000 {
		log.Printf("CreatePremiumRequest: Warning - face photo URL is very large (%d bytes), this might cause DB issues", len(facePhotoURL))
	}

	premiumRequestID := uuid.New()
	premiumRequest := &model.PremiumRequest{
		ID:                 premiumRequestID,
		UserID:             userIdUUID,
		Reason:             req.Reason,
		IdentityCardNumber: encryptedIDNumber,
		FacePhotoURL:       facePhotoURL,
		Status:             "pending",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		log.Printf("CreatePremiumRequest: Failed to begin transaction: %v", tx.Error)
		return nil, status.Error(codes.Internal, "Database error")
	}

	err = tx.Create(premiumRequest).Error
	if err != nil {
		tx.Rollback()
		log.Printf("CreatePremiumRequest: Error creating premium request: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create premium request: "+err.Error())
	}

	if err = tx.Commit().Error; err != nil {
		log.Printf("CreatePremiumRequest: Error committing transaction: %v", err)
		return nil, status.Error(codes.Internal, "Failed to commit premium request")
	}

	log.Printf("CreatePremiumRequest: Successfully created premium request with ID %s", premiumRequest.ID)
	return &userpb.CreatePremiumRequestResponse{
		Success: true,
		Message: "Premium verification request submitted successfully",
	}, nil
}

func encryptSensitiveData(data string) (string, error) {

	hashedBytes := sha256.Sum256([]byte(data))
	hashedString := hex.EncodeToString(hashedBytes[:])

	return hashedString, nil
}
