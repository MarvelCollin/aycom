package handlers

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"aycom/backend/proto/user"
	"aycom/backend/services/user/model"
	"aycom/backend/services/user/service"
)

type UserHandler struct {
	user.UnimplementedUserServiceServer
	svc      service.UserService
	adminSvc *service.AdminService
}

func NewUserHandler(svc service.UserService, adminSvc *service.AdminService) *UserHandler {
	return &UserHandler{
		svc:      svc,
		adminSvc: adminSvc,
	}
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
		Bio:               u.Bio,
		Location:          u.Location,
		Website:           u.Website,
		ProfilePictureUrl: u.ProfilePictureURL,
		BannerUrl:         u.BannerURL,
		FollowerCount:     int32(u.FollowerCount),
		FollowingCount:    int32(u.FollowingCount),
		IsVerified:        u.IsVerified,
		IsAdmin:           u.IsAdmin,
		IsBanned:          u.IsBanned,
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
	// Set a default limit if not provided
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	// Log the request
	log.Printf("GetRecommendedUsers: Processing request with limit: %d", limit)

	// Get recommended users from service
	users, err := h.svc.GetRecommendedUsers(ctx, limit)
	if err != nil {
		log.Printf("GetRecommendedUsers: Error getting recommended users: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get recommended users")
	}

	// Convert users to proto format
	userProtos := make([]*user.User, 0, len(users))
	for _, u := range users {
		userProto := mapUserModelToProto(u)
		if userProto != nil {
			userProtos = append(userProtos, userProto)
		}
	}

	log.Printf("GetRecommendedUsers: Successfully returned %d recommended users", len(userProtos))

	return &user.GetRecommendedUsersResponse{
		Users: userProtos,
	}, nil
}

func (h *UserHandler) GetAllUsers(ctx context.Context, req *user.GetAllUsersRequest) (*user.GetAllUsersResponse, error) {
	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}

	limit := int(req.GetLimit())
	if limit <= 0 || limit > 100 {
		limit = 10 // Default limit
	}

	// Use !GetSortDesc() instead of GetAscending()
	users, total, err := h.svc.GetAllUsers(ctx, page, limit, req.GetSortBy(), !req.GetSortDesc())
	if err != nil {
		return nil, err
	}

	userProtos := make([]*user.User, 0, len(users))
	for _, u := range users {
		userProto := mapUserModelToProto(u)
		if userProto != nil {
			userProto.IsVerified = u.IsVerified
			userProtos = append(userProtos, userProto)
		}
	}

	return &user.GetAllUsersResponse{
		Users:      userProtos,
		TotalCount: int32(total),
		Page:       int32(page),
		// Remove TotalPages field as it doesn't exist in the proto
	}, nil
}

func (h *UserHandler) BanUser(ctx context.Context, req *user.BanUserRequest) (*user.BanUserResponse, error) {
	return h.adminSvc.BanUser(ctx, req)
}

func (h *UserHandler) SendNewsletter(ctx context.Context, req *user.SendNewsletterRequest) (*user.SendNewsletterResponse, error) {
	return h.adminSvc.SendNewsletter(ctx, req)
}

func (h *UserHandler) GetCommunityRequests(ctx context.Context, req *user.GetCommunityRequestsRequest) (*user.GetCommunityRequestsResponse, error) {
	return h.adminSvc.GetCommunityRequests(ctx, req)
}

func (h *UserHandler) ProcessCommunityRequest(ctx context.Context, req *user.ProcessCommunityRequestRequest) (*user.ProcessCommunityRequestResponse, error) {
	return h.adminSvc.ProcessCommunityRequest(ctx, req)
}

func (h *UserHandler) GetPremiumRequests(ctx context.Context, req *user.GetPremiumRequestsRequest) (*user.GetPremiumRequestsResponse, error) {
	return h.adminSvc.GetPremiumRequests(ctx, req)
}

func (h *UserHandler) ProcessPremiumRequest(ctx context.Context, req *user.ProcessPremiumRequestRequest) (*user.ProcessPremiumRequestResponse, error) {
	return h.adminSvc.ProcessPremiumRequest(ctx, req)
}

func (h *UserHandler) GetReportRequests(ctx context.Context, req *user.GetReportRequestsRequest) (*user.GetReportRequestsResponse, error) {
	return h.adminSvc.GetReportRequests(ctx, req)
}

func (h *UserHandler) ProcessReportRequest(ctx context.Context, req *user.ProcessReportRequestRequest) (*user.ProcessReportRequestResponse, error) {
	return h.adminSvc.ProcessReportRequest(ctx, req)
}

func (h *UserHandler) GetThreadCategories(ctx context.Context, req *user.GetThreadCategoriesRequest) (*user.GetThreadCategoriesResponse, error) {
	return h.adminSvc.GetThreadCategories(ctx, req)
}

func (h *UserHandler) CreateThreadCategory(ctx context.Context, req *user.CreateThreadCategoryRequest) (*user.CreateThreadCategoryResponse, error) {
	return h.adminSvc.CreateThreadCategory(ctx, req)
}

func (h *UserHandler) UpdateThreadCategory(ctx context.Context, req *user.UpdateThreadCategoryRequest) (*user.UpdateThreadCategoryResponse, error) {
	return h.adminSvc.UpdateThreadCategory(ctx, req)
}

func (h *UserHandler) DeleteThreadCategory(ctx context.Context, req *user.DeleteThreadCategoryRequest) (*user.DeleteThreadCategoryResponse, error) {
	return h.adminSvc.DeleteThreadCategory(ctx, req)
}

func (h *UserHandler) GetCommunityCategories(ctx context.Context, req *user.GetCommunityCategoriesRequest) (*user.GetCommunityCategoriesResponse, error) {
	return h.adminSvc.GetCommunityCategories(ctx, req)
}

func (h *UserHandler) CreateCommunityCategory(ctx context.Context, req *user.CreateCommunityCategoryRequest) (*user.CreateCommunityCategoryResponse, error) {
	return h.adminSvc.CreateCommunityCategory(ctx, req)
}

func (h *UserHandler) UpdateCommunityCategory(ctx context.Context, req *user.UpdateCommunityCategoryRequest) (*user.UpdateCommunityCategoryResponse, error) {
	return h.adminSvc.UpdateCommunityCategory(ctx, req)
}

func (h *UserHandler) DeleteCommunityCategory(ctx context.Context, req *user.DeleteCommunityCategoryRequest) (*user.DeleteCommunityCategoryResponse, error) {
	return h.adminSvc.DeleteCommunityCategory(ctx, req)
}

// GetUserByUsername handles fetching a user by username
func (h *UserHandler) GetUserByUsername(ctx context.Context, req *user.GetUserByUsernameRequest) (*user.GetUserByUsernameResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "Username is required")
	}
	u, err := h.svc.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	return &user.GetUserByUsernameResponse{User: mapUserModelToProto(u)}, nil
}

// IsUserBlocked checks if a user is blocked by another user
func (h *UserHandler) IsUserBlocked(ctx context.Context, req *user.IsUserBlockedRequest) (*user.IsUserBlockedResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you would check if the user is blocked
	return &user.IsUserBlockedResponse{IsBlocked: false}, nil
}

// IsFollowing checks if a user is following another user
func (h *UserHandler) IsFollowing(ctx context.Context, req *user.IsFollowingRequest) (*user.IsFollowingResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you would check if the user is following
	return &user.IsFollowingResponse{IsFollowing: false}, nil
}

// FollowUser handles following a user
func (h *UserHandler) FollowUser(ctx context.Context, req *user.FollowUserRequest) (*user.FollowUserResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you would create a follow relationship
	return &user.FollowUserResponse{Success: true}, nil
}

// UnfollowUser handles unfollowing a user
func (h *UserHandler) UnfollowUser(ctx context.Context, req *user.UnfollowUserRequest) (*user.UnfollowUserResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you would delete a follow relationship
	return &user.UnfollowUserResponse{Success: true}, nil
}

// GetFollowers gets a user's followers
func (h *UserHandler) GetFollowers(ctx context.Context, req *user.GetFollowersRequest) (*user.GetFollowersResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you would get the user's followers
	return &user.GetFollowersResponse{
		Followers:  []*user.User{},
		TotalCount: 0,
		Page:       req.GetPage(),
		Limit:      req.GetLimit(),
	}, nil
}

// GetFollowing gets users a user is following
func (h *UserHandler) GetFollowing(ctx context.Context, req *user.GetFollowingRequest) (*user.GetFollowingResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you would get the users the user is following
	return &user.GetFollowingResponse{
		Following:  []*user.User{},
		TotalCount: 0,
		Page:       req.GetPage(),
		Limit:      req.GetLimit(),
	}, nil
}

// SearchUsers searches for users based on a query
func (h *UserHandler) SearchUsers(ctx context.Context, req *user.SearchUsersRequest) (*user.SearchUsersResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you would search for users
	return &user.SearchUsersResponse{
		Users:      []*user.User{},
		TotalCount: 0,
		Page:       req.GetPage(),
		Limit:      req.GetLimit(),
	}, nil
}
