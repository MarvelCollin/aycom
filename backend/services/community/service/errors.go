package service

import "errors"

// Error definitions for the chat service
var (
	ErrChatNotFound       = errors.New("chat not found")
	ErrNotChatParticipant = errors.New("user is not a participant in this chat")
	ErrUserNotFound       = errors.New("user not found")
	ErrMessageNotFound    = errors.New("message not found")
	ErrInvalidIDFormat    = errors.New("invalid ID format")
	ErrPermissionDenied   = errors.New("permission denied")
)
