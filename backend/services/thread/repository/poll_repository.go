package repository

import (
	"time"

	"aycom/backend/services/thread/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PollRepository defines the methods for poll-related database operations
type PollRepository interface {
	// Poll methods
	CreatePoll(poll *model.Poll) error
	FindPollByID(id string) (*model.Poll, error)
	FindPollByThreadID(threadID string) (*model.Poll, error)
	UpdatePoll(poll *model.Poll) error
	DeletePoll(id string) error
	IsPollClosed(pollID string) (bool, error)

	// Poll option methods
	CreatePollOptions(options []*model.PollOption) error
	FindPollOptionsByPollID(pollID string) ([]*model.PollOption, error)
	FindPollOptionByID(optionID string) (*model.PollOption, error)

	// Poll vote methods
	CreateVote(vote *model.PollVote) error
	DeleteVote(userID, pollID string) error
	FindVoteByUserAndPoll(userID, pollID string) (*model.PollVote, error)
	GetPollVoteCounts(pollID string) (map[string]int64, int64, error)
}

// PostgresPollRepository is the PostgreSQL implementation of PollRepository
type PostgresPollRepository struct {
	db *gorm.DB
}

// NewPollRepository creates a new poll repository
func NewPollRepository(db *gorm.DB) PollRepository {
	return &PostgresPollRepository{
		db: db,
	}
}

// CreatePoll creates a new poll
func (r *PostgresPollRepository) CreatePoll(poll *model.Poll) error {
	return r.db.Create(poll).Error
}

// FindPollByID finds a poll by ID
func (r *PostgresPollRepository) FindPollByID(id string) (*model.Poll, error) {
	pollID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var poll model.Poll
	if err := r.db.Where("poll_id = ?", pollID).First(&poll).Error; err != nil {
		return nil, err
	}

	return &poll, nil
}

// FindPollByThreadID finds a poll by thread ID
func (r *PostgresPollRepository) FindPollByThreadID(threadID string) (*model.Poll, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return nil, err
	}

	var poll model.Poll
	if err := r.db.Where("thread_id = ?", threadUUID).First(&poll).Error; err != nil {
		return nil, err
	}

	return &poll, nil
}

// UpdatePoll updates an existing poll
func (r *PostgresPollRepository) UpdatePoll(poll *model.Poll) error {
	return r.db.Save(poll).Error
}

// DeletePoll deletes a poll
func (r *PostgresPollRepository) DeletePoll(id string) error {
	pollID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	return r.db.Delete(&model.Poll{}, "poll_id = ?", pollID).Error
}

// IsPollClosed checks if a poll is closed
func (r *PostgresPollRepository) IsPollClosed(pollID string) (bool, error) {
	poll, err := r.FindPollByID(pollID)
	if err != nil {
		return false, err
	}

	return poll.ClosesAt.Before(time.Now()), nil
}

// CreatePollOptions creates poll options
func (r *PostgresPollRepository) CreatePollOptions(options []*model.PollOption) error {
	return r.db.CreateInBatches(options, len(options)).Error
}

// FindPollOptionsByPollID finds all options for a poll
func (r *PostgresPollRepository) FindPollOptionsByPollID(pollID string) ([]*model.PollOption, error) {
	pollUUID, err := uuid.Parse(pollID)
	if err != nil {
		return nil, err
	}

	var options []*model.PollOption
	if err := r.db.Where("poll_id = ?", pollUUID).Find(&options).Error; err != nil {
		return nil, err
	}

	return options, nil
}

// FindPollOptionByID finds a poll option by ID
func (r *PostgresPollRepository) FindPollOptionByID(optionID string) (*model.PollOption, error) {
	optionUUID, err := uuid.Parse(optionID)
	if err != nil {
		return nil, err
	}

	var option model.PollOption
	if err := r.db.Where("option_id = ?", optionUUID).First(&option).Error; err != nil {
		return nil, err
	}

	return &option, nil
}

// CreateVote adds a vote to a poll option
func (r *PostgresPollRepository) CreateVote(vote *model.PollVote) error {
	return r.db.Create(vote).Error
}

// DeleteVote removes a vote from a poll
func (r *PostgresPollRepository) DeleteVote(userID, pollID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	pollUUID, err := uuid.Parse(pollID)
	if err != nil {
		return err
	}

	return r.db.Where("user_id = ? AND poll_id = ?", userUUID, pollUUID).Delete(&model.PollVote{}).Error
}

// FindVoteByUserAndPoll finds a user's vote on a specific poll
func (r *PostgresPollRepository) FindVoteByUserAndPoll(userID, pollID string) (*model.PollVote, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	pollUUID, err := uuid.Parse(pollID)
	if err != nil {
		return nil, err
	}

	var vote model.PollVote
	result := r.db.Where("user_id = ? AND poll_id = ?", userUUID, pollUUID).First(&vote)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // No vote found, not an error
		}
		return nil, result.Error
	}

	return &vote, nil
}

// GetPollVoteCounts gets vote counts for all options in a poll
func (r *PostgresPollRepository) GetPollVoteCounts(pollID string) (map[string]int64, int64, error) {
	pollUUID, err := uuid.Parse(pollID)
	if err != nil {
		return nil, 0, err
	}

	// Get all votes for this poll
	var votes []model.PollVote
	if err := r.db.Where("poll_id = ?", pollUUID).Find(&votes).Error; err != nil {
		return nil, 0, err
	}

	// Count votes by option
	voteCounts := make(map[string]int64)
	for _, vote := range votes {
		optionID := vote.OptionID.String()
		voteCounts[optionID]++
	}

	return voteCounts, int64(len(votes)), nil
}
