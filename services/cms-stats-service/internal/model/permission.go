package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role represents user roles in the system
type Role string

const (
	RoleWriter    Role = "writer"
	RoleEditor    Role = "editor"
	RoleModerator Role = "moderator"
)

// ResourceType represents the type of resource for permissions
type ResourceType string

const (
	ResourceTypeCategory    ResourceType = "category"
	ResourceTypeEventStream ResourceType = "event_stream"
)

// Permission represents access control for categories and event lines
type Permission struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ResourceType ResourceType       `json:"resourceType" bson:"resourceType"`
	ResourceID   primitive.ObjectID `json:"resourceId" bson:"resourceId"`
	Role         Role               `json:"role" bson:"role"`
	UserID       string             `json:"userId,omitempty" bson:"userId,omitempty"`
	GroupID      string             `json:"groupId,omitempty" bson:"groupId,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    string             `json:"createdBy" bson:"createdBy"`
}
