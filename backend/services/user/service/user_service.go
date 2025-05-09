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

type UserService interface {
	CreateUserProfile(ctx context.Context, req *user.CreateUserRequest) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUserProfile(ctx context.Context, req *user.UpdateUserRequest) (*model.User, error)
	UpdateUserVerificationStatus(ctx context.Context, req *user.UpdateUserVerificationStatusRequest) error
	DeleteUser(ctx context.Context, id string) error
	LoginUser(ctx context.Context, req *user.LoginUserRequest) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)

	FollowUser(ctx context.Context, req *model.FollowUserRequest) error
	UnfollowUser(ctx context.Context, req *model.UnfollowUserRequest) error
	GetFollowers(ctx context.Context, req *model.GetFollowersRequest) ([]*model.User, int, error)
	GetFollowing(ctx context.Context, req *model.GetFollowingRequest) ([]*model.User, int, error)
	SearchUsers(ctx context.Context, req *model.SearchUsersRequest) ([]*model.User, int, error)
	GetRecommendedUsers(ctx context.Context, userID string, limit int) ([]*model.User, error)
}

type userService struct {
	repo db.UserRepository
}

func NewUserService(repo db.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

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
		IsVerified:            false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if userProto.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", userProto.DateOfBirth)
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

func (s *userService) UpdateUserProfile(ctx context.Context, req *user.UpdateUserRequest) (*model.User, error) {
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
	// Check if user exists before attempting deletion
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

func (s *userService) LoginUser(ctx context.Context, req *user.LoginUserRequest) (*model.User, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "Email and password are required")
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

func (s *userService) UpdateUserVerificationStatus(ctx context.Context, req *user.UpdateUserVerificationStatusRequest) error {
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
		return status.Error(codes.AlreadyExists, "User is already following this account")
	}

	follow := &model.Follow{
		FollowerID: followerUUID,
		FollowedID: followedUUID,
		CreatedAt:  time.Now(),
	}

	err = s.repo.CreateFollow(follow)
	if err != nil {
		log.Printf("Error creating follow relationship: %v", err)
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

	// Check if the follow relationship exists
	exists, err := s.repo.CheckFollowExists(followerUUID, followedUUID)
	if err != nil {
		log.Printf("Error checking follow existence: %v", err)
		return status.Error(codes.Internal, "Failed to check follow relationship")
	}

	if !exists {
		return status.Error(codes.NotFound, "Follow relationship not found")
	}

	err = s.repo.DeleteFollow(followerUUID, followedUUID)
	if err != nil {
		log.Printf("Error deleting follow relationship: %v", err)
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

func (s *userService) GetRecommendedUsers(ctx context.Context, userID string, limit int) ([]*model.User, error) {
	if limit <= 0 {
		limit = 10
	}

	users, err := s.repo.GetRecommendedUsers(limit, userID)
	if err != nil {
		log.Printf("Error getting recommended users: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get recommended users")
	}

	return users, nil
}
