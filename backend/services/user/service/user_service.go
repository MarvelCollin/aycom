package service

import (
	"context"
	"errors"
	"log"
	"time"

	"aycom/backend/services/user/db"
	"aycom/backend/services/user/model"
	"aycom/backend/services/user/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// UserService defines the methods for user-related operations
type UserService interface {
	CreateUserProfile(ctx context.Context, req *proto.CreateUserRequest) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUserProfile(ctx context.Context, req *proto.UpdateUserRequest) (*model.User, error)
	UpdateUserVerificationStatus(ctx context.Context, req *proto.UpdateUserVerificationStatusRequest) error
	DeleteUser(ctx context.Context, id string) error
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
func (s *userService) CreateUserProfile(ctx context.Context, req *proto.CreateUserRequest) (*model.User, error) {
	if req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "Missing user information")
	}
	userProto := req.User
	if userProto.Id == "" || userProto.Username == "" || userProto.Email == "" || userProto.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing required user profile information")
	}
	userID, err := uuid.Parse(userProto.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID format: %v", err)
	}
	existingUser, err := s.repo.FindUserByEmail(userProto.Email)
	if err == nil && existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "User with this email already exists")
	}
	existingUser, err = s.repo.FindUserByUsername(userProto.Username)
	if err == nil && existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "Username already taken")
	}
	user := &model.User{
		ID:                userID,
		Username:          userProto.Username,
		Email:             userProto.Email,
		Name:              userProto.Name,
		Gender:            userProto.Gender,
		ProfilePictureURL: userProto.ProfilePictureUrl,
		BannerURL:         userProto.BannerUrl,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	// DateOfBirth, Bio, Location, Website, etc. can be mapped if needed
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
func (s *userService) UpdateUserProfile(ctx context.Context, req *proto.UpdateUserRequest) (*model.User, error) {
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
func (s *userService) UpdateUserVerificationStatus(ctx context.Context, req *proto.UpdateUserVerificationStatusRequest) error {
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
