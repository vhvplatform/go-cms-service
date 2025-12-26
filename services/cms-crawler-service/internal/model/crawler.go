package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CrawlerCampaign represents a campaign for crawling articles
type CrawlerCampaign struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID    string             `bson:"tenant_id" json:"tenantId"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	IsActive    bool               `bson:"is_active" json:"isActive"`
	Schedule    string             `bson:"schedule" json:"schedule"` // Cron expression
	
	// Source IDs for this campaign
	SourceIDs   []primitive.ObjectID `bson:"source_ids" json:"sourceIds"`
	
	// Auto-approval settings
	AutoApprove bool `bson:"auto_approve" json:"autoApprove"`
	
	// Retention settings
	RetentionDays int `bson:"retention_days" json:"retentionDays"` // Days to keep crawled articles
	
	CreatedBy   string    `bson:"created_by" json:"createdBy"`
	CreatedAt   time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updatedAt"`
	LastRunAt   time.Time `bson:"last_run_at,omitempty" json:"lastRunAt,omitempty"`
}

// CrawlerSource represents a source configuration for crawling
type CrawlerSource struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID    string             `bson:"tenant_id" json:"tenantId"`
	Name        string             `bson:"name" json:"name"`
	URL         string             `bson:"url" json:"url"`
	Type        string             `bson:"type" json:"type"` // rss, html, api
	
	// Extraction configuration
	ExtractionConfig ExtractionConfig `bson:"extraction_config" json:"extractionConfig"`
	
	// Anti-crawler bypass
	UserAgents     []string          `bson:"user_agents,omitempty" json:"userAgents,omitempty"`
	Headers        map[string]string `bson:"headers,omitempty" json:"headers,omitempty"`
	UseProxy       bool              `bson:"use_proxy" json:"useProxy"`
	ProxyURL       string            `bson:"proxy_url,omitempty" json:"proxyUrl,omitempty"`
	DelayMs        int               `bson:"delay_ms" json:"delayMs"` // Delay between requests
	
	// Auto-approval for this source
	AutoApprove bool `bson:"auto_approve" json:"autoApprove"`
	
	// Tracking
	IsActive      bool      `bson:"is_active" json:"isActive"`
	LastCrawledAt time.Time `bson:"last_crawled_at,omitempty" json:"lastCrawledAt,omitempty"`
	TotalCrawled  int       `bson:"total_crawled" json:"totalCrawled"`
	TotalErrors   int       `bson:"total_errors" json:"totalErrors"`
	
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

// ExtractionConfig defines how to extract content from a page
type ExtractionConfig struct {
	TitleSelector       string            `bson:"title_selector" json:"titleSelector"`             // CSS selector
	TitleXPath          string            `bson:"title_xpath,omitempty" json:"titleXPath,omitempty"`
	ContentSelector     string            `bson:"content_selector" json:"contentSelector"`
	ContentXPath        string            `bson:"content_xpath,omitempty" json:"contentXPath,omitempty"`
	ImageSelector       string            `bson:"image_selector,omitempty" json:"imageSelector,omitempty"`
	ImageXPath          string            `bson:"image_xpath,omitempty" json:"imageXPath,omitempty"`
	AuthorSelector      string            `bson:"author_selector,omitempty" json:"authorSelector,omitempty"`
	DateSelector        string            `bson:"date_selector,omitempty" json:"dateSelector,omitempty"`
	DateFormat          string            `bson:"date_format,omitempty" json:"dateFormat,omitempty"`
	TagsSelector        string            `bson:"tags_selector,omitempty" json:"tagsSelector,omitempty"`
	RemoveSelectors     []string          `bson:"remove_selectors,omitempty" json:"removeSelectors,omitempty"` // Elements to remove
	AttributeMapping    map[string]string `bson:"attribute_mapping,omitempty" json:"attributeMapping,omitempty"`
	UseReadability      bool              `bson:"use_readability" json:"useReadability"` // Use readability algorithm
}

// CrawlerArticle represents a crawled article
type CrawlerArticle struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID       string             `bson:"tenant_id" json:"tenantId"`
	CampaignID     primitive.ObjectID `bson:"campaign_id" json:"campaignId"`
	SourceID       primitive.ObjectID `bson:"source_id" json:"sourceId"`
	SourceURL      string             `bson:"source_url" json:"sourceUrl"`
	
	// Article data
	Title          string   `bson:"title" json:"title"`
	Content        string   `bson:"content" json:"content"`
	Summary        string   `bson:"summary,omitempty" json:"summary,omitempty"`
	Author         string   `bson:"author,omitempty" json:"author,omitempty"`
	ImageURL       string   `bson:"image_url,omitempty" json:"imageUrl,omitempty"`
	Tags           []string `bson:"tags,omitempty" json:"tags,omitempty"`
	PublishedAt    time.Time `bson:"published_at,omitempty" json:"publishedAt,omitempty"`
	
	// Metadata
	RawHTML        string            `bson:"raw_html,omitempty" json:"rawHtml,omitempty"`
	Metadata       map[string]string `bson:"metadata,omitempty" json:"metadata,omitempty"`
	
	// Status
	Status         string             `bson:"status" json:"status"` // pending, approved, rejected, converted
	ApprovedBy     string             `bson:"approved_by,omitempty" json:"approvedBy,omitempty"`
	ApprovedAt     time.Time          `bson:"approved_at,omitempty" json:"approvedAt,omitempty"`
	RejectedReason string             `bson:"rejected_reason,omitempty" json:"rejectedReason,omitempty"`
	ConvertedToID  primitive.ObjectID `bson:"converted_to_id,omitempty" json:"convertedToId,omitempty"` // Article ID after conversion
	
	// Similarity grouping
	SimilarityGroupID primitive.ObjectID `bson:"similarity_group_id,omitempty" json:"similarityGroupId,omitempty"`
	ContentHash       string             `bson:"content_hash" json:"contentHash"` // For duplicate detection
	
	// Media download tracking
	MediaDownloaded bool `bson:"media_downloaded" json:"mediaDownloaded"`
	
	CrawledAt time.Time `bson:"crawled_at" json:"crawledAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

// ContentSimilarityGroup represents a group of similar articles
type ContentSimilarityGroup struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	TenantID      string               `bson:"tenant_id" json:"tenantId"`
	ArticleIDs    []primitive.ObjectID `bson:"article_ids" json:"articleIds"`
	Representative primitive.ObjectID  `bson:"representative" json:"representative"` // Main article ID
	Title         string               `bson:"title" json:"title"`
	Similarity    float64              `bson:"similarity" json:"similarity"` // 0-1
	CreatedAt     time.Time            `bson:"created_at" json:"createdAt"`
	UpdatedAt     time.Time            `bson:"updated_at" json:"updatedAt"`
}

// CrawlerStats represents crawling statistics
type CrawlerStats struct {
	TenantID       string    `json:"tenantId"`
	TotalCrawled   int       `json:"totalCrawled"`
	PendingReview  int       `json:"pendingReview"`
	Approved       int       `json:"approved"`
	Rejected       int       `json:"rejected"`
	Converted      int       `json:"converted"`
	SimilarGroups  int       `json:"similarGroups"`
	LastCrawledAt  time.Time `json:"lastCrawledAt"`
}
