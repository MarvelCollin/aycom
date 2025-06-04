package repository

import (
	"time"

	"aycom/backend/services/thread/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PollRepository interface {

	CreatePoll(poll *model.Poll) error
	FindPollByID(id string) (*model.Poll, error)
	FindPollByThreadID(threadID string) (*model.Poll, error)
	UpdatePoll(poll *model.Poll) error
	DeletePoll(id string) error
	IsPollClosed(pollID string) (bool, error)

	CreatePollOptions(options []*model.PollOption) error
	FindPollOptionsByPollID(pollID string) ([]*model.PollOption, error)
	FindPollOptionByID(optionID string) (*model.PollOption, error)

	CreateVote(vote *model.PollVote) error
	DeleteVote(userID, pollID string) error
	FindVoteByUserAndPoll(userID, pollID string) (*model.PollVote, error)
	GetPollVoteCounts(pollID string) (map[string]int64, int64, error)
}

type PostgresPollRepository struct {
	db *gorm.DB
}

func NewPollRepository(db *gorm.DB) PollRepository {
	return &PostgresPollRepository{
		db: db,
	}
}

func (r *PostgresPollRepository) CreatePoll(poll *model.Poll) error {
	return r.db.Create(poll).Error
}

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

func (r *PostgresPollRepository) UpdatePoll(poll *model.Poll) error {
	return r.db.Save(poll).Error
}

func (r *PostgresPollRepository) DeletePoll(id string) error {
	pollID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	return r.db.Delete(&model.Poll{}, "poll_id = ?", pollID).Error
}

func (r *PostgresPollRepository) IsPollClosed(pollID string) (bool, error) {
	poll, err := r.FindPollByID(pollID)
	if err != nil {
		return false, err
	}

	return poll.ClosesAt.Before(time.Now()), nil
}

func (r *PostgresPollRepository) CreatePollOptions(options []*model.PollOption) error {
	return r.db.CreateInBatches(options, len(options)).Error
}

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

func (r *PostgresPollRepository) CreateVote(vote *model.PollVote) error {
	return r.db.Create(vote).Error
}

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
			return nil, nil 
		}
		return nil, result.Error
	}

	return &vote, nil
}

func (r *PostgresPollRepository) GetPollVoteCounts(pollID string) (map[string]int64, int64, error) {
	pollUUID, err := uuid.Parse(pollID)
	if err != nil {
		return nil, 0, err
	}

	var votes []model.PollVote
	if err := r.db.Where("poll_id = ?", pollUUID).Find(&votes).Error; err != nil {
		return nil, 0, err
	}

	voteCounts := make(map[string]int64)
	for _, vote := range votes {
		optionID := vote.OptionID.String()
		voteCounts[optionID]++
	}

	return voteCounts, int64(len(votes)), nil
}