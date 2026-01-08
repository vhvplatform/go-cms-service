package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Poll represents a poll/survey within an article
type Poll struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ArticleID     primitive.ObjectID `json:"articleId" bson:"articleId"`
	Question      string             `json:"question" bson:"question"`
	Options       []PollOption       `json:"options" bson:"options"`
	IsActive      bool               `json:"isActive" bson:"isActive"`
	IsMultiple    bool               `json:"isMultiple" bson:"isMultiple"`       // Allow multiple selections
	MaxSelections int                `json:"maxSelections" bson:"maxSelections"` // Max selections if multiple
	TotalVotes    int                `json:"totalVotes" bson:"totalVotes"`
	StartDate     time.Time          `json:"startDate" bson:"startDate"`
	EndDate       *time.Time         `json:"endDate,omitempty" bson:"endDate,omitempty"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy     string             `json:"createdBy" bson:"createdBy"`
}

// PollOption represents an option in a poll
type PollOption struct {
	ID         string  `json:"id" bson:"id"`
	Text       string  `json:"text" bson:"text"`
	VoteCount  int     `json:"voteCount" bson:"voteCount"`
	Percentage float64 `json:"percentage" bson:"-"` // Calculated field, not stored
}

// PollVote represents a user's vote on a poll
type PollVote struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PollID    primitive.ObjectID `json:"pollId" bson:"pollId"`
	UserID    string             `json:"userId" bson:"userId"`
	OptionIDs []string           `json:"optionIds" bson:"optionIds"` // Support multiple selections
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	IPAddress string             `json:"ipAddress,omitempty" bson:"ipAddress,omitempty"` // For tracking
}
