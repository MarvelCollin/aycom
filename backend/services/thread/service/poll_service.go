package service

import (
	"context"
	"errors"
	"time"

	"aycom/backend/services/thread/db"
	"aycom/backend/services/thread/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PollService defines the interface for poll operations
type PollService interface {
	CreatePoll(ctx context.Context, threadID string, req *proto.PollInfo) (*db.Poll, error)
	GetPollByID(ctx context.Context, pollID string) (*db.Poll, error)
	GetPollByThreadID(ctx context.Context, threadID string) (*db.Poll, error)
	AddVoteToPoll(ctx context.Context, pollID, optionID, userID string) error
	GetPollResults(ctx context.Context, pollID string, userID *string) (*PollResults, error)
}

// PollResults represents the results of a poll
type PollResults struct {
	PollID          uuid.UUID
	ThreadID        uuid.UUID
	Question        string
	ClosesAt        time.Time
	Options         []PollOptionResult
	TotalVotes      int64
	HasUserVoted    bool
	UserVotedOption *uuid.UUID
	IsClosed        bool
}

// PollOptionResult represents the result of a poll option
type PollOptionResult struct {
	OptionID   uuid.UUID
	Text       string
	VoteCount  int64
	Percentage float32
}

// pollService implements the PollService interface
type pollService struct {
	pollRepo db.PollRepository
}

// NewPollService creates a new poll service
func NewPollService(pollRepo db.PollRepository) PollService {
	return &pollService{
		pollRepo: pollRepo,
	}
}

// CreatePoll creates a new poll for a thread
func (s *pollService) CreatePoll(ctx context.Context, threadID string, req *proto.PollInfo) (*db.Poll, error) {
	// Validate required fields
	if threadID == "" || req.Question == "" || len(req.Options) < 2 {
		return nil, status.Error(codes.InvalidArgument, "Thread ID, question, and at least 2 options are required")
	}

	// Parse thread ID
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid thread ID: %v", err)
	}

	// Set closing time if not provided
	closesAt := time.Now().Add(24 * time.Hour) // Default: 24 hours from now
	if req.EndTime != nil {
		closesAt = req.EndTime.AsTime()
	}

	// Set who can vote if not provided (this might not be in your PollInfo message)
	whoCanVote := "Everyone" // Default

	// Create poll
	pollID := uuid.New()
	poll := &db.Poll{
		PollID:     pollID,
		ThreadID:   threadUUID,
		Question:   req.Question,
		ClosesAt:   closesAt,
		WhoCanVote: whoCanVote,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Create poll in database
	if err := s.pollRepo.CreatePoll(poll); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create poll: %v", err)
	}

	// Create poll options
	pollOptions := make([]db.PollOption, 0, len(req.Options))
	modelPointers := make([]*db.PollOption, 0, len(req.Options))
	for _, optionText := range req.Options {
		option := &db.PollOption{
			OptionID:  uuid.New(),
			PollID:    pollID,
			Text:      optionText,
			CreatedAt: time.Now(),
		}
		modelPointers = append(modelPointers, option)
		pollOptions = append(pollOptions, *option)
	}

	// Create poll options in database
	if err := s.pollRepo.CreatePollOptions(modelPointers); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create poll options: %v", err)
	}

	// Load options into poll
	poll.Options = pollOptions

	return poll, nil
}

// GetPollByID retrieves a poll by its ID
func (s *pollService) GetPollByID(ctx context.Context, pollID string) (*db.Poll, error) {
	if pollID == "" {
		return nil, status.Error(codes.InvalidArgument, "Poll ID is required")
	}

	poll, err := s.pollRepo.FindPollByID(pollID)
	if err != nil {
		if errors.Is(err, db.ErrPollNotFound) {
			return nil, status.Errorf(codes.NotFound, "Poll with ID %s not found", pollID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve poll: %v", err)
	}

	// Load options
	optionPointers, err := s.pollRepo.FindPollOptionsByPollID(pollID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve poll options: %v", err)
	}

	options := make([]db.PollOption, len(optionPointers))
	for i, opt := range optionPointers {
		options[i] = *opt
	}
	poll.Options = options

	return poll, nil
}

// GetPollByThreadID retrieves a poll by thread ID
func (s *pollService) GetPollByThreadID(ctx context.Context, threadID string) (*db.Poll, error) {
	if threadID == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	poll, err := s.pollRepo.FindPollByThreadID(threadID)
	if err != nil {
		if errors.Is(err, db.ErrPollNotFound) {
			return nil, status.Errorf(codes.NotFound, "Poll for thread with ID %s not found", threadID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve poll: %v", err)
	}

	// Load options
	optionPointers, err := s.pollRepo.FindPollOptionsByPollID(poll.PollID.String())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve poll options: %v", err)
	}

	options := make([]db.PollOption, len(optionPointers))
	for i, opt := range optionPointers {
		options[i] = *opt
	}
	poll.Options = options

	return poll, nil
}

// AddVoteToPoll adds a vote to a poll option
func (s *pollService) AddVoteToPoll(ctx context.Context, pollID, optionID, userID string) error {
	// Validate required fields
	if pollID == "" || optionID == "" || userID == "" {
		return status.Error(codes.InvalidArgument, "Poll ID, Option ID, and User ID are required")
	}

	// Parse IDs
	pollUUID, err := uuid.Parse(pollID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid poll ID: %v", err)
	}

	optionUUID, err := uuid.Parse(optionID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid option ID: %v", err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	// Check if poll exists
	poll, err := s.pollRepo.FindPollByID(pollID)
	if err != nil {
		if errors.Is(err, db.ErrPollNotFound) {
			return status.Errorf(codes.NotFound, "Poll with ID %s not found", pollID)
		}
		return status.Errorf(codes.Internal, "Failed to retrieve poll: %v", err)
	}

	// Check if poll is closed
	if poll.ClosesAt.Before(time.Now()) {
		return status.Error(codes.FailedPrecondition, "Poll is closed")
	}

	// Check if option exists
	_, err = s.pollRepo.FindPollOptionByID(optionID)
	if err != nil {
		return status.Errorf(codes.NotFound, "Option with ID %s not found", optionID)
	}

	// Check if user has already voted
	existingVote, err := s.pollRepo.FindVoteByUserAndPoll(userID, pollID)
	if err == nil && existingVote != nil {
		// Delete existing vote
		if err := s.pollRepo.DeleteVote(userID, pollID); err != nil {
			return status.Errorf(codes.Internal, "Failed to delete existing vote: %v", err)
		}
	}

	// Create new vote
	vote := &db.PollVote{
		VoteID:    uuid.New(),
		PollID:    pollUUID,
		OptionID:  optionUUID,
		UserID:    userUUID,
		CreatedAt: time.Now(),
	}

	if err := s.pollRepo.CreateVote(vote); err != nil {
		return status.Errorf(codes.Internal, "Failed to create vote: %v", err)
	}

	return nil
}

// GetPollResults gets the results of a poll
func (s *pollService) GetPollResults(ctx context.Context, pollID string, userID *string) (*PollResults, error) {
	if pollID == "" {
		return nil, status.Error(codes.InvalidArgument, "Poll ID is required")
	}

	// Get poll
	poll, err := s.pollRepo.FindPollByID(pollID)
	if err != nil {
		if errors.Is(err, db.ErrPollNotFound) {
			return nil, status.Errorf(codes.NotFound, "Poll with ID %s not found", pollID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve poll: %v", err)
	}

	// Get options
	options, err := s.pollRepo.FindPollOptionsByPollID(pollID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve poll options: %v", err)
	}

	// Get vote counts
	voteCounts, totalVotes, err := s.pollRepo.GetPollVoteCounts(pollID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve vote counts: %v", err)
	}

	// Check if poll is closed
	isClosed, err := s.pollRepo.IsPollClosed(pollID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check if poll is closed: %v", err)
	}

	// Prepare option results
	optionResults := make([]PollOptionResult, 0, len(options))
	for _, option := range options {
		voteCount := voteCounts[option.OptionID.String()]
		percentage := float32(0)
		if totalVotes > 0 {
			percentage = float32(voteCount) / float32(totalVotes) * 100
		}

		optionResult := PollOptionResult{
			OptionID:   option.OptionID,
			Text:       option.Text,
			VoteCount:  voteCount,
			Percentage: percentage,
		}
		optionResults = append(optionResults, optionResult)
	}

	// Check if user has voted
	var hasUserVoted bool
	var userVotedOption *uuid.UUID
	if userID != nil {
		existingVote, err := s.pollRepo.FindVoteByUserAndPoll(*userID, pollID)
		if err == nil && existingVote != nil {
			hasUserVoted = true
			userVotedOption = &existingVote.OptionID
		}
	}

	results := &PollResults{
		PollID:          poll.PollID,
		ThreadID:        poll.ThreadID,
		Question:        poll.Question,
		ClosesAt:        poll.ClosesAt,
		Options:         optionResults,
		TotalVotes:      totalVotes,
		HasUserVoted:    hasUserVoted,
		UserVotedOption: userVotedOption,
		IsClosed:        isClosed,
	}

	return results, nil
}

// Add ErrPollNotFound to repository package to properly handle not found errors
var ErrPollNotFound = errors.New("poll not found")
