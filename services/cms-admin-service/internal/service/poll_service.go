package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PollService handles poll business logic
type PollService struct {
	repo        *repository.PollRepository
	articleRepo *repository.ArticleRepository
}

// NewPollService creates a new poll service
func NewPollService(repo *repository.PollRepository, articleRepo *repository.ArticleRepository) *PollService {
	return &PollService{
		repo:        repo,
		articleRepo: articleRepo,
	}
}

// CreatePoll creates a new poll for an article
func (s *PollService) CreatePoll(ctx context.Context, poll *model.Poll, userID string, userRole model.Role) error {
	// Only editors and moderators can create polls
	if userRole != model.RoleEditor && userRole != model.RoleModerator {
		return fmt.Errorf("insufficient permissions: only editors and moderators can create polls")
	}

	// Validate poll
	poll.Question = strings.TrimSpace(poll.Question)
	if poll.Question == "" {
		return fmt.Errorf("poll question cannot be empty")
	}

	if len(poll.Options) < 2 {
		return fmt.Errorf("poll must have at least 2 options")
	}

	if len(poll.Options) > 10 {
		return fmt.Errorf("poll cannot have more than 10 options")
	}

	// Validate and assign IDs to options
	for i := range poll.Options {
		poll.Options[i].Text = strings.TrimSpace(poll.Options[i].Text)
		if poll.Options[i].Text == "" {
			return fmt.Errorf("poll option text cannot be empty")
		}
		if poll.Options[i].ID == "" {
			poll.Options[i].ID = fmt.Sprintf("option_%d", i+1)
		}
	}

	// Validate multiple selection settings
	if poll.IsMultiple && poll.MaxSelections > len(poll.Options) {
		poll.MaxSelections = len(poll.Options)
	}

	poll.CreatedBy = userID

	// Create poll
	if err := s.repo.Create(ctx, poll); err != nil {
		return err
	}

	// Update article to link poll
	article, err := s.articleRepo.FindByID(ctx, poll.ArticleID)
	if err != nil {
		return fmt.Errorf("article not found: %w", err)
	}

	article.HasPoll = true
	article.PollID = &poll.ID

	return s.articleRepo.Update(ctx, article)
}

// GetPoll gets a poll by ID
func (s *PollService) GetPoll(ctx context.Context, pollID primitive.ObjectID, userID string) (*model.Poll, *model.PollVote, error) {
	poll, err := s.repo.FindByID(ctx, pollID)
	if err != nil {
		return nil, nil, err
	}

	// Get user's vote if they voted
	var userVote *model.PollVote
	if userID != "" {
		userVote, _ = s.repo.GetUserVote(ctx, pollID, userID)
	}

	return poll, userVote, nil
}

// GetArticlePoll gets the poll for an article
func (s *PollService) GetArticlePoll(ctx context.Context, articleID primitive.ObjectID, userID string) (*model.Poll, *model.PollVote, error) {
	poll, err := s.repo.FindByArticleID(ctx, articleID)
	if err != nil {
		return nil, nil, err
	}

	// Get user's vote if they voted
	var userVote *model.PollVote
	if userID != "" {
		userVote, _ = s.repo.GetUserVote(ctx, poll.ID, userID)
	}

	return poll, userVote, nil
}

// VoteOnPoll records a user's vote on a poll
func (s *PollService) VoteOnPoll(ctx context.Context, pollID primitive.ObjectID, userID string, optionIDs []string, ipAddress string) error {
	if userID == "" {
		return fmt.Errorf("user must be logged in to vote")
	}

	if len(optionIDs) == 0 {
		return fmt.Errorf("must select at least one option")
	}

	vote := &model.PollVote{
		PollID:    pollID,
		UserID:    userID,
		OptionIDs: optionIDs,
		IPAddress: ipAddress,
	}

	return s.repo.Vote(ctx, vote)
}

// UpdatePoll updates a poll (before any votes)
func (s *PollService) UpdatePoll(ctx context.Context, poll *model.Poll, userID string, userRole model.Role) error {
	// Only editors and moderators can update polls
	if userRole != model.RoleEditor && userRole != model.RoleModerator {
		return fmt.Errorf("insufficient permissions: only editors and moderators can update polls")
	}

	// Get existing poll to check if it has votes
	existing, err := s.repo.FindByID(ctx, poll.ID)
	if err != nil {
		return err
	}

	if existing.TotalVotes > 0 {
		return fmt.Errorf("cannot update poll after votes have been cast")
	}

	// Validate poll
	poll.Question = strings.TrimSpace(poll.Question)
	if poll.Question == "" {
		return fmt.Errorf("poll question cannot be empty")
	}

	if len(poll.Options) < 2 {
		return fmt.Errorf("poll must have at least 2 options")
	}

	return s.repo.Update(ctx, poll)
}

// SetPollStatus activates or deactivates a poll
func (s *PollService) SetPollStatus(ctx context.Context, pollID primitive.ObjectID, isActive bool, userRole model.Role) error {
	// Only editors and moderators can change poll status
	if userRole != model.RoleEditor && userRole != model.RoleModerator {
		return fmt.Errorf("insufficient permissions: only editors and moderators can change poll status")
	}

	return s.repo.SetPollStatus(ctx, pollID, isActive)
}

// GetPollResults gets poll results
func (s *PollService) GetPollResults(ctx context.Context, pollID primitive.ObjectID) (*model.Poll, error) {
	return s.repo.GetPollResults(ctx, pollID)
}

// HasUserVoted checks if a user has voted on a poll
func (s *PollService) HasUserVoted(ctx context.Context, pollID primitive.ObjectID, userID string) (bool, error) {
	return s.repo.HasUserVoted(ctx, pollID, userID)
}
