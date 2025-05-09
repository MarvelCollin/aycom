// Package handlers provides the handlers for the user service
package handlers

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"aycom/backend/proto/user"
	"aycom/backend/services/user/model"
	"aycom/backend/services/user/service"
)

// UserHandler implements the user service gRPC interface
type UserHandler struct {
	user.UnimplementedUserServiceServer
	svc service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetService returns the underlying service instance
func (h *UserHandler) GetService() service.UserService {
	return h.svc
}

// mapUserModelToProto maps a model.User to a proto.User
func mapUserModelToProto(u *model.User) *user.User {
	if u == nil {
		return nil
	}

	protoUser := &user.User{
		Id:                u.ID.String(),
		Name:              u.Name,
		Username:          u.Username,
		Email:             u.Email,
		Gender:            u.Gender,
		ProfilePictureUrl: u.ProfilePictureURL,
		BannerUrl:         u.BannerURL,
	}

	// Handle optional time fields
	if u.DateOfBirth != nil {
		protoUser.DateOfBirth = u.DateOfBirth.Format("2006-01-02")
	}

	if !u.CreatedAt.IsZero() {
		protoUser.CreatedAt = u.CreatedAt.Format(time.RFC3339)
	}

	if !u.UpdatedAt.IsZero() {
		protoUser.UpdatedAt = u.UpdatedAt.Format(time.RFC3339)
	}

	return protoUser
}

// GetUser handles the GetUser gRPC request
func (h *UserHandler) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	u, err := h.svc.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &user.GetUserResponse{User: mapUserModelToProto(u)}, nil
}

// CreateUser handles the CreateUser gRPC request
func (h *UserHandler) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	u, err := h.svc.CreateUserProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	return &user.CreateUserResponse{User: mapUserModelToProto(u)}, nil
}

// UpdateUser handles the UpdateUser gRPC request
func (h *UserHandler) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	u, err := h.svc.UpdateUserProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	return &user.UpdateUserResponse{User: mapUserModelToProto(u)}, nil
}

// DeleteUser handles the DeleteUser gRPC request
func (h *UserHandler) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	err := h.svc.DeleteUser(ctx, req.UserId)
	if err != nil {
		// Service layer should return gRPC status errors
		return nil, err
	}
	return &user.DeleteUserResponse{Success: true, Message: "User deleted successfully"}, nil
}

// UpdateUserVerificationStatus handles the UpdateUserVerificationStatus gRPC request
func (h *UserHandler) UpdateUserVerificationStatus(ctx context.Context, req *user.UpdateUserVerificationStatusRequest) (*user.UpdateUserVerificationStatusResponse, error) {
	err := h.svc.UpdateUserVerificationStatus(ctx, req)
	if err != nil {
		// Service layer should return gRPC status errors
		return nil, err
	}
	return &user.UpdateUserVerificationStatusResponse{Success: true, Message: "Verification status updated"}, nil
}

// LoginUser handles user authentication
func (h *UserHandler) LoginUser(ctx context.Context, req *user.LoginUserRequest) (*user.LoginUserResponse, error) {
	u, err := h.svc.LoginUser(ctx, req)
	if err != nil {
		// Error already logged and mapped to gRPC status in service layer
		return nil, err
	}

	// Use the common mapping function for consistency
	protoUser := mapUserModelToProto(u)

	return &user.LoginUserResponse{User: protoUser}, nil
}

// GetUserByEmail handles the GetUserByEmail gRPC request
func (h *UserHandler) GetUserByEmail(ctx context.Context, req *user.GetUserByEmailRequest) (*user.GetUserByEmailResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}
	u, err := h.svc.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	return &user.GetUserByEmailResponse{User: mapUserModelToProto(u)}, nil
}

// GetRecommendedUsers handles the GetRecommendedUsers gRPC request
func (h *UserHandler) GetRecommendedUsers(ctx context.Context, req *user.GetRecommendedUsersRequest) (*user.GetRecommendedUsersResponse, error) {
	// Get recommended users from the service
	users, err := h.svc.GetRecommendedUsers(ctx, req.UserId, int(req.Limit))
	if err != nil {
		return nil, err // Service layer should return gRPC status errors
	}

	// Map user models to protos
	userProtos := make([]*user.User, 0, len(users))
	for _, u := range users {
		userProto := mapUserModelToProto(u)
		if userProto != nil {
			userProtos = append(userProtos, userProto)
		}
	}

	return &user.GetRecommendedUsersResponse{
		Users: userProtos,
	}, nil
}
