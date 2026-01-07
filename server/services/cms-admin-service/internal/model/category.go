package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CategoryType represents the type of category
type CategoryType string

const (
	CategoryTypeArticle CategoryType = "Article"
	CategoryTypeLink    CategoryType = "Link"
)

// Category represents a category in the tree structure
type Category struct {
	ID           primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	Name         string              `json:"name" bson:"name"`
	Slug         string              `json:"slug" bson:"slug"`
	Description  string              `json:"description" bson:"description"`
	CategoryType CategoryType        `json:"categoryType" bson:"categoryType"`
	ArticleType  ArticleType         `json:"articleType,omitempty" bson:"articleType,omitempty"`   // Only for CategoryType == Article
	CategoryLink string              `json:"categoryLink,omitempty" bson:"categoryLink,omitempty"` // Only for CategoryType == Link
	ParentID     *primitive.ObjectID `json:"parentId,omitempty" bson:"parentId,omitempty"`
	Ordering     int                 `json:"ordering" bson:"ordering"`
	CreatedAt    time.Time           `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time           `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    string              `json:"createdBy" bson:"createdBy"`
}
