package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActionType represents the type of action performed
type ActionType string

const (
	ActionTypeCreate    ActionType = "create"
	ActionTypeUpdate    ActionType = "update"
	ActionTypePublish   ActionType = "publish"
	ActionTypeApprove   ActionType = "approve"
	ActionTypeReject    ActionType = "reject"
	ActionTypeDelete    ActionType = "delete"
	ActionTypeRestore   ActionType = "restore"
	ActionTypeArchive   ActionType = "archive"
	ActionTypeUnarchive ActionType = "unarchive"
)

// ActionLog represents a log entry for actions performed on articles
type ActionLog struct {
	ID         primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	ArticleID  primitive.ObjectID     `json:"articleId" bson:"articleId"`
	ActionType ActionType             `json:"actionType" bson:"actionType"`
	UserID     string                 `json:"userId" bson:"userId"`
	UserName   string                 `json:"userName" bson:"userName"`
	UserRole   Role                   `json:"userRole" bson:"userRole"`
	Timestamp  time.Time              `json:"timestamp" bson:"timestamp"`
	Changes    map[string]interface{} `json:"changes,omitempty" bson:"changes,omitempty"`
	Note       string                 `json:"note,omitempty" bson:"note,omitempty"`
	IPAddress  string                 `json:"ipAddress,omitempty" bson:"ipAddress,omitempty"`
	UserAgent  string                 `json:"userAgent,omitempty" bson:"userAgent,omitempty"`
	OldStatus  ArticleStatus          `json:"oldStatus,omitempty" bson:"oldStatus,omitempty"`
	NewStatus  ArticleStatus          `json:"newStatus,omitempty" bson:"newStatus,omitempty"`
	VersionID  *primitive.ObjectID    `json:"versionId,omitempty" bson:"versionId,omitempty"`
}
