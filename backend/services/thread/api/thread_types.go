package handlers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

// Create temporary request types to fill in for the missing protobuf types
// These should be removed once the protobuf generation is fixed

// BookmarkReplyRequest message
type BookmarkReplyRequest struct {
	ReplyId string `json:"reply_id"`
	UserId  string `json:"user_id"`
}

// RemoveReplyBookmarkRequest message
type RemoveReplyBookmarkRequest struct {
	ReplyId string `json:"reply_id"`
	UserId  string `json:"user_id"`
}

// BookmarkReply adds a bookmark to a reply
func (h *ThreadHandler) BookmarkReply(ctx context.Context, req *BookmarkReplyRequest) (*emptypb.Empty, error) {
	// Call the interaction service to bookmark the reply
	err := h.interactionService.BookmarkReply(ctx, req.UserId, req.ReplyId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RemoveReplyBookmark removes a bookmark from a reply
func (h *ThreadHandler) RemoveReplyBookmark(ctx context.Context, req *RemoveReplyBookmarkRequest) (*emptypb.Empty, error) {
	// Call the interaction service to remove the bookmark from the reply
	err := h.interactionService.RemoveReplyBookmark(ctx, req.UserId, req.ReplyId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
