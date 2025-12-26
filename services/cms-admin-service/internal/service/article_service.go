package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ViewQueue interface for dependency injection
type ViewQueue interface {
	Enqueue(articleID primitive.ObjectID) error
}

// ArticleService handles article business logic
type ArticleService struct {
	repo              *repository.ArticleRepository
	permissionRepo    *repository.PermissionRepository
	viewStatsRepo     *repository.ViewStatsRepository
	viewQueue         ViewQueue
	actionLogRepo     *repository.ActionLogRepository
	versionRepo       *repository.ArticleVersionRepository
	rejectionNoteRepo *repository.RejectionNoteRepository
	imageDownloader   *util.ImageDownloader
}

// NewArticleService creates a new article service
func NewArticleService(
	repo *repository.ArticleRepository,
	permissionRepo *repository.PermissionRepository,
	viewStatsRepo *repository.ViewStatsRepository,
	viewQueue ViewQueue,
	actionLogRepo *repository.ActionLogRepository,
	versionRepo *repository.ArticleVersionRepository,
	rejectionNoteRepo *repository.RejectionNoteRepository,
	imageDownloader *util.ImageDownloader,
) *ArticleService {
	return &ArticleService{
		repo:              repo,
		permissionRepo:    permissionRepo,
		viewStatsRepo:     viewStatsRepo,
		viewQueue:         viewQueue,
		actionLogRepo:     actionLogRepo,
		versionRepo:       versionRepo,
		rejectionNoteRepo: rejectionNoteRepo,
		imageDownloader:   imageDownloader,
	}
}

// Create creates a new article
func (s *ArticleService) Create(ctx context.Context, article *model.Article, userID string) error {
	// Generate slug if not provided
	if article.Slug == "" {
		article.Slug = s.generateSlug(article.Title)
	}

	// Process and download external images if image downloader is available
	if s.imageDownloader != nil && article.Content != "" {
		processedContent, _, err := s.imageDownloader.ProcessHTMLImages(article.Content)
		if err == nil {
			article.Content = processedContent
		}
	}

	// Calculate character and image counts
	article.CharCount = s.repo.CalculateCharCount(article.Content)
	article.ImageCount = s.repo.CalculateImageCount(article.ContentBlocks)

	// Set defaults
	if article.Status == "" {
		article.Status = model.ArticleStatusDraft
	}
	article.CreatedBy = userID
	article.CurrentVersion = 1

	// Create article
	if err := s.repo.Create(ctx, article); err != nil {
		return err
	}

	// Log action
	if s.actionLogRepo != nil {
		s.logAction(ctx, &model.ActionLog{
			ArticleID:  article.ID,
			ActionType: model.ActionTypeCreate,
			UserID:     userID,
			NewStatus:  article.Status,
		})
	}

	// Create initial version
	if s.versionRepo != nil {
		s.createVersion(ctx, article, userID, "Initial version")
	}

	return nil
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

	// Process and download external images if image downloader is available
	if s.imageDownloader != nil && article.Content != "" {
		processedContent, _, err := s.imageDownloader.ProcessHTMLImages(article.Content)
		if err == nil {
			article.Content = processedContent
		}
	}

	// Recalculate counts
	article.CharCount = s.repo.CalculateCharCount(article.Content)
	article.ImageCount = s.repo.CalculateImageCount(article.ContentBlocks)

	// Increment version
	article.CurrentVersion = existing.CurrentVersion + 1

	// Update article
	if err := s.repo.Update(ctx, article); err != nil {
		return err
	}

	// Log action
	if s.actionLogRepo != nil {
		s.logAction(ctx, &model.ActionLog{
			ArticleID:  article.ID,
			ActionType: model.ActionTypeUpdate,
			UserID:     userID,
			OldStatus:  existing.Status,
			NewStatus:  article.Status,
		})
	}

	// Create version snapshot
	if s.versionRepo != nil {
		s.createVersion(ctx, article, userID, "Article updated")
	}

	return nil
}

// Delete soft deletes an article
func (s *ArticleService) Delete(ctx context.Context, id primitive.ObjectID, userID string) error {
	// Get article for logging
	article, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	oldStatus := article.Status

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Log action
	if s.actionLogRepo != nil {
		s.logAction(ctx, &model.ActionLog{
			ArticleID:  id,
			ActionType: model.ActionTypeDelete,
			UserID:     userID,
			OldStatus:  oldStatus,
			NewStatus:  model.ArticleStatusDeleted,
		})
	}

	return nil
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

	oldStatus := article.Status

	// Check if article can be published
	if article.Status == model.ArticleStatusPublished {
		return fmt.Errorf("article is already published")
	}

	// Check publish date
	now := time.Now()
	if article.PublishAt.After(now) {
		return fmt.Errorf("article is scheduled for future publication")
	}

	if err := s.repo.UpdateStatus(ctx, id, model.ArticleStatusPublished, userID); err != nil {
		return err
	}

	// Log action
	if s.actionLogRepo != nil {
		s.logAction(ctx, &model.ActionLog{
			ArticleID:  id,
			ActionType: model.ActionTypePublish,
			UserID:     userID,
			UserRole:   userRole,
			OldStatus:  oldStatus,
			NewStatus:  model.ArticleStatusPublished,
		})
	}

	return nil
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

// RejectArticle rejects an article with a note
func (s *ArticleService) RejectArticle(ctx context.Context, id primitive.ObjectID, userID string, userName string, userRole model.Role, note string) error {
	// Only editors and moderators can reject
	if userRole != model.RoleEditor && userRole != model.RoleModerator {
		return fmt.Errorf("insufficient permissions: only editors and moderators can reject articles")
	}

	article, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	oldStatus := article.Status
	article.Status = model.ArticleStatusDraft

	if err := s.repo.Update(ctx, article); err != nil {
		return err
	}

	// Create rejection note in separate table
	if s.rejectionNoteRepo != nil {
		rejectionNote := &model.RejectionNote{
			ArticleID: article.ID,
			UserID:    userID,
			UserName:  userName,
			UserRole:  userRole,
			Note:      note,
			IsResolved: false,
		}
		if err := s.rejectionNoteRepo.Create(ctx, rejectionNote); err != nil {
			return err
		}
	}

	// Log action
	if s.actionLogRepo != nil {
		s.logAction(ctx, &model.ActionLog{
			ArticleID:  article.ID,
			ActionType: model.ActionTypeReject,
			UserID:     userID,
			UserName:   userName,
			UserRole:   userRole,
			Note:       note,
			OldStatus:  oldStatus,
			NewStatus:  article.Status,
		})
	}

	return nil
}

// GetArticleVersions gets all versions of an article
func (s *ArticleService) GetArticleVersions(ctx context.Context, articleID primitive.ObjectID) ([]*model.ArticleVersion, error) {
	if s.versionRepo == nil {
		return nil, fmt.Errorf("version repository not initialized")
	}
	return s.versionRepo.FindByArticleID(ctx, articleID)
}

// GetArticleVersion gets a specific version of an article
func (s *ArticleService) GetArticleVersion(ctx context.Context, versionID primitive.ObjectID) (*model.ArticleVersion, error) {
	if s.versionRepo == nil {
		return nil, fmt.Errorf("version repository not initialized")
	}
	return s.versionRepo.FindByID(ctx, versionID)
}

// RestoreArticleVersion restores an article to a specific version
func (s *ArticleService) RestoreArticleVersion(ctx context.Context, articleID primitive.ObjectID, versionNum int, userID string, userRole model.Role) error {
	// Only editors and moderators can restore versions
	if userRole != model.RoleEditor && userRole != model.RoleModerator {
		return fmt.Errorf("insufficient permissions: only editors and moderators can restore versions")
	}

	if s.versionRepo == nil {
		return fmt.Errorf("version repository not initialized")
	}

	// Get the version to restore
	version, err := s.versionRepo.FindByVersionNumber(ctx, articleID, versionNum)
	if err != nil {
		return err
	}

	// Get current article
	article, err := s.repo.FindByID(ctx, articleID)
	if err != nil {
		return err
	}

	oldStatus := article.Status

	// Restore from full snapshot if available, otherwise from version fields
	if version.FullSnapshot != nil {
		// Preserve ID and creation info
		version.FullSnapshot.ID = article.ID
		version.FullSnapshot.CreatedAt = article.CreatedAt
		version.FullSnapshot.CreatedBy = article.CreatedBy
		version.FullSnapshot.CurrentVersion = article.CurrentVersion + 1
		article = version.FullSnapshot
	} else {
		// Restore basic fields
		article.Title = version.Title
		article.Subtitle = version.Subtitle
		article.Slug = version.Slug
		article.ArticleType = version.ArticleType
		article.CategoryID = version.CategoryID
		article.Summary = version.Summary
		article.Content = version.Content
		article.Author = version.Author
		article.Tags = version.Tags
		article.SEO = version.SEO
		article.CurrentVersion++
	}

	// Update article
	if err := s.repo.Update(ctx, article); err != nil {
		return err
	}

	// Log action
	if s.actionLogRepo != nil {
		s.logAction(ctx, &model.ActionLog{
			ArticleID:  article.ID,
			ActionType: model.ActionTypeRestore,
			UserID:     userID,
			UserRole:   userRole,
			Note:       fmt.Sprintf("Restored to version %d", versionNum),
			OldStatus:  oldStatus,
			NewStatus:  article.Status,
			VersionID:  &version.ID,
		})
	}

	// Create a new version entry for the restore
	s.createVersion(ctx, article, userID, fmt.Sprintf("Restored from version %d", versionNum))

	return nil
}

// GetActionLogs gets action logs for an article
func (s *ArticleService) GetActionLogs(ctx context.Context, articleID primitive.ObjectID, page, limit int) ([]*model.ActionLog, int64, error) {
	if s.actionLogRepo == nil {
		return nil, 0, fmt.Errorf("action log repository not initialized")
	}
	return s.actionLogRepo.FindByArticleID(ctx, articleID, page, limit)
}

// GetUserActionLogs gets action logs by a specific user
func (s *ArticleService) GetUserActionLogs(ctx context.Context, userID string, page, limit int) ([]*model.ActionLog, int64, error) {
	if s.actionLogRepo == nil {
		return nil, 0, fmt.Errorf("action log repository not initialized")
	}
	return s.actionLogRepo.FindByUserID(ctx, userID, page, limit)
}

// logAction logs an action to the action log
func (s *ArticleService) logAction(ctx context.Context, log *model.ActionLog) error {
	if s.actionLogRepo == nil {
		return nil
	}
	return s.actionLogRepo.Create(ctx, log)
}

// createVersion creates a version snapshot of an article
func (s *ArticleService) createVersion(ctx context.Context, article *model.Article, userID string, note string) error {
	if s.versionRepo == nil {
		return nil
	}

	// Get latest version number
	latestVersion, err := s.versionRepo.GetLatestVersionNumber(ctx, article.ID)
	if err != nil {
		return err
	}

	version := &model.ArticleVersion{
		ArticleID:   article.ID,
		VersionNum:  latestVersion + 1,
		Title:       article.Title,
		Subtitle:    article.Subtitle,
		Slug:        article.Slug,
		ArticleType: article.ArticleType,
		CategoryID:  article.CategoryID,
		Summary:     article.Summary,
		Content:     article.Content,
		Author:      article.Author,
		Tags:        article.Tags,
		SEO:         article.SEO,
		Status:      article.Status,
		CreatedBy:   userID,
		ChangeNote:  note,
		FullSnapshot: article, // Store full article for complete restoration
	}

	return s.versionRepo.Create(ctx, version)
}

// AddRejectionNote adds a rejection note to an article (for conversation thread)
func (s *ArticleService) AddRejectionNote(ctx context.Context, articleID primitive.ObjectID, userID string, userName string, userRole model.Role, note string, parentID *primitive.ObjectID) error {
	if s.rejectionNoteRepo == nil {
		return fmt.Errorf("rejection note repository not initialized")
	}

	rejectionNote := &model.RejectionNote{
		ArticleID: articleID,
		UserID:    userID,
		UserName:  userName,
		UserRole:  userRole,
		Note:      note,
		ParentID:  parentID,
		IsResolved: false,
	}

	return s.rejectionNoteRepo.Create(ctx, rejectionNote)
}

// GetRejectionNotes gets all rejection notes for an article
func (s *ArticleService) GetRejectionNotes(ctx context.Context, articleID primitive.ObjectID) ([]*model.RejectionNote, error) {
	if s.rejectionNoteRepo == nil {
		return nil, fmt.Errorf("rejection note repository not initialized")
	}
	return s.rejectionNoteRepo.FindByArticleID(ctx, articleID)
}

// ResolveRejectionNotes marks all rejection notes for an article as resolved
func (s *ArticleService) ResolveRejectionNotes(ctx context.Context, articleID primitive.ObjectID, userID string, userRole model.Role) error {
	// Only editors and moderators can resolve rejection notes
	if userRole != model.RoleEditor && userRole != model.RoleModerator {
		return fmt.Errorf("insufficient permissions: only editors and moderators can resolve rejection notes")
	}

	if s.rejectionNoteRepo == nil {
		return fmt.Errorf("rejection note repository not initialized")
	}

	return s.rejectionNoteRepo.MarkAsResolved(ctx, articleID)
}

// GetUnresolvedRejectionCount gets the count of unresolved rejection notes for an article
func (s *ArticleService) GetUnresolvedRejectionCount(ctx context.Context, articleID primitive.ObjectID) (int64, error) {
	if s.rejectionNoteRepo == nil {
		return 0, nil
	}
	return s.rejectionNoteRepo.CountUnresolvedByArticleID(ctx, articleID)
}

// FindByTag finds articles with a specific tag
func (s *ArticleService) FindByTag(ctx context.Context, tag string, page, limit int) ([]*model.Article, int64, error) {
	return s.repo.FindByTag(ctx, tag, page, limit)
}

// FindByAuthor finds articles by author ID
func (s *ArticleService) FindByAuthor(ctx context.Context, authorID string, page, limit int) ([]*model.Article, int64, error) {
	return s.repo.FindByAuthor(ctx, authorID, page, limit)
}

// GetRelatedArticles gets related articles for an article
func (s *ArticleService) GetRelatedArticles(ctx context.Context, articleID primitive.ObjectID) ([]*model.Article, error) {
	article, err := s.repo.FindByID(ctx, articleID)
	if err != nil {
		return nil, err
	}
	
	// If manually assigned related articles exist, return them
	if len(article.RelatedArticles) > 0 {
		return s.repo.FindRelatedArticles(ctx, article.RelatedArticles)
	}
	
	// Otherwise, find similar articles by tags
	return s.repo.FindSimilarArticlesByTags(ctx, articleID, article.Tags, 5)
}

// UpdateRelatedArticles updates related articles for an article
func (s *ArticleService) UpdateRelatedArticles(ctx context.Context, articleID primitive.ObjectID, relatedIDs []primitive.ObjectID, userID string, userRole model.Role) error {
	// Only editors and moderators can update related articles
	if userRole != model.RoleEditor && userRole != model.RoleModerator {
		return fmt.Errorf("insufficient permissions: only editors and moderators can update related articles")
	}
	
	// Validate that related articles exist
	if len(relatedIDs) > 0 {
		_, err := s.repo.FindRelatedArticles(ctx, relatedIDs)
		if err != nil {
			return fmt.Errorf("failed to validate related articles: %w", err)
		}
	}
	
	return s.repo.UpdateRelatedArticles(ctx, articleID, relatedIDs)
}
