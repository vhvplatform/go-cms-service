package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PollRepository handles poll data operations
type PollRepository struct {
	collection     *mongo.Collection
	voteCollection *mongo.Collection
}

// NewPollRepository creates a new poll repository
func NewPollRepository(db *mongo.Database) *PollRepository {
	return &PollRepository{
		collection:     db.Collection("polls"),
		voteCollection: db.Collection("poll_votes"),
	}
}

// Create creates a new poll
func (r *PollRepository) Create(ctx context.Context, poll *model.Poll) error {
	poll.ID = primitive.NewObjectID()
	poll.CreatedAt = time.Now()
	poll.UpdatedAt = time.Now()
	poll.TotalVotes = 0

	// Initialize vote counts for all options
	for i := range poll.Options {
		poll.Options[i].VoteCount = 0
	}

	_, err := r.collection.InsertOne(ctx, poll)
	return err
}

// FindByID finds a poll by ID
func (r *PollRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Poll, error) {
	var poll model.Poll
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&poll)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Calculate percentages
	r.calculatePercentages(&poll)

	return &poll, nil
}

// FindByArticleID finds the poll for an article
func (r *PollRepository) FindByArticleID(ctx context.Context, articleID primitive.ObjectID) (*model.Poll, error) {
	var poll model.Poll
	err := r.collection.FindOne(ctx, bson.M{"articleId": articleID}).Decode(&poll)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Calculate percentages
	r.calculatePercentages(&poll)

	return &poll, nil
}

// Update updates a poll
func (r *PollRepository) Update(ctx context.Context, poll *model.Poll) error {
	poll.UpdatedAt = time.Now()

	filter := bson.M{"_id": poll.ID}
	update := bson.M{"$set": poll}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Vote records a vote on a poll
func (r *PollRepository) Vote(ctx context.Context, vote *model.PollVote) error {
	// Check if user already voted
	existing := r.voteCollection.FindOne(ctx, bson.M{
		"pollId": vote.PollID,
		"userId": vote.UserID,
	})
	if existing.Err() == nil {
		return fmt.Errorf("user has already voted on this poll")
	}

	// Get poll to validate options and check if active
	poll, err := r.FindByID(ctx, vote.PollID)
	if err != nil {
		return err
	}

	if !poll.IsActive {
		return fmt.Errorf("poll is not active")
	}

	// Check if poll has ended
	if poll.EndDate != nil && time.Now().After(*poll.EndDate) {
		return fmt.Errorf("poll has ended")
	}

	// Validate option IDs
	validOptions := make(map[string]bool)
	for _, opt := range poll.Options {
		validOptions[opt.ID] = true
	}

	for _, optID := range vote.OptionIDs {
		if !validOptions[optID] {
			return fmt.Errorf("invalid option ID: %s", optID)
		}
	}

	// Check multiple selection rules
	if !poll.IsMultiple && len(vote.OptionIDs) > 1 {
		return fmt.Errorf("poll does not allow multiple selections")
	}

	if poll.IsMultiple && poll.MaxSelections > 0 && len(vote.OptionIDs) > poll.MaxSelections {
		return fmt.Errorf("exceeded maximum selections: %d", poll.MaxSelections)
	}

	// Create vote
	vote.ID = primitive.NewObjectID()
	vote.CreatedAt = time.Now()

	if _, err := r.voteCollection.InsertOne(ctx, vote); err != nil {
		return err
	}

	// Update poll vote counts
	update := bson.M{
		"$inc": bson.M{"totalVotes": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	// Increment vote count for each selected option
	for _, optID := range vote.OptionIDs {
		for i, opt := range poll.Options {
			if opt.ID == optID {
				fieldName := fmt.Sprintf("options.%d.voteCount", i)
				update["$inc"].(bson.M)[fieldName] = 1
				break
			}
		}
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": vote.PollID}, update)
	return err
}

// HasUserVoted checks if a user has already voted on a poll
func (r *PollRepository) HasUserVoted(ctx context.Context, pollID primitive.ObjectID, userID string) (bool, error) {
	count, err := r.voteCollection.CountDocuments(ctx, bson.M{
		"pollId": pollID,
		"userId": userID,
	})
	return count > 0, err
}

// GetUserVote gets a user's vote on a poll
func (r *PollRepository) GetUserVote(ctx context.Context, pollID primitive.ObjectID, userID string) (*model.PollVote, error) {
	var vote model.PollVote
	err := r.voteCollection.FindOne(ctx, bson.M{
		"pollId": pollID,
		"userId": userID,
	}).Decode(&vote)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &vote, nil
}

// GetPollResults gets detailed results of a poll
func (r *PollRepository) GetPollResults(ctx context.Context, pollID primitive.ObjectID) (*model.Poll, error) {
	return r.FindByID(ctx, pollID)
}

// SetPollStatus activates or deactivates a poll
func (r *PollRepository) SetPollStatus(ctx context.Context, pollID primitive.ObjectID, isActive bool) error {
	update := bson.M{
		"$set": bson.M{
			"isActive":  isActive,
			"updatedAt": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": pollID}, update)
	return err
}

// calculatePercentages calculates vote percentages for poll options
func (r *PollRepository) calculatePercentages(poll *model.Poll) {
	if poll.TotalVotes == 0 {
		for i := range poll.Options {
			poll.Options[i].Percentage = 0
		}
		return
	}

	for i := range poll.Options {
		poll.Options[i].Percentage = float64(poll.Options[i].VoteCount) / float64(poll.TotalVotes) * 100
	}
}

// DeleteByArticleID deletes poll associated with an article
func (r *PollRepository) DeleteByArticleID(ctx context.Context, articleID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"articleId": articleID})
	return err
}
