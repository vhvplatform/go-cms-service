package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArticleStatistics represents aggregated statistics for articles
type ArticleStatistics struct {
	TotalArticles      int64                       `json:"totalArticles"`
	PublishedArticles  int64                       `json:"publishedArticles"`
	DraftArticles      int64                       `json:"draftArticles"`
	PendingArticles    int64                       `json:"pendingArticles"`
	ArchivedArticles   int64                       `json:"archivedArticles"`
	TotalViews         int64                       `json:"totalViews"`
	ArticlesByType     map[string]int64            `json:"articlesByType"`
	ArticlesByCategory map[string]int64            `json:"articlesByCategory"`
	ArticlesByAuthor   map[string]int64            `json:"articlesByAuthor"`
	RecentArticles     []*Article                  `json:"recentArticles,omitempty"`
	TopViewedArticles  []*ArticleViewSummary       `json:"topViewedArticles"`
	Period             StatisticsPeriod            `json:"period"`
	GeneratedAt        time.Time                   `json:"generatedAt"`
}

// ArticleViewSummary represents a summary of article views
type ArticleViewSummary struct {
	ArticleID   primitive.ObjectID `json:"articleId"`
	Title       string             `json:"title"`
	Slug        string             `json:"slug"`
	ArticleType ArticleType        `json:"articleType"`
	ViewCount   int64              `json:"viewCount"`
	CategoryID  primitive.ObjectID `json:"categoryId"`
}

// CategoryStatistics represents statistics for a category
type CategoryStatistics struct {
	CategoryID      primitive.ObjectID `json:"categoryId"`
	CategoryName    string             `json:"categoryName"`
	TotalArticles   int64              `json:"totalArticles"`
	PublishedCount  int64              `json:"publishedCount"`
	TotalViews      int64              `json:"totalViews"`
	AverageViews    float64            `json:"averageViews"`
	LastUpdated     time.Time          `json:"lastUpdated"`
}

// AuthorStatistics represents statistics for an author
type AuthorStatistics struct {
	AuthorID        string    `json:"authorId"`
	AuthorName      string    `json:"authorName"`
	TotalArticles   int64     `json:"totalArticles"`
	PublishedCount  int64     `json:"publishedCount"`
	DraftCount      int64     `json:"draftCount"`
	TotalViews      int64     `json:"totalViews"`
	AverageViews    float64   `json:"averageViews"`
	LastPublished   time.Time `json:"lastPublished"`
}

// ViewTrendData represents view trends over time
type ViewTrendData struct {
	Date      time.Time `json:"date"`
	ViewCount int64     `json:"viewCount"`
	Articles  int64     `json:"articles"`
}

// StatisticsPeriod represents a time period for statistics
type StatisticsPeriod struct {
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Type      string    `json:"type"` // daily, weekly, monthly, yearly, custom
}

// TenantStatistics represents statistics for a tenant
type TenantStatistics struct {
	TenantID           primitive.ObjectID `json:"tenantId"`
	TenantName         string             `json:"tenantName"`
	TotalArticles      int64              `json:"totalArticles"`
	TotalViews         int64              `json:"totalViews"`
	TotalAuthors       int64              `json:"totalAuthors"`
	TotalCategories    int64              `json:"totalCategories"`
	ArticlesByType     map[string]int64   `json:"articlesByType"`
	MostActiveAuthors  []AuthorStatistics `json:"mostActiveAuthors"`
	TopCategories      []CategoryStatistics `json:"topCategories"`
}

// PerformanceMetrics represents performance metrics
type PerformanceMetrics struct {
	AveragePublishTime float64 `json:"averagePublishTime"` // Average time from draft to published (hours)
	AverageReviewTime  float64 `json:"averageReviewTime"`  // Average time in review (hours)
	ArticlesPerDay     float64 `json:"articlesPerDay"`
	ViewsPerDay        float64 `json:"viewsPerDay"`
	Period             StatisticsPeriod `json:"period"`
}
