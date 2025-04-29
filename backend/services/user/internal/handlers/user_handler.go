package handlers

import (
	"context"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/model"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserHandler implements the proto.UserServiceServer interface
type UserHandler struct {
	proto.UnimplementedUserServiceServer // Embed for forward compatibility
	svc                                  service.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// mapUserModelToProto converts internal model.User to proto.UserResponse
func mapUserModelToProto(user *model.User) *proto.UserResponse {
	if user == nil {
		return nil
	}
	var createdAt *timestamppb.Timestamp
	if !user.CreatedAt.IsZero() {
		createdAt = timestamppb.New(user.CreatedAt)
	}
	var updatedAt *timestamppb.Timestamp
	if !user.UpdatedAt.IsZero() {
		updatedAt = timestamppb.New(user.UpdatedAt)
	}

	var dobStr string
	if user.DateOfBirth != nil {
		dobStr = user.DateOfBirth.Format("2006-01-02")
	}

	return &proto.UserResponse{
		UserId:            user.ID.String(),
		Name:              user.Name,
		Email:             user.Email,
		Username:          user.Username,
		Gender:            user.Gender,
		DateOfBirth:       dobStr,
		ProfilePictureUrl: user.ProfilePictureURL,
		BannerUrl:         user.BannerURL,
		IsVerified:        user.IsVerified,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}
}

// GetUser handles the GetUser gRPC request
func (h *UserHandler) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.UserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	user, err := h.svc.GetUserByID(ctx, req.UserId)
	if err != nil {
		// Service layer should return gRPC status errors
		return nil, err
	}
	return mapUserModelToProto(user), nil
}

// CreateUserProfile handles the CreateUserProfile gRPC request
func (h *UserHandler) CreateUserProfile(ctx context.Context, req *proto.CreateUserProfileRequest) (*proto.UserResponse, error) {
	user, err := h.svc.CreateUserProfile(ctx, req)
	if err != nil {
		// Service layer should return gRPC status errors
		return nil, err
	}
	return mapUserModelToProto(user), nil
}

// UpdateUser handles the UpdateUser gRPC request
func (h *UserHandler) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UserResponse, error) {
	user, err := h.svc.UpdateUserProfile(ctx, req) // Call the renamed service method
	if err != nil {
		// Service layer should return gRPC status errors
		return nil, err
	}
	return mapUserModelToProto(user), nil
}

// DeleteUser handles the DeleteUser gRPC request
func (h *UserHandler) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	err := h.svc.DeleteUser(ctx, req.UserId)
	if err != nil {
		// Service layer should return gRPC status errors
		return nil, err
	}
	return &proto.DeleteUserResponse{Success: true, Message: "User deleted successfully"}, nil
}

// UpdateUserVerificationStatus handles the UpdateUserVerificationStatus gRPC request
func (h *UserHandler) UpdateUserVerificationStatus(ctx context.Context, req *proto.UpdateUserVerificationStatusRequest) (*proto.UpdateUserVerificationStatusResponse, error) {
	err := h.svc.UpdateUserVerificationStatus(ctx, req)
	if err != nil {
		// Service layer should return gRPC status errors
		return nil, err
	}
	return &proto.UpdateUserVerificationStatusResponse{Success: true, Message: "Verification status updated"}, nil
}

// Ensure file ends cleanly
