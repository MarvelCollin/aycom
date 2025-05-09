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

type UserHandler struct {
	user.UnimplementedUserServiceServer
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) GetService() service.UserService {
	return h.svc
}

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

func (h *UserHandler) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	u, err := h.svc.CreateUserProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	return &user.CreateUserResponse{User: mapUserModelToProto(u)}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	u, err := h.svc.UpdateUserProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	return &user.UpdateUserResponse{User: mapUserModelToProto(u)}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	err := h.svc.DeleteUser(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &user.DeleteUserResponse{Success: true, Message: "User deleted successfully"}, nil
}

func (h *UserHandler) UpdateUserVerificationStatus(ctx context.Context, req *user.UpdateUserVerificationStatusRequest) (*user.UpdateUserVerificationStatusResponse, error) {
	err := h.svc.UpdateUserVerificationStatus(ctx, req)
	if err != nil {
		return nil, err
	}
	return &user.UpdateUserVerificationStatusResponse{Success: true, Message: "Verification status updated"}, nil
}

func (h *UserHandler) LoginUser(ctx context.Context, req *user.LoginUserRequest) (*user.LoginUserResponse, error) {
	u, err := h.svc.LoginUser(ctx, req)
	if err != nil {
		return nil, err
	}

	protoUser := mapUserModelToProto(u)

	return &user.LoginUserResponse{User: protoUser}, nil
}

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

func (h *UserHandler) GetRecommendedUsers(ctx context.Context, req *user.GetRecommendedUsersRequest) (*user.GetRecommendedUsersResponse, error) {
	users, err := h.svc.GetRecommendedUsers(ctx, req.UserId, int(req.Limit))
	if err != nil {
		return nil, err
	}

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
