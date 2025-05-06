package handlers

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userProto "aycom/backend/proto/user"
	"aycom/backend/services/user/model"
	"aycom/backend/services/user/service"
)

type UserHandler struct {
	userProto.UnimplementedUserServiceServer
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func mapUserModelToProto(user *model.User) *userProto.User {
	if user == nil {
		return nil
	}

	protoUser := &userProto.User{
		Id:                user.ID.String(),
		Name:              user.Name,
		Username:          user.Username,
		Email:             user.Email,
		Gender:            user.Gender,
		ProfilePictureUrl: user.ProfilePictureURL,
		BannerUrl:         user.BannerURL,
		// IsVerified field needs to be added to the proto
		// For now, we'll handle this in the client code
	}

	// Handle optional time fields
	if user.DateOfBirth != nil {
		protoUser.DateOfBirth = user.DateOfBirth.Format("2006-01-02")
	}

	if !user.CreatedAt.IsZero() {
		protoUser.CreatedAt = user.CreatedAt.Format(time.RFC3339)
	}

	if !user.UpdatedAt.IsZero() {
		protoUser.UpdatedAt = user.UpdatedAt.Format(time.RFC3339)
	}

	return protoUser
}

// GetUser handles the GetUser gRPC request
func (h *UserHandler) GetUser(ctx context.Context, req *userProto.GetUserRequest) (*userProto.GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	u, err := h.svc.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &userProto.GetUserResponse{User: mapUserModelToProto(u)}, nil
}

// CreateUser handles the CreateUser gRPC request
func (h *UserHandler) CreateUser(ctx context.Context, req *userProto.CreateUserRequest) (*userProto.CreateUserResponse, error) {
	u, err := h.svc.CreateUserProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	return &userProto.CreateUserResponse{User: mapUserModelToProto(u)}, nil
}

// UpdateUser handles the UpdateUser gRPC request
func (h *UserHandler) UpdateUser(ctx context.Context, req *userProto.UpdateUserRequest) (*userProto.UpdateUserResponse, error) {
	u, err := h.svc.UpdateUserProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	return &userProto.UpdateUserResponse{User: mapUserModelToProto(u)}, nil
}

// DeleteUser handles the DeleteUser gRPC request
func (h *UserHandler) DeleteUser(ctx context.Context, req *userProto.DeleteUserRequest) (*userProto.DeleteUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	err := h.svc.DeleteUser(ctx, req.UserId)
	if err != nil {
		// Service layer should return gRPC status errors
		return nil, err
	}
	return &userProto.DeleteUserResponse{Success: true, Message: "User deleted successfully"}, nil
}

// UpdateUserVerificationStatus handles the UpdateUserVerificationStatus gRPC request
func (h *UserHandler) UpdateUserVerificationStatus(ctx context.Context, req *userProto.UpdateUserVerificationStatusRequest) (*userProto.UpdateUserVerificationStatusResponse, error) {
	err := h.svc.UpdateUserVerificationStatus(ctx, req)
	if err != nil {
		// Service layer should return gRPC status errors
		return nil, err
	}
	return &userProto.UpdateUserVerificationStatusResponse{Success: true, Message: "Verification status updated"}, nil
}

// LoginUser handles user authentication
func (h *UserHandler) LoginUser(ctx context.Context, req *userProto.LoginUserRequest) (*userProto.LoginUserResponse, error) {
	u, err := h.svc.LoginUser(ctx, req)
	if err != nil {
		// Error already logged and mapped to gRPC status in service layer
		return nil, err
	}

	// Use the common mapping function for consistency
	protoUser := mapUserModelToProto(u)

	return &userProto.LoginUserResponse{User: protoUser}, nil
}

// GetUserByEmail handles the GetUserByEmail gRPC request
func (h *UserHandler) GetUserByEmail(ctx context.Context, req *userProto.GetUserByEmailRequest) (*userProto.GetUserByEmailResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}
	u, err := h.svc.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	return &userProto.GetUserByEmailResponse{User: mapUserModelToProto(u)}, nil
}

// TODO: Implement these functions when proto definitions are updated
// func (h *UserHandler) FollowUser(ctx context.Context, req *userProto.FollowUserRequest) (*userProto.FollowUserResponse, error) { ... }
// func (h *UserHandler) UnfollowUser(ctx context.Context, req *userProto.UnfollowUserRequest) (*userProto.UnfollowUserResponse, error) { ... }
// func (h *UserHandler) GetFollowers(ctx context.Context, req *userProto.GetFollowersRequest) (*userProto.GetFollowersResponse, error) { ... }
// func (h *UserHandler) GetFollowing(ctx context.Context, req *userProto.GetFollowingRequest) (*userProto.GetFollowingResponse, error) { ... }
// func (h *UserHandler) SearchUsers(ctx context.Context, req *userProto.SearchUsersRequest) (*userProto.SearchUsersResponse, error) { ... }
