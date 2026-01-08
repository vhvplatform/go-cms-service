package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArticleType represents the type of article
type ArticleType string

const (
	ArticleTypeNews          ArticleType = "News"
	ArticleTypeVideo         ArticleType = "Video"
	ArticleTypePhotoGallery  ArticleType = "PhotoGallery"
	ArticleTypeLegalDocument ArticleType = "LegalDocument"
	ArticleTypeStaffProfile  ArticleType = "StaffProfile"
	ArticleTypeJob           ArticleType = "Job"
	ArticleTypeProcedure     ArticleType = "Procedure"
	ArticleTypeDownload      ArticleType = "Download"
	ArticleTypePodcast       ArticleType = "Podcast"
	ArticleTypeEventInfo     ArticleType = "EventInfo"
	ArticleTypeInfographic   ArticleType = "Infographic"
	ArticleTypeDestination   ArticleType = "Destination"
	ArticleTypePartner       ArticleType = "Partner"
	ArticleTypePDF           ArticleType = "PDF"
)

// ArticleStatus represents the publication status of an article
type ArticleStatus string

const (
	ArticleStatusDraft         ArticleStatus = "draft"
	ArticleStatusPendingReview ArticleStatus = "pending_review"
	ArticleStatusPublished     ArticleStatus = "published"
	ArticleStatusArchived      ArticleStatus = "archived"
	ArticleStatusDeleted       ArticleStatus = "deleted"
)

// Author represents the author information
type Author struct {
	ID         string `json:"id" bson:"id"`
	Name       string `json:"name" bson:"name"`
	ProfileURL string `json:"profileUrl" bson:"profileUrl"`
}

// Source represents the source information
type Source struct {
	Name string `json:"name" bson:"name"`
	URL  string `json:"url" bson:"url"`
}

// SEO represents SEO metadata
type SEO struct {
	Title       string   `json:"title" bson:"title"`
	Description string   `json:"description" bson:"description"`
	Keywords    []string `json:"keywords" bson:"keywords"`
	Canonical   string   `json:"canonical" bson:"canonical"`
	NoIndex     bool     `json:"noIndex" bson:"noIndex"`
}

// Attachment represents a file attachment
type Attachment struct {
	Type     string                 `json:"type" bson:"type"`
	URL      string                 `json:"url" bson:"url"`
	Metadata map[string]interface{} `json:"metadata" bson:"metadata"`
}

// ContentBlock represents a content block for easier counting
type ContentBlock struct {
	Type    string                 `json:"type" bson:"type"` // text, image, video, etc.
	Content string                 `json:"content" bson:"content"`
	Data    map[string]interface{} `json:"data" bson:"data"`
}

// GalleryImage represents an image in a photo gallery
type GalleryImage struct {
	URL     string                 `json:"url" bson:"url"`
	Caption string                 `json:"caption" bson:"caption"`
	Meta    map[string]interface{} `json:"meta" bson:"meta"`
}

// ReadingStats represents embedded reading statistics
type ReadingStats struct {
	DailyViews map[string]int `json:"dailyViews" bson:"dailyViews"` // date -> count
	LastUpdate time.Time      `json:"lastUpdate" bson:"lastUpdate"`
}

// AccessControl represents access control settings for an article
type AccessControl struct {
	IsPublic         bool     `json:"isPublic" bson:"isPublic"`                 // If false, requires authentication
	RequiresLogin    bool     `json:"requiresLogin" bson:"requiresLogin"`       // Requires user to be logged in
	AllowedUserIDs   []string `json:"allowedUserIds" bson:"allowedUserIds"`     // Specific users who can view
	AllowedGroupIDs  []string `json:"allowedGroupIds" bson:"allowedGroupIds"`   // Groups that can view
	AllowedRoles     []Role   `json:"allowedRoles" bson:"allowedRoles"`         // Roles that can view
	DeniedUserIDs    []string `json:"deniedUserIds" bson:"deniedUserIds"`       // Explicitly denied users
	RequiresPurchase bool     `json:"requiresPurchase" bson:"requiresPurchase"` // Requires payment/subscription
	IsPremium        bool     `json:"isPremium" bson:"isPremium"`               // Premium content flag
}

// CommentConfig represents comment configuration for an article
type CommentConfig struct {
	Enabled         bool `json:"enabled" bson:"enabled"`                 // Comments enabled
	RequireApproval bool `json:"requireApproval" bson:"requireApproval"` // Require moderation
	AllowAnonymous  bool `json:"allowAnonymous" bson:"allowAnonymous"`   // Allow anonymous comments
	AllowNested     bool `json:"allowNested" bson:"allowNested"`         // Allow nested replies
	MaxNestingLevel int  `json:"maxNestingLevel" bson:"maxNestingLevel"` // Max nesting level (default 3)
	AutoCloseAfter  int  `json:"autoCloseAfter" bson:"autoCloseAfter"`   // Auto-close after N days (0 = never)
}

// Article represents the main article document
type Article struct {
	ID            primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	TenantID      primitive.ObjectID     `json:"tenantId" bson:"tenantId"` // Tenant ownership
	Title         string                 `json:"title" bson:"title"`
	Subtitle      string                 `json:"subtitle" bson:"subtitle"`
	Slug          string                 `json:"slug" bson:"slug"`
	ArticleType   ArticleType            `json:"articleType" bson:"articleType"`
	CategoryID    primitive.ObjectID     `json:"categoryId" bson:"categoryId"`
	EventStreamID *primitive.ObjectID    `json:"eventStreamId,omitempty" bson:"eventStreamId,omitempty"`
	Summary       string                 `json:"summary" bson:"summary"`
	Content       string                 `json:"content" bson:"content"`
	ContentBlocks []ContentBlock         `json:"contentBlocks" bson:"contentBlocks"`
	Author        Author                 `json:"author" bson:"author"`
	Contributors  []Author               `json:"contributors" bson:"contributors"`
	Source        *Source                `json:"source,omitempty" bson:"source,omitempty"`
	Tags          []string               `json:"tags" bson:"tags"`
	SEO           SEO                    `json:"seo" bson:"seo"`
	CustomFields  map[string]interface{} `json:"customFields" bson:"customFields"`
	Attachments   []Attachment           `json:"attachments" bson:"attachments"`
	Featured      bool                   `json:"featured" bson:"featured"`
	Hot           bool                   `json:"hot" bson:"hot"`
	Ordering      int                    `json:"ordering" bson:"ordering"`
	PublishAt     time.Time              `json:"publishAt" bson:"publishAt"`
	CreatedAt     time.Time              `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt" bson:"updatedAt"`
	PublishedBy   string                 `json:"publishedBy" bson:"publishedBy"`
	CreatedBy     string                 `json:"createdBy" bson:"createdBy"`
	ExpiredAt     *time.Time             `json:"expiredAt,omitempty" bson:"expiredAt,omitempty"`
	Status        ArticleStatus          `json:"status" bson:"status"`
	IsCommentable bool                   `json:"isCommentable" bson:"isCommentable"` // Can users comment on this article
	CommentConfig CommentConfig          `json:"commentConfig" bson:"commentConfig"` // Comment configuration
	CharCount     int                    `json:"charCount" bson:"charCount"`
	ImageCount    int                    `json:"imageCount" bson:"imageCount"`
	ViewCount     int                    `json:"viewCount" bson:"viewCount"`
	ReadingStats  *ReadingStats          `json:"readingStats,omitempty" bson:"readingStats,omitempty"`
	AccessControl AccessControl          `json:"accessControl" bson:"accessControl"` // Access control settings

	// Version management
	CurrentVersion int `json:"currentVersion" bson:"currentVersion"` // Current version number

	// Related articles
	RelatedArticles []primitive.ObjectID `json:"relatedArticles,omitempty" bson:"relatedArticles,omitempty"`

	// Poll/Survey
	HasPoll bool                `json:"hasPoll" bson:"hasPoll"`                   // Indicates if article has a poll
	PollID  *primitive.ObjectID `json:"pollId,omitempty" bson:"pollId,omitempty"` // Reference to poll

	// Type-specific fields
	// Video
	VideoURL  string `json:"videoUrl,omitempty" bson:"videoUrl,omitempty"`
	Duration  int    `json:"duration,omitempty" bson:"duration,omitempty"` // in seconds
	Thumbnail string `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`

	// Photo Gallery
	Images        []GalleryImage `json:"images,omitempty" bson:"images,omitempty"`
	GalleryLayout string         `json:"galleryLayout,omitempty" bson:"galleryLayout,omitempty"`

	// Legal Document
	IssuedDate    *time.Time `json:"issuedDate,omitempty" bson:"issuedDate,omitempty"`
	LawNumber     string     `json:"lawNumber,omitempty" bson:"lawNumber,omitempty"`
	EffectiveDate *time.Time `json:"effectiveDate,omitempty" bson:"effectiveDate,omitempty"`
	PDFAttachment string     `json:"pdfAttachment,omitempty" bson:"pdfAttachment,omitempty"`

	// Podcast
	AudioURL      string `json:"audioUrl,omitempty" bson:"audioUrl,omitempty"`
	EpisodeNumber int    `json:"episodeNumber,omitempty" bson:"episodeNumber,omitempty"`

	// Download/PDF
	FileURL  string `json:"fileUrl,omitempty" bson:"fileUrl,omitempty"`
	FileSize int64  `json:"fileSize,omitempty" bson:"fileSize,omitempty"`
	MimeType string `json:"mimeType,omitempty" bson:"mimeType,omitempty"`

	// Event
	EventStart *time.Time `json:"eventStart,omitempty" bson:"eventStart,omitempty"`
	EventEnd   *time.Time `json:"eventEnd,omitempty" bson:"eventEnd,omitempty"`
	Venue      string     `json:"venue,omitempty" bson:"venue,omitempty"`
	Organizer  string     `json:"organizer,omitempty" bson:"organizer,omitempty"`
}

// ArticleView represents daily view statistics
type ArticleView struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ArticleID primitive.ObjectID `json:"articleId" bson:"articleId"`
	Date      time.Time          `json:"date" bson:"date"`
	Views     int                `json:"views" bson:"views"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}
