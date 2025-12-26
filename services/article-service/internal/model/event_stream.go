package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventStream represents a timeline of events
type EventStream struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Slug        string             `json:"slug" bson:"slug"`
	Description string             `json:"description" bson:"description"`
	StartAt     time.Time          `json:"startAt" bson:"startAt"`
	EndAt       *time.Time         `json:"endAt,omitempty" bson:"endAt,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   string             `json:"createdBy" bson:"createdBy"`
}
