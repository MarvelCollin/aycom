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

// UserHandler implements the proto.UserServiceServer interface
type UserHandler struct {
	user.UnimplementedUserServiceServer // Embed for forward compatibility
	svc                                 service.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// mapUserModelToProto converts internal model.User to proto.User
func mapUserModelToProto(user *model.User) *user.User {
	if user == nil {
		return nil
	}

	protoUser := &user.User{
		Id:                user.ID.String(),
		Name:              user.Name,
		Username:          user.Username,
		Email:             user.Email,
		Gender:            user.Gender,
		ProfilePictureUrl: user.ProfilePictureURL,
		BannerUrl:         user.BannerURL,
		// Don't include sensitive fields
		// Password, SecurityAnswer, etc.
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
	userProto := mapUserModelToProto(u)
	
	return &user.LoginUserResponse{User: userProto}, nil
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
