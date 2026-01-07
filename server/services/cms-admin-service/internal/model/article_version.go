package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArticleVersion represents a historical version of an article
type ArticleVersion struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ArticleID   primitive.ObjectID `json:"articleId" bson:"articleId"`
	VersionNum  int                `json:"versionNum" bson:"versionNum"`
	Title       string             `json:"title" bson:"title"`
	Subtitle    string             `json:"subtitle" bson:"subtitle"`
	Slug        string             `json:"slug" bson:"slug"`
	ArticleType ArticleType        `json:"articleType" bson:"articleType"`
	CategoryID  primitive.ObjectID `json:"categoryId" bson:"categoryId"`
	Summary     string             `json:"summary" bson:"summary"`
	Content     string             `json:"content" bson:"content"`
	Author      Author             `json:"author" bson:"author"`
	Tags        []string           `json:"tags" bson:"tags"`
	SEO         SEO                `json:"seo" bson:"seo"`
	Status      ArticleStatus      `json:"status" bson:"status"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	CreatedBy   string             `json:"createdBy" bson:"createdBy"`

	// Store entire article snapshot for full restore capability
	FullSnapshot *Article `json:"fullSnapshot,omitempty" bson:"fullSnapshot,omitempty"`

	// Metadata about this version
	ChangeNote   string `json:"changeNote,omitempty" bson:"changeNote,omitempty"`
	RestoredFrom *int   `json:"restoredFrom,omitempty" bson:"restoredFrom,omitempty"` // Version number this was restored from
}
