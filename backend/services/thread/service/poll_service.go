package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"aycom/backend/proto/thread"
	"aycom/backend/services/thread/model"
	"aycom/backend/services/thread/repository"
)

type PollService interface {
	CreatePoll(ctx context.Context, threadID string, req *thread.PollRequest) (*model.Poll, []*model.PollOption, error)
	GetPollByID(ctx context.Context, pollID string) (*model.Poll, []*model.PollOption, error)
	GetPollByThreadID(ctx context.Context, threadID string) (*model.Poll, []*model.PollOption, error)
	AddVoteToPoll(ctx context.Context, pollID, optionID, userID string) error
	GetPollResults(ctx context.Context, pollID string, userID *string) (*PollResults, error)
}

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

type PollOptionResult struct {
	OptionID   uuid.UUID
	Text       string
	VoteCount  int64
	Percentage float32
}

type pollService struct {
	pollRepo repository.PollRepository
}

func NewPollService(pollRepo repository.PollRepository) PollService {
	return &pollService{
		pollRepo: pollRepo,
	}
}

func (s *pollService) CreatePoll(ctx context.Context, threadID string, req *thread.PollRequest) (*model.Poll, []*model.PollOption, error) {

	if threadID == "" || req.Question == "" || len(req.Options) < 2 {
		return nil, nil, status.Error(codes.InvalidArgument, "Thread ID, question, and at least 2 options are required")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "Invalid thread ID: %v", err)
	}

	closesAt := time.Now().Add(24 * time.Hour) 
	if req.EndTime != nil {
		closesAt = req.EndTime.AsTime()
	}

	whoCanVote := "Everyone" 

	pollID := uuid.New()
	poll := &model.Poll{
		PollID:     pollID,
		ThreadID:   threadUUID,
		Question:   req.Question,
		ClosesAt:   closesAt,
		WhoCanVote: whoCanVote,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.pollRepo.CreatePoll(poll); err != nil {
		return nil, nil, status.Errorf(codes.Internal, "Failed to create poll: %v", err)
	}

	pollOptions := make([]*model.PollOption, 0, len(req.Options))
	for _, optionText := range req.Options {
		option := &model.PollOption{
			OptionID:  uuid.New(),
			PollID:    pollID,
			Text:      optionText,
			CreatedAt: time.Now(),
		}
		pollOptions = append(pollOptions, option)
	}

	if err := s.pollRepo.CreatePollOptions(pollOptions); err != nil {
		return nil, nil, status.Errorf(codes.Internal, "Failed to create poll options: %v", err)
	}

	return poll, pollOptions, nil
}

func (s *pollService) GetPollByID(ctx context.Context, pollID string) (*model.Poll, []*model.PollOption, error) {
	if pollID == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "Poll ID is required")
	}

	poll, err := s.pollRepo.FindPollByID(pollID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, status.Errorf(codes.NotFound, "Poll with ID %s not found", pollID)
		}
		return nil, nil, status.Errorf(codes.Internal, "Failed to retrieve poll: %v", err)
	}

	options, err := s.pollRepo.FindPollOptionsByPollID(pollID)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "Failed to retrieve poll options: %v", err)
	}

	return poll, options, nil
}

func (s *pollService) GetPollByThreadID(ctx context.Context, threadID string) (*model.Poll, []*model.PollOption, error) {
	if threadID == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	poll, err := s.pollRepo.FindPollByThreadID(threadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, status.Errorf(codes.NotFound, "Poll for thread with ID %s not found", threadID)
		}
		return nil, nil, status.Errorf(codes.Internal, "Failed to retrieve poll: %v", err)
	}

	options, err := s.pollRepo.FindPollOptionsByPollID(poll.PollID.String())
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "Failed to retrieve poll options: %v", err)
	}

	return poll, options, nil
}

func (s *pollService) AddVoteToPoll(ctx context.Context, pollID, optionID, userID string) error {

	if pollID == "" || optionID == "" || userID == "" {
		return status.Error(codes.InvalidArgument, "Poll ID, Option ID, and User ID are required")
	}

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

	poll, err := s.pollRepo.FindPollByID(pollID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "Poll with ID %s not found", pollID)
		}
		return status.Errorf(codes.Internal, "Failed to retrieve poll: %v", err)
	}

	if poll.ClosesAt.Before(time.Now()) {
		return status.Error(codes.FailedPrecondition, "Poll is closed")
	}

	_, err = s.pollRepo.FindPollOptionByID(optionID)
	if err != nil {
		return status.Errorf(codes.NotFound, "Option with ID %s not found", optionID)
	}

	existingVote, err := s.pollRepo.FindVoteByUserAndPoll(userID, pollID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has voted: %v", err)
	}

	if existingVote != nil {

		if err := s.pollRepo.DeleteVote(userID, pollID); err != nil {
			return status.Errorf(codes.Internal, "Failed to delete existing vote: %v", err)
		}
	}

	vote := &model.PollVote{
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

func (s *pollService) GetPollResults(ctx context.Context, pollID string, userID *string) (*PollResults, error) {
	if pollID == "" {
		return nil, status.Error(codes.InvalidArgument, "Poll ID is required")
	}

	poll, err := s.pollRepo.FindPollByID(pollID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Poll with ID %s not found", pollID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve poll: %v", err)
	}

	options, err := s.pollRepo.FindPollOptionsByPollID(pollID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve poll options: %v", err)
	}

	voteCounts, totalVotes, err := s.pollRepo.GetPollVoteCounts(pollID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve vote counts: %v", err)
	}

	isClosed, err := s.pollRepo.IsPollClosed(pollID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check if poll is closed: %v", err)
	}

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