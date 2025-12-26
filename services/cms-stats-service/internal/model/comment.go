package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentStatus represents the moderation status of a comment
type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"
	CommentStatusApproved CommentStatus = "approved"
	CommentStatusRejected CommentStatus = "rejected"
	CommentStatusDeleted  CommentStatus = "deleted"
)

// Comment represents a user comment on an article
type Comment struct {
	ID             primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	ArticleID      primitive.ObjectID  `json:"articleId" bson:"articleId"`
	ParentID       *primitive.ObjectID `json:"parentId,omitempty" bson:"parentId,omitempty"` // For nested comments
	UserID         string              `json:"userId" bson:"userId"`
	UserName       string              `json:"userName" bson:"userName"`
	Content        string              `json:"content" bson:"content"`
	Status         CommentStatus       `json:"status" bson:"status"`
	LikeCount      int                 `json:"likeCount" bson:"likeCount"`
	ReplyCount     int                 `json:"replyCount" bson:"replyCount"`
	Level          int                 `json:"level" bson:"level"` // Nesting level: 0, 1, 2 (max 3 levels)
	IsReported     bool                `json:"isReported" bson:"isReported"`
	ReportCount    int                 `json:"reportCount" bson:"reportCount"`
	CreatedAt      time.Time           `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time           `json:"updatedAt" bson:"updatedAt"`
	ModeratedBy    string              `json:"moderatedBy,omitempty" bson:"moderatedBy,omitempty"`
	ModeratedAt    *time.Time          `json:"moderatedAt,omitempty" bson:"moderatedAt,omitempty"`
	ModerationNote string              `json:"moderationNote,omitempty" bson:"moderationNote,omitempty"`
}

// CommentLike represents a user's like on a comment
type CommentLike struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CommentID primitive.ObjectID `json:"commentId" bson:"commentId"`
	UserID    string             `json:"userId" bson:"userId"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

// CommentReport represents a violation report on a comment
type CommentReport struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CommentID  primitive.ObjectID `json:"commentId" bson:"commentId"`
	ReporterID string             `json:"reporterId" bson:"reporterId"`
	Reason     string             `json:"reason" bson:"reason"`
	Status     string             `json:"status" bson:"status"` // pending, reviewed, dismissed
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	ReviewedBy string             `json:"reviewedBy,omitempty" bson:"reviewedBy,omitempty"`
	ReviewedAt *time.Time         `json:"reviewedAt,omitempty" bson:"reviewedAt,omitempty"`
}

// UserRateLimit tracks comment rate limiting per user
type UserRateLimit struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       string             `json:"userId" bson:"userId"`
	CommentCount int                `json:"commentCount" bson:"commentCount"`
	WindowStart  time.Time          `json:"windowStart" bson:"windowStart"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// FavoriteArticle represents a user's favorite/saved article
type FavoriteArticle struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    string             `json:"userId" bson:"userId"`
	ArticleID primitive.ObjectID `json:"articleId" bson:"articleId"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}
