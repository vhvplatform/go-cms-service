package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/cache"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PublicArticleService handles article business logic for public/user-facing APIs with caching
type PublicArticleService struct {
	repo      *repository.ArticleRepository
	cache     cache.Cache
	cacheTTL  time.Duration
	viewQueue ViewQueue
}

// NewPublicArticleService creates a new public article service
func NewPublicArticleService(
	repo *repository.ArticleRepository,
	cache cache.Cache,
	cacheTTL time.Duration,
	viewQueue ViewQueue,
) *PublicArticleService {
	return &PublicArticleService{
		repo:      repo,
		cache:     cache,
		cacheTTL:  cacheTTL,
		viewQueue: viewQueue,
	}
}

// GetArticleByID gets a published article by ID with caching
func (s *PublicArticleService) GetArticleByID(ctx context.Context, id primitive.ObjectID) (*model.Article, error) {
	cacheKey := fmt.Sprintf("article:public:id:%s", id.Hex())

	// Try to get from cache
	var article model.Article
	err := s.cache.Get(ctx, cacheKey, &article)
	if err == nil {
		log.Printf("Cache hit for article ID: %s", id.Hex())
		return &article, nil
	}

	// Cache miss - get from database
	log.Printf("Cache miss for article ID: %s", id.Hex())
	dbArticle, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only return published articles that are not expired
	if !s.isPubliclyAccessible(dbArticle) {
		return nil, fmt.Errorf("article not found or not accessible")
	}

	// Store in cache
	if err := s.cache.Set(ctx, cacheKey, dbArticle, s.cacheTTL); err != nil {
		log.Printf("Failed to cache article: %v", err)
	}

	return dbArticle, nil
}

// GetArticleBySlug gets a published article by slug with caching
func (s *PublicArticleService) GetArticleBySlug(ctx context.Context, slug string) (*model.Article, error) {
	cacheKey := fmt.Sprintf("article:public:slug:%s", slug)

	// Try to get from cache
	var article model.Article
	err := s.cache.Get(ctx, cacheKey, &article)
	if err == nil {
		log.Printf("Cache hit for article slug: %s", slug)
		return &article, nil
	}

	// Cache miss - get from database
	log.Printf("Cache miss for article slug: %s", slug)
	dbArticle, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Only return published articles that are not expired
	if !s.isPubliclyAccessible(dbArticle) {
		return nil, fmt.Errorf("article not found or not accessible")
	}

	// Store in cache
	if err := s.cache.Set(ctx, cacheKey, dbArticle, s.cacheTTL); err != nil {
		log.Printf("Failed to cache article: %v", err)
	}

	return dbArticle, nil
}

// ListPublicArticles lists published articles with caching
func (s *PublicArticleService) ListPublicArticles(ctx context.Context, filter map[string]interface{}, page, limit int, sort map[string]int) ([]*model.Article, int64, error) {
	// Create cache key based on filter, page, limit, sort
	cacheKey := fmt.Sprintf("articles:public:list:%v:%d:%d:%v", filter, page, limit, sort)

	// Try to get from cache
	var cachedResult struct {
		Articles []*model.Article `json:"articles"`
		Total    int64            `json:"total"`
	}
	err := s.cache.Get(ctx, cacheKey, &cachedResult)
	if err == nil {
		log.Printf("Cache hit for article list")
		return cachedResult.Articles, cachedResult.Total, nil
	}

	// Cache miss - get from database
	log.Printf("Cache miss for article list")

	// Force published status filter
	filter["status"] = model.ArticleStatusPublished

	// Add publish date filter
	now := time.Now()
	if filter["publishAt"] == nil {
		filter["publishAt"] = map[string]interface{}{"$lte": now}
	}

	articles, total, err := s.repo.FindAll(ctx, filter, page, limit, sort)
	if err != nil {
		return nil, 0, err
	}

	// Filter out expired articles
	accessibleArticles := make([]*model.Article, 0, len(articles))
	for _, article := range articles {
		if s.isPubliclyAccessible(article) {
			accessibleArticles = append(accessibleArticles, article)
		}
	}

	// Store in cache
	cachedResult.Articles = accessibleArticles
	cachedResult.Total = int64(len(accessibleArticles))
	if err := s.cache.Set(ctx, cacheKey, cachedResult, s.cacheTTL); err != nil {
		log.Printf("Failed to cache article list: %v", err)
	}

	return accessibleArticles, total, nil
}

// IncrementViewCount increments the view count for an article (no cache)
func (s *PublicArticleService) IncrementViewCount(ctx context.Context, id primitive.ObjectID) error {
	// Enqueue view event for asynchronous processing
	if s.viewQueue != nil {
		return s.viewQueue.Enqueue(id)
	}

	// Fallback to synchronous processing if queue not available
	return s.repo.IncrementViewCount(ctx, id)
}

// InvalidateArticleCache invalidates cache for a specific article (called when admin updates)
func (s *PublicArticleService) InvalidateArticleCache(ctx context.Context, article *model.Article) error {
	log.Printf("Invalidating cache for article: %s", article.ID.Hex())

	keys := []string{
		fmt.Sprintf("article:public:id:%s", article.ID.Hex()),
		fmt.Sprintf("article:public:slug:%s", article.Slug),
	}

	if err := s.cache.Delete(ctx, keys...); err != nil {
		log.Printf("Failed to invalidate article cache: %v", err)
		return err
	}

	// Invalidate list caches
	if err := s.cache.DeletePattern(ctx, "articles:public:list:*"); err != nil {
		log.Printf("Failed to invalidate article list cache: %v", err)
		return err
	}

	log.Printf("Successfully invalidated cache for article: %s", article.ID.Hex())
	return nil
}

// InvalidateAllArticleCaches invalidates all article caches
func (s *PublicArticleService) InvalidateAllArticleCaches(ctx context.Context) error {
	log.Println("Invalidating all article caches")

	patterns := []string{
		"article:public:*",
		"articles:public:*",
	}

	for _, pattern := range patterns {
		if err := s.cache.DeletePattern(ctx, pattern); err != nil {
			log.Printf("Failed to invalidate cache pattern %s: %v", pattern, err)
			return err
		}
	}

	log.Println("Successfully invalidated all article caches")
	return nil
}

// isPubliclyAccessible checks if an article is publicly accessible
func (s *PublicArticleService) isPubliclyAccessible(article *model.Article) bool {
	now := time.Now()

	// Must be published
	if article.Status != model.ArticleStatusPublished {
		return false
	}

	// Must be past publish date
	if article.PublishAt.After(now) {
		return false
	}

	// Must not be expired
	if article.ExpiredAt != nil && now.After(*article.ExpiredAt) {
		return false
	}

	return true
}
