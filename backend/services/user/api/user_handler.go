package handlers

import (
	"aycom/backend/proto/user"
	"aycom/backend/services/user/model"
	"aycom/backend/services/user/service"
	"context"
	"log"
	"reflect"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		IsPrivate:         u.IsPrivate,
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

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	log.Printf("GetRecommendedUsers: Processing request with limit: %d", limit)

	users, err := h.svc.GetRecommendedUsers(ctx, limit)
	if err != nil {
		log.Printf("GetRecommendedUsers: Error getting recommended users: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get recommended users")
	}

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
		limit = 10
	}

	searchQuery := ""
	newsletterOnly := false

	// Check if the request has GetSearchQuery and GetNewsletterOnly fields
	// in case we're dealing with an older protobuf definition
	if req != nil {
		// Use reflection to check if the fields exist
		reqValue := reflect.ValueOf(req)
		searchQueryMethod := reqValue.MethodByName("GetSearchQuery")
		newsletterOnlyMethod := reqValue.MethodByName("GetNewsletterOnly")

		if searchQueryMethod.IsValid() {
			results := searchQueryMethod.Call(nil)
			if len(results) > 0 {
				searchQuery = results[0].String()
			}
		}

		if newsletterOnlyMethod.IsValid() {
			results := newsletterOnlyMethod.Call(nil)
			if len(results) > 0 {
				newsletterOnly = results[0].Bool()
			}
		}
	}

	users, total, err := h.svc.GetAllUsers(ctx, page, limit, req.GetSortBy(), !req.GetSortDesc(), searchQuery, newsletterOnly)
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

func (h *UserHandler) CreatePremiumRequest(ctx context.Context, req *user.CreatePremiumRequestRequest) (*user.CreatePremiumRequestResponse, error) {
	return h.svc.CreatePremiumRequest(ctx, req)
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

func (h *UserHandler) IsFollowing(ctx context.Context, req *user.IsFollowingRequest) (*user.IsFollowingResponse, error) {
	if req.GetFollowerId() == "" || req.GetFollowedId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Follower ID and followed ID are required")
	}

	exists, err := h.svc.IsFollowing(ctx, req.GetFollowerId(), req.GetFollowedId())
	if err != nil {
		log.Printf("Error checking follow relationship: %v", err)
		return nil, err
	}

	return &user.IsFollowingResponse{IsFollowing: exists}, nil
}

func (h *UserHandler) FollowUser(ctx context.Context, req *user.FollowUserRequest) (*user.FollowUserResponse, error) {
	if req.GetFollowerId() == "" || req.GetFollowedId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Follower ID and followed ID are required")
	}

	wasAlreadyFollowing, err := h.svc.IsFollowing(ctx, req.GetFollowerId(), req.GetFollowedId())
	if err != nil {
		log.Printf("Error checking follow status: %v", err)
		return nil, err
	}

	if wasAlreadyFollowing {
		return &user.FollowUserResponse{
			Success:             true,
			Message:             "Already following this user",
			WasAlreadyFollowing: true,
			IsNowFollowing:      true,
		}, nil
	}

	followReq := &model.FollowUserRequest{
		FollowerID: req.GetFollowerId(),
		FollowedID: req.GetFollowedId(),
	}

	err = h.svc.FollowUser(ctx, followReq)
	if err != nil {
		log.Printf("Error following user: %v", err)
		return nil, err
	}

	return &user.FollowUserResponse{
		Success:             true,
		Message:             "User followed successfully",
		WasAlreadyFollowing: false,
		IsNowFollowing:      true,
	}, nil
}

func (h *UserHandler) UnfollowUser(ctx context.Context, req *user.UnfollowUserRequest) (*user.UnfollowUserResponse, error) {
	if req.GetFollowerId() == "" || req.GetFollowedId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Follower ID and followed ID are required")
	}

	wasFollowing, err := h.svc.IsFollowing(ctx, req.GetFollowerId(), req.GetFollowedId())
	if err != nil {
		log.Printf("Error checking follow status: %v", err)
		return nil, err
	}

	if !wasFollowing {
		return &user.UnfollowUserResponse{
			Success:        true,
			Message:        "Not following this user",
			WasFollowing:   false,
			IsNowFollowing: false,
		}, nil
	}

	unfollowReq := &model.UnfollowUserRequest{
		FollowerID: req.GetFollowerId(),
		FollowedID: req.GetFollowedId(),
	}

	err = h.svc.UnfollowUser(ctx, unfollowReq)
	if err != nil {
		log.Printf("Error unfollowing user: %v", err)
		return nil, err
	}

	return &user.UnfollowUserResponse{
		Success:        true,
		Message:        "User unfollowed successfully",
		WasFollowing:   true,
		IsNowFollowing: false,
	}, nil
}

func (h *UserHandler) GetFollowers(ctx context.Context, req *user.GetFollowersRequest) (*user.GetFollowersResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 20
	}

	followersReq := &model.GetFollowersRequest{
		UserID: req.GetUserId(),
		Page:   page,
		Limit:  limit,
	}

	followers, total, err := h.svc.GetFollowers(ctx, followersReq)
	if err != nil {
		log.Printf("Error getting followers: %v", err)
		return nil, err
	}

	protoFollowers := make([]*user.User, len(followers))
	for i, follower := range followers {
		protoFollowers[i] = mapUserModelToProto(follower)
	}

	return &user.GetFollowersResponse{
		Followers:  protoFollowers,
		TotalCount: int32(total),
		Page:       int32(page),
		Limit:      int32(limit),
	}, nil
}

func (h *UserHandler) GetFollowing(ctx context.Context, req *user.GetFollowingRequest) (*user.GetFollowingResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 20
	}

	followingReq := &model.GetFollowingRequest{
		UserID: req.GetUserId(),
		Page:   page,
		Limit:  limit,
	}

	following, total, err := h.svc.GetFollowing(ctx, followingReq)
	if err != nil {
		log.Printf("Error getting following: %v", err)
		return nil, err
	}

	protoFollowing := make([]*user.User, len(following))
	for i, followedUser := range following {
		protoFollowing[i] = mapUserModelToProto(followedUser)
	}

	return &user.GetFollowingResponse{
		Following:  protoFollowing,
		TotalCount: int32(total),
		Page:       int32(page),
		Limit:      int32(limit),
	}, nil
}

func (h *UserHandler) SearchUsers(ctx context.Context, req *user.SearchUsersRequest) (*user.SearchUsersResponse, error) {

	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 10
	}

	searchReq := &model.SearchUsersRequest{
		Query:  req.GetQuery(),
		Filter: "",
		Page:   page,
		Limit:  limit,
	}

	users, totalCount, err := h.svc.SearchUsers(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	protoUsers := make([]*user.User, len(users))
	for i, u := range users {
		protoUsers[i] = mapUserModelToProto(u)
	}

	return &user.SearchUsersResponse{
		Users:      protoUsers,
		TotalCount: int32(totalCount),
		Page:       int32(page),
		Limit:      int32(limit),
	}, nil
}
