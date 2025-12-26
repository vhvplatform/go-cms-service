package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/vhvplatform/go-cms-service/services/article-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/article-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ViewQueue interface for dependency injection
type ViewQueue interface {
	Enqueue(articleID primitive.ObjectID) error
}

// ArticleService handles article business logic
type ArticleService struct {
	repo           *repository.ArticleRepository
	permissionRepo *repository.PermissionRepository
	viewStatsRepo  *repository.ViewStatsRepository
	viewQueue      ViewQueue
}

// NewArticleService creates a new article service
func NewArticleService(
	repo *repository.ArticleRepository,
	permissionRepo *repository.PermissionRepository,
	viewStatsRepo *repository.ViewStatsRepository,
	viewQueue ViewQueue,
) *ArticleService {
	return &ArticleService{
		repo:           repo,
		permissionRepo: permissionRepo,
		viewStatsRepo:  viewStatsRepo,
		viewQueue:      viewQueue,
	}
}

// Create creates a new article
func (s *ArticleService) Create(ctx context.Context, article *model.Article, userID string) error {
	// Generate slug if not provided
	if article.Slug == "" {
		article.Slug = s.generateSlug(article.Title)
	}

	// Calculate character and image counts
	article.CharCount = s.repo.CalculateCharCount(article.Content)
	article.ImageCount = s.repo.CalculateImageCount(article.ContentBlocks)

	// Set defaults
	if article.Status == "" {
		article.Status = model.ArticleStatusDraft
	}
	article.CreatedBy = userID

	return s.repo.Create(ctx, article)
}

// FindByID finds an article by ID
func (s *ArticleService) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Article, error) {
	return s.repo.FindByID(ctx, id)
}

// FindBySlug finds an article by slug
func (s *ArticleService) FindBySlug(ctx context.Context, slug string) (*model.Article, error) {
	return s.repo.FindBySlug(ctx, slug)
}

// Update updates an article
func (s *ArticleService) Update(ctx context.Context, article *model.Article, userID string, userRole model.Role) error {
	// Get existing article to check status
	existing, err := s.repo.FindByID(ctx, article.ID)
	if err != nil {
		return err
	}

	// Check if article has been reviewed (moved beyond draft/pending_review)
	if existing.Status == model.ArticleStatusPublished || existing.Status == model.ArticleStatusArchived {
		// Only editors and moderators can edit reviewed articles
		if userRole != model.RoleEditor && userRole != model.RoleModerator {
			return fmt.Errorf("cannot edit article: article has been reviewed and published, only editors and moderators can make changes")
		}
	}

	// If article is published or archived, and user is only a writer, block the update
	if existing.CreatedBy == userID && userRole == model.RoleWriter {
		if existing.Status == model.ArticleStatusPublished || existing.Status == model.ArticleStatusArchived {
			return fmt.Errorf("cannot edit article: article has been reviewed, contact an editor or moderator for changes")
		}
	}

	// Recalculate counts
	article.CharCount = s.repo.CalculateCharCount(article.Content)
	article.ImageCount = s.repo.CalculateImageCount(article.ContentBlocks)

	return s.repo.Update(ctx, article)
}

// Delete soft deletes an article
func (s *ArticleService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

// FindAll finds articles with filters and pagination
func (s *ArticleService) FindAll(ctx context.Context, filter map[string]interface{}, page, limit int, sort map[string]int) ([]*model.Article, int64, error) {
	return s.repo.FindAll(ctx, filter, page, limit, sort)
}

// Publish publishes an article
func (s *ArticleService) Publish(ctx context.Context, id primitive.ObjectID, userID string, userRole model.Role) error {
	// Only editors and moderators can publish
	if userRole != model.RoleEditor && userRole != model.RoleModerator {
		return fmt.Errorf("insufficient permissions: only editors and moderators can publish articles")
	}

	article, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if article can be published
	if article.Status == model.ArticleStatusPublished {
		return fmt.Errorf("article is already published")
	}

	// Check publish date
	now := time.Now()
	if article.PublishAt.After(now) {
		return fmt.Errorf("article is scheduled for future publication")
	}

	return s.repo.UpdateStatus(ctx, id, model.ArticleStatusPublished, userID)
}

// UpdateStatus updates article status
func (s *ArticleService) UpdateStatus(ctx context.Context, id primitive.ObjectID, status model.ArticleStatus, userID string, userRole model.Role) error {
	// Get existing article
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Check permissions based on status transition
	if existing.Status == model.ArticleStatusPublished || existing.Status == model.ArticleStatusArchived {
		// Only editors and moderators can change status of published/archived articles
		if userRole != model.RoleEditor && userRole != model.RoleModerator {
			return fmt.Errorf("cannot change article status: article has been reviewed, only editors and moderators can change status")
		}
	}

	// Writers can only change status of their own draft/pending_review articles
	if userRole == model.RoleWriter {
		if existing.CreatedBy != userID {
			return fmt.Errorf("cannot change article status: you can only change status of your own articles")
		}
		
		// Writers cannot publish articles, only submit for review
		if status == model.ArticleStatusPublished {
			return fmt.Errorf("cannot publish article: only editors and moderators can publish articles")
		}
	}

	return s.repo.UpdateStatus(ctx, id, status, userID)
}

// Reorder updates article ordering
func (s *ArticleService) Reorder(ctx context.Context, articles []struct {
	ID       primitive.ObjectID
	Ordering int
}) error {
	for _, item := range articles {
		if err := s.repo.UpdateOrdering(ctx, item.ID, item.Ordering); err != nil {
			return err
		}
	}
	return nil
}

// SetFeatured sets the featured flag for an article
func (s *ArticleService) SetFeatured(ctx context.Context, id primitive.ObjectID, featured bool) error {
	article, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	article.Featured = featured
	return s.repo.Update(ctx, article)
}

// SetHot sets the hot flag for an article
func (s *ArticleService) SetHot(ctx context.Context, id primitive.ObjectID, hot bool) error {
	article, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	article.Hot = hot
	return s.repo.Update(ctx, article)
}

// IncrementViewCount increments the view count for an article using queue
func (s *ArticleService) IncrementViewCount(ctx context.Context, id primitive.ObjectID) error {
	// Enqueue view event for asynchronous processing
	if s.viewQueue != nil {
		return s.viewQueue.Enqueue(id)
	}
	
	// Fallback to synchronous processing if queue not available
	if err := s.repo.IncrementViewCount(ctx, id); err != nil {
		return err
	}
	return s.viewStatsRepo.RecordView(ctx, id, time.Now())
}

// GetArticleStats gets view statistics for an article
func (s *ArticleService) GetArticleStats(ctx context.Context, articleID primitive.ObjectID, startDate, endDate time.Time) ([]*model.ArticleView, error) {
	return s.viewStatsRepo.GetArticleStats(ctx, articleID, startDate, endDate)
}

// Search performs full-text search on articles
func (s *ArticleService) Search(ctx context.Context, query string, page, limit int) ([]*model.Article, int64, error) {
	filter := map[string]interface{}{
		"q":      query,
		"status": model.ArticleStatusPublished,
	}

	return s.repo.FindAll(ctx, filter, page, limit, nil)
}

// GetPublishableArticles gets articles ready to be published
func (s *ArticleService) GetPublishableArticles(ctx context.Context) ([]*model.Article, error) {
	return s.repo.FindArticlesToPublish(ctx)
}

// GetExpirableArticles gets articles ready to be expired
func (s *ArticleService) GetExpirableArticles(ctx context.Context) ([]*model.Article, error) {
	return s.repo.FindArticlesToExpire(ctx)
}

// CheckPermission checks if a user has permission to perform an action on an article
func (s *ArticleService) CheckPermission(ctx context.Context, userID string, categoryID primitive.ObjectID, requiredRole model.Role) (bool, error) {
	return s.permissionRepo.CheckPermission(ctx, userID, model.ResourceTypeCategory, categoryID, requiredRole)
}

// generateSlug generates a URL-friendly slug from a title
func (s *ArticleService) generateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Remove special characters
	reg := regexp.MustCompile("[^a-z0-9\\s-]")
	slug = reg.ReplaceAllString(slug, "")

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove multiple hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")

	// Add timestamp to ensure uniqueness
	timestamp := time.Now().Unix()
	slug = fmt.Sprintf("%s-%d", slug, timestamp)

	return slug
}
