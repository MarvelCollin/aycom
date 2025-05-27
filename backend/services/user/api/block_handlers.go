package handlers

import (
	"context"
	"log"

	"aycom/backend/proto/user"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BlockUser handles a user blocking another user
func (h *UserHandler) BlockUser(ctx context.Context, req *user.BlockUserRequest) (*user.BlockUserResponse, error) {
	if req.UserId == "" || req.BlockedById == "" {
		log.Printf("BlockUser: Missing required parameters, userId: %s, blockedById: %s", req.UserId, req.BlockedById)
		return nil, status.Error(codes.InvalidArgument, "Both user_id and blocked_by_id are required")
	}

	if req.UserId == req.BlockedById {
		log.Printf("BlockUser: User tried to block themselves, userId: %s", req.UserId)
		return nil, status.Error(codes.InvalidArgument, "User cannot block themselves")
	}

	err := h.svc.BlockUser(ctx, req.BlockedById, req.UserId)
	if err != nil {
		log.Printf("BlockUser: Error blocking user: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to block user: %v", err)
	}

	log.Printf("BlockUser: User %s successfully blocked user %s", req.BlockedById, req.UserId)
	return &user.BlockUserResponse{
		Success: true,
		Message: "User blocked successfully",
	}, nil
}

// UnblockUser handles a user unblocking another user
func (h *UserHandler) UnblockUser(ctx context.Context, req *user.UnblockUserRequest) (*user.UnblockUserResponse, error) {
	if req.UserId == "" || req.UnblockedById == "" {
		log.Printf("UnblockUser: Missing required parameters, userId: %s, unblockedById: %s", req.UserId, req.UnblockedById)
		return nil, status.Error(codes.InvalidArgument, "Both user_id and unblocked_by_id are required")
	}

	if req.UserId == req.UnblockedById {
		log.Printf("UnblockUser: User tried to unblock themselves, userId: %s", req.UserId)
		return nil, status.Error(codes.InvalidArgument, "User cannot unblock themselves")
	}

	err := h.svc.UnblockUser(ctx, req.UnblockedById, req.UserId)
	if err != nil {
		log.Printf("UnblockUser: Error unblocking user: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to unblock user: %v", err)
	}

	log.Printf("UnblockUser: User %s successfully unblocked user %s", req.UnblockedById, req.UserId)
	return &user.UnblockUserResponse{
		Success: true,
		Message: "User unblocked successfully",
	}, nil
}

// IsUserBlocked checks if a user is blocked by another user
func (h *UserHandler) IsUserBlocked(ctx context.Context, req *user.IsUserBlockedRequest) (*user.IsUserBlockedResponse, error) {
	if req.UserId == "" || req.BlockedById == "" {
		log.Printf("IsUserBlocked: Missing required parameters, userId: %s, blockedById: %s", req.UserId, req.BlockedById)
		return nil, status.Error(codes.InvalidArgument, "Both user_id and blocked_by_id are required")
	}

	isBlocked, err := h.svc.IsUserBlocked(ctx, req.UserId, req.BlockedById)
	if err != nil {
		log.Printf("IsUserBlocked: Error checking block status: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to check blocked status: %v", err)
	}

	log.Printf("IsUserBlocked: User %s is blocked by %s: %t", req.UserId, req.BlockedById, isBlocked)
	return &user.IsUserBlockedResponse{
		IsBlocked: isBlocked,
	}, nil
}

// ReportUser handles a user reporting another user
func (h *UserHandler) ReportUser(ctx context.Context, req *user.ReportUserRequest) (*user.ReportUserResponse, error) {
	if req.UserId == "" || req.ReportedById == "" || req.Reason == "" {
		log.Printf("ReportUser: Missing required parameters, userId: %s, reportedById: %s, reason: %s", req.UserId, req.ReportedById, req.Reason)
		return nil, status.Error(codes.InvalidArgument, "User ID, reporter ID, and reason are required")
	}

	if req.UserId == req.ReportedById {
		log.Printf("ReportUser: User tried to report themselves, userId: %s", req.UserId)
		return nil, status.Error(codes.InvalidArgument, "User cannot report themselves")
	}

	err := h.svc.ReportUser(ctx, req.ReportedById, req.UserId, req.Reason)
	if err != nil {
		log.Printf("ReportUser: Error reporting user: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to report user: %v", err)
	}

	log.Printf("ReportUser: User %s successfully reported user %s: %s", req.ReportedById, req.UserId, req.Reason)
	return &user.ReportUserResponse{
		Success: true,
		Message: "User reported successfully. Our team will review this report.",
	}, nil
}
