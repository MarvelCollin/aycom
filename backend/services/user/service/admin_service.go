package service

import (
	"context"
	"log"
	"time"

	"aycom/backend/proto/user"
	"aycom/backend/services/user/model"
	"aycom/backend/services/user/repository"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapUserModelToProto converts a user model to a proto message
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

// AdminService handles admin-related business logic
type AdminService struct {
	adminRepo *repository.AdminRepository
	userRepo  repository.UserRepository
}

// NewAdminService creates a new admin service
func NewAdminService(adminRepo *repository.AdminRepository, userRepo repository.UserRepository) *AdminService {
	return &AdminService{
		adminRepo: adminRepo,
		userRepo:  userRepo,
	}
}

// BanUser bans or unbans a user
func (s *AdminService) BanUser(ctx context.Context, req *user.BanUserRequest) (*user.BanUserResponse, error) {
	// Validate input
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Check if user exists
	existingUser, err := s.userRepo.FindUserByID(req.UserId)
	if err != nil {
		log.Printf("Error finding user by ID %s: %v", req.UserId, err)
		return nil, status.Error(codes.NotFound, "User not found")
	}

	// Ban or unban the user
	err = s.adminRepo.BanUser(req.UserId, req.Ban)
	if err != nil {
		log.Printf("Error updating ban status for user %s: %v", req.UserId, err)
		return nil, status.Error(codes.Internal, "Failed to update user ban status")
	}

	action := "unbanned"
	if req.Ban {
		action = "banned"
	}

	return &user.BanUserResponse{
		Success: true,
		Message: "User " + existingUser.Username + " has been " + action,
	}, nil
}

// SendNewsletter sends a newsletter to subscribed users
func (s *AdminService) SendNewsletter(ctx context.Context, req *user.SendNewsletterRequest) (*user.SendNewsletterResponse, error) {
	// Validate input
	if req.Subject == "" {
		return nil, status.Error(codes.InvalidArgument, "Newsletter subject is required")
	}
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "Newsletter content is required")
	}

	// Create newsletter record
	adminID, err := uuid.Parse(req.AdminId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid admin ID")
	}

	newsletter := &model.Newsletter{
		Subject:   req.Subject,
		Content:   req.Content,
		SentBy:    adminID,
		SentAt:    time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save newsletter to database
	err = s.adminRepo.CreateNewsletter(newsletter)
	if err != nil {
		log.Printf("Error creating newsletter: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create newsletter")
	}

	// Get all subscribed users
	subscribedUsers, err := s.adminRepo.GetSubscribedUsers()
	if err != nil {
		log.Printf("Error getting subscribed users: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get newsletter subscribers")
	}

	// In a real implementation, you would send emails to all subscribed users here
	// For now, we'll just return the count of recipients
	recipientsCount := len(subscribedUsers)

	return &user.SendNewsletterResponse{
		Success:         true,
		Message:         "Newsletter sent successfully",
		RecipientsCount: int32(recipientsCount),
	}, nil
}

// GetCommunityRequests gets community creation requests with pagination
func (s *AdminService) GetCommunityRequests(ctx context.Context, req *user.GetCommunityRequestsRequest) (*user.GetCommunityRequestsResponse, error) {
	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}

	limit := int(req.GetLimit())
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// Get requests from repository
	requests, total, err := s.adminRepo.GetCommunityRequests(page, limit, req.GetStatus())
	if err != nil {
		log.Printf("Error getting community requests: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get community requests")
	}

	// Map to proto message
	protoRequests := make([]*user.CommunityRequest, 0, len(requests))
	for _, r := range requests {
		// Get requester info
		requester, err := s.userRepo.FindUserByID(r.UserID.String())
		if err != nil {
			log.Printf("Warning: Could not find user %s for community request: %v", r.UserID, err)
			continue
		}

		protoRequest := &user.CommunityRequest{
			Id:          r.ID.String(),
			UserId:      r.UserID.String(),
			Name:        r.Name,
			Description: r.Description,
			CategoryId:  r.CategoryID.String(),
			Status:      r.Status,
			CreatedAt:   r.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   r.UpdatedAt.Format(time.RFC3339),
			Requester:   mapUserModelToProto(requester),
		}
		protoRequests = append(protoRequests, protoRequest)
	}

	return &user.GetCommunityRequestsResponse{
		Requests:   protoRequests,
		TotalCount: int32(total),
		Page:       int32(page),
		Limit:      int32(limit),
	}, nil
}

// ProcessCommunityRequest approves or rejects a community creation request
func (s *AdminService) ProcessCommunityRequest(ctx context.Context, req *user.ProcessCommunityRequestRequest) (*user.ProcessCommunityRequestResponse, error) {
	// Validate input
	if req.RequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "Request ID is required")
	}

	// Check if request exists
	_, err := s.adminRepo.GetCommunityRequestByID(req.RequestId)
	if err != nil {
		log.Printf("Error finding community request by ID %s: %v", req.RequestId, err)
		return nil, status.Error(codes.NotFound, "Community request not found")
	}

	// Process the request
	err = s.adminRepo.ProcessCommunityRequest(req.RequestId, req.Approve)
	if err != nil {
		log.Printf("Error processing community request %s: %v", req.RequestId, err)
		return nil, status.Error(codes.Internal, "Failed to process community request")
	}

	action := "rejected"
	if req.Approve {
		action = "approved"
	}

	// In a real implementation, you would send an email notification to the user here

	return &user.ProcessCommunityRequestResponse{
		Success: true,
		Message: "Community request has been " + action,
	}, nil
}

// GetPremiumRequests gets premium user requests with pagination
func (s *AdminService) GetPremiumRequests(ctx context.Context, req *user.GetPremiumRequestsRequest) (*user.GetPremiumRequestsResponse, error) {
	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}

	limit := int(req.GetLimit())
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// Get requests from repository
	requests, total, err := s.adminRepo.GetPremiumRequests(page, limit, req.GetStatus())
	if err != nil {
		log.Printf("Error getting premium requests: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get premium requests")
	}

	// Map to proto message
	protoRequests := make([]*user.PremiumRequest, 0, len(requests))
	for _, r := range requests {
		// Get requester info
		requester, err := s.userRepo.FindUserByID(r.UserID.String())
		if err != nil {
			log.Printf("Warning: Could not find user %s for premium request: %v", r.UserID, err)
			continue
		}

		protoRequest := &user.PremiumRequest{
			Id:        r.ID.String(),
			UserId:    r.UserID.String(),
			Reason:    r.Reason,
			Status:    r.Status,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
			UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
			Requester: mapUserModelToProto(requester),
		}
		protoRequests = append(protoRequests, protoRequest)
	}

	return &user.GetPremiumRequestsResponse{
		Requests:   protoRequests,
		TotalCount: int32(total),
		Page:       int32(page),
		Limit:      int32(limit),
	}, nil
}

// ProcessPremiumRequest approves or rejects a premium user request
func (s *AdminService) ProcessPremiumRequest(ctx context.Context, req *user.ProcessPremiumRequestRequest) (*user.ProcessPremiumRequestResponse, error) {
	// Validate input
	if req.RequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "Request ID is required")
	}

	// Check if request exists
	_, err := s.adminRepo.GetPremiumRequestByID(req.RequestId)
	if err != nil {
		log.Printf("Error finding premium request by ID %s: %v", req.RequestId, err)
		return nil, status.Error(codes.NotFound, "Premium request not found")
	}

	// Process the request
	err = s.adminRepo.ProcessPremiumRequest(req.RequestId, req.Approve)
	if err != nil {
		log.Printf("Error processing premium request %s: %v", req.RequestId, err)
		return nil, status.Error(codes.Internal, "Failed to process premium request")
	}

	action := "rejected"
	if req.Approve {
		action = "approved"
	}

	// In a real implementation, you would send an email notification to the user here

	return &user.ProcessPremiumRequestResponse{
		Success: true,
		Message: "Premium request has been " + action,
	}, nil
}

// GetReportRequests gets user report requests with pagination
func (s *AdminService) GetReportRequests(ctx context.Context, req *user.GetReportRequestsRequest) (*user.GetReportRequestsResponse, error) {
	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}

	limit := int(req.GetLimit())
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// Get requests from repository
	requests, total, err := s.adminRepo.GetReportRequests(page, limit, req.GetStatus())
	if err != nil {
		log.Printf("Error getting report requests: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get report requests")
	}

	// Map to proto message
	protoRequests := make([]*user.ReportRequest, 0, len(requests))
	for _, r := range requests {
		// Get reporter info
		reporter, err := s.userRepo.FindUserByID(r.ReporterID.String())
		if err != nil {
			log.Printf("Warning: Could not find reporter %s for report request: %v", r.ReporterID, err)
			continue
		}

		// Get reported user info
		reportedUser, err := s.userRepo.FindUserByID(r.ReportedUserID.String())
		if err != nil {
			log.Printf("Warning: Could not find reported user %s for report request: %v", r.ReportedUserID, err)
			continue
		}

		protoRequest := &user.ReportRequest{
			Id:             r.ID.String(),
			ReporterId:     r.ReporterID.String(),
			ReportedUserId: r.ReportedUserID.String(),
			Reason:         r.Reason,
			Status:         r.Status,
			CreatedAt:      r.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      r.UpdatedAt.Format(time.RFC3339),
			Reporter:       mapUserModelToProto(reporter),
			ReportedUser:   mapUserModelToProto(reportedUser),
		}
		protoRequests = append(protoRequests, protoRequest)
	}

	return &user.GetReportRequestsResponse{
		Requests:   protoRequests,
		TotalCount: int32(total),
		Page:       int32(page),
		Limit:      int32(limit),
	}, nil
}

// ProcessReportRequest approves or rejects a report request
func (s *AdminService) ProcessReportRequest(ctx context.Context, req *user.ProcessReportRequestRequest) (*user.ProcessReportRequestResponse, error) {
	// Validate input
	if req.RequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "Request ID is required")
	}

	// Check if request exists
	report, err := s.adminRepo.GetReportRequestByID(req.RequestId)
	if err != nil {
		log.Printf("Error finding report request by ID %s: %v", req.RequestId, err)
		return nil, status.Error(codes.NotFound, "Report request not found")
	}

	// Process the request
	err = s.adminRepo.ProcessReportRequest(req.RequestId, req.Approve)
	if err != nil {
		log.Printf("Error processing report request %s: %v", req.RequestId, err)
		return nil, status.Error(codes.Internal, "Failed to process report request")
	}

	// If approved, ban the reported user
	if req.Approve {
		err = s.adminRepo.BanUser(report.ReportedUserID.String(), true)
		if err != nil {
			log.Printf("Error banning reported user %s: %v", report.ReportedUserID, err)
			return nil, status.Error(codes.Internal, "Failed to ban reported user")
		}
	}

	action := "rejected"
	if req.Approve {
		action = "approved and user has been banned"
	}

	// In a real implementation, you would send email notifications here

	return &user.ProcessReportRequestResponse{
		Success: true,
		Message: "Report has been " + action,
	}, nil
}

// Thread Category Methods

// GetThreadCategories gets thread categories with pagination
func (s *AdminService) GetThreadCategories(ctx context.Context, req *user.GetThreadCategoriesRequest) (*user.GetThreadCategoriesResponse, error) {
	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}

	limit := int(req.GetLimit())
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// Get categories from repository
	categories, total, err := s.adminRepo.GetThreadCategories(page, limit)
	if err != nil {
		log.Printf("Error getting thread categories: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get thread categories")
	}

	// Map to proto message
	protoCategories := make([]*user.ThreadCategory, 0, len(categories))
	for _, c := range categories {
		protoCategory := &user.ThreadCategory{
			Id:          c.ID.String(),
			Name:        c.Name,
			Description: c.Description,
			CreatedAt:   c.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		}
		protoCategories = append(protoCategories, protoCategory)
	}

	return &user.GetThreadCategoriesResponse{
		Categories: protoCategories,
		TotalCount: int32(total),
		Page:       int32(page),
		Limit:      int32(limit),
	}, nil
}

// CreateThreadCategory creates a new thread category
func (s *AdminService) CreateThreadCategory(ctx context.Context, req *user.CreateThreadCategoryRequest) (*user.CreateThreadCategoryResponse, error) {
	// Validate input
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Category name is required")
	}

	// Create category
	category := &model.ThreadCategory{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	err := s.adminRepo.CreateThreadCategory(category)
	if err != nil {
		log.Printf("Error creating thread category: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create thread category")
	}

	return &user.CreateThreadCategoryResponse{
		Category: &user.ThreadCategory{
			Id:          category.ID.String(),
			Name:        category.Name,
			Description: category.Description,
			CreatedAt:   category.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   category.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// UpdateThreadCategory updates an existing thread category
func (s *AdminService) UpdateThreadCategory(ctx context.Context, req *user.UpdateThreadCategoryRequest) (*user.UpdateThreadCategoryResponse, error) {
	// Validate input
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Category ID is required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Category name is required")
	}

	// Update in database
	err := s.adminRepo.UpdateThreadCategory(req.Id, req.Name, req.Description)
	if err != nil {
		log.Printf("Error updating thread category: %v", err)
		return nil, status.Error(codes.Internal, "Failed to update thread category")
	}

	// Return updated category
	return &user.UpdateThreadCategoryResponse{
		Category: &user.ThreadCategory{
			Id:          req.Id,
			Name:        req.Name,
			Description: req.Description,
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
	}, nil
}

// DeleteThreadCategory deletes a thread category
func (s *AdminService) DeleteThreadCategory(ctx context.Context, req *user.DeleteThreadCategoryRequest) (*user.DeleteThreadCategoryResponse, error) {
	// Validate input
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Category ID is required")
	}

	// Delete from database
	err := s.adminRepo.DeleteThreadCategory(req.Id)
	if err != nil {
		log.Printf("Error deleting thread category: %v", err)
		return nil, status.Error(codes.Internal, "Failed to delete thread category")
	}

	return &user.DeleteThreadCategoryResponse{
		Success: true,
		Message: "Thread category deleted successfully",
	}, nil
}

// Community Category Methods

// GetCommunityCategories gets community categories with pagination
func (s *AdminService) GetCommunityCategories(ctx context.Context, req *user.GetCommunityCategoriesRequest) (*user.GetCommunityCategoriesResponse, error) {
	page := int(req.GetPage())
	if page <= 0 {
		page = 1
	}

	limit := int(req.GetLimit())
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// Get categories from repository
	categories, total, err := s.adminRepo.GetCommunityCategories(page, limit)
	if err != nil {
		log.Printf("Error getting community categories: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get community categories")
	}

	// Map to proto message
	protoCategories := make([]*user.CommunityCategory, 0, len(categories))
	for _, c := range categories {
		protoCategory := &user.CommunityCategory{
			Id:          c.ID.String(),
			Name:        c.Name,
			Description: c.Description,
			CreatedAt:   c.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		}
		protoCategories = append(protoCategories, protoCategory)
	}

	return &user.GetCommunityCategoriesResponse{
		Categories: protoCategories,
		TotalCount: int32(total),
		Page:       int32(page),
		Limit:      int32(limit),
	}, nil
}

// CreateCommunityCategory creates a new community category
func (s *AdminService) CreateCommunityCategory(ctx context.Context, req *user.CreateCommunityCategoryRequest) (*user.CreateCommunityCategoryResponse, error) {
	// Validate input
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Category name is required")
	}

	// Create category
	category := &model.CommunityCategory{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	err := s.adminRepo.CreateCommunityCategory(category)
	if err != nil {
		log.Printf("Error creating community category: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create community category")
	}

	return &user.CreateCommunityCategoryResponse{
		Category: &user.CommunityCategory{
			Id:          category.ID.String(),
			Name:        category.Name,
			Description: category.Description,
			CreatedAt:   category.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   category.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// UpdateCommunityCategory updates an existing community category
func (s *AdminService) UpdateCommunityCategory(ctx context.Context, req *user.UpdateCommunityCategoryRequest) (*user.UpdateCommunityCategoryResponse, error) {
	// Validate input
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Category ID is required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Category name is required")
	}

	// Update in database
	err := s.adminRepo.UpdateCommunityCategory(req.Id, req.Name, req.Description)
	if err != nil {
		log.Printf("Error updating community category: %v", err)
		return nil, status.Error(codes.Internal, "Failed to update community category")
	}

	// Return updated category
	return &user.UpdateCommunityCategoryResponse{
		Category: &user.CommunityCategory{
			Id:          req.Id,
			Name:        req.Name,
			Description: req.Description,
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
	}, nil
}

// DeleteCommunityCategory deletes a community category
func (s *AdminService) DeleteCommunityCategory(ctx context.Context, req *user.DeleteCommunityCategoryRequest) (*user.DeleteCommunityCategoryResponse, error) {
	// Validate input
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Category ID is required")
	}

	// Delete from database
	err := s.adminRepo.DeleteCommunityCategory(req.Id)
	if err != nil {
		log.Printf("Error deleting community category: %v", err)
		return nil, status.Error(codes.Internal, "Failed to delete community category")
	}

	return &user.DeleteCommunityCategoryResponse{
		Success: true,
		Message: "Community category deleted successfully",
	}, nil
}
