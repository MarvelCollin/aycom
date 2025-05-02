package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"aycom/backend/services/user/model"
	"aycom/backend/services/user/proto"
	"aycom/backend/services/user/service"
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

// mapUserModelToProto converts internal model.User to proto.User
func mapUserModelToProto(user *model.User) *proto.User {
	if user == nil {
		return nil
	}
	return &proto.User{
		Id:                user.ID.String(),
		Name:              user.Name,
		Username:          user.Username,
		Email:             user.Email,
		Gender:            user.Gender,
		DateOfBirth:       "", // map as needed
		ProfilePictureUrl: user.ProfilePictureURL,
		BannerUrl:         user.BannerURL,
		CreatedAt:         "", // map as needed
		UpdatedAt:         "", // map as needed
	}
}

// GetUser handles the GetUser gRPC request
func (h *UserHandler) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	user, err := h.svc.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &proto.GetUserResponse{User: mapUserModelToProto(user)}, nil
}

// CreateUser handles the CreateUser gRPC request
func (h *UserHandler) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	user, err := h.svc.CreateUserProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	return &proto.CreateUserResponse{User: mapUserModelToProto(user)}, nil
}

// UpdateUser handles the UpdateUser gRPC request
func (h *UserHandler) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
	user, err := h.svc.UpdateUserProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	return &proto.UpdateUserResponse{User: mapUserModelToProto(user)}, nil
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
