package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PermissionGroup represents a group of permissions for managing content across multiple categories
type PermissionGroup struct {
	ID          primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name        string               `json:"name" bson:"name"`
	Description string               `json:"description" bson:"description"`
	CategoryIDs []primitive.ObjectID `json:"categoryIds" bson:"categoryIds"` // Categories this group has access to
	UserIDs     []string             `json:"userIds" bson:"userIds"`         // Users in this group
	GroupIDs    []string             `json:"groupIds" bson:"groupIds"`       // External group IDs (e.g., from LDAP, AD)
	Role        Role                 `json:"role" bson:"role"`               // Role assigned to this group
	CreatedAt   time.Time            `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   string               `json:"createdBy" bson:"createdBy"`
}
