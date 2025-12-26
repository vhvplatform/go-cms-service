package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RejectionNote represents a note/comment in the rejection conversation thread
type RejectionNote struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ArticleID  primitive.ObjectID `json:"articleId" bson:"articleId"`
	UserID     string             `json:"userId" bson:"userId"`
	UserName   string             `json:"userName" bson:"userName"`
	UserRole   Role               `json:"userRole" bson:"userRole"`
	Note       string             `json:"note" bson:"note"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
	IsResolved bool               `json:"isResolved" bson:"isResolved"`
	ParentID   *primitive.ObjectID `json:"parentId,omitempty" bson:"parentId,omitempty"` // For threaded conversations
}
