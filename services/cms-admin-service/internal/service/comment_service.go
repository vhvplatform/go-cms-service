package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	MaxCommentLength      = 2000
	MaxCommentsPerHour    = 10
	MaxCommentNestingLevel = 2 // 0, 1, 2 = 3 levels
)

// CommentService handles comment business logic
type CommentService struct {
	repo *repository.CommentRepository
}

// NewCommentService creates a new comment service
func NewCommentService(repo *repository.CommentRepository) *CommentService {
	return &CommentService{
		repo: repo,
	}
}

// CreateComment creates a new comment
func (s *CommentService) CreateComment(ctx context.Context, comment *model.Comment, userID string, userName string) error {
	// Validate content
	comment.Content = strings.TrimSpace(comment.Content)
	if comment.Content == "" {
		return fmt.Errorf("comment content cannot be empty")
	}
	if len(comment.Content) > MaxCommentLength {
		return fmt.Errorf("comment exceeds maximum length of %d characters", MaxCommentLength)
	}
	
	// Check rate limit
	allowed, err := s.repo.CheckRateLimit(ctx, userID, MaxCommentsPerHour)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return fmt.Errorf("rate limit exceeded: maximum %d comments per hour", MaxCommentsPerHour)
	}
	
	// Set user info
	comment.UserID = userID
	comment.UserName = userName
	
	// Handle nested comments
	if comment.ParentID != nil {
		parent, err := s.repo.FindByID(ctx, *comment.ParentID)
		if err != nil {
			return fmt.Errorf("parent comment not found: %w", err)
		}
		
		// Check nesting level
		comment.Level = parent.Level + 1
		if comment.Level > MaxCommentNestingLevel {
			return fmt.Errorf("maximum comment nesting level (%d) exceeded", MaxCommentNestingLevel+1)
		}
		
		// Increment parent reply count
		if err := s.repo.IncrementReplyCount(ctx, *comment.ParentID); err != nil {
			return fmt.Errorf("failed to increment reply count: %w", err)
		}
	} else {
		comment.Level = 0
	}
	
	return s.repo.Create(ctx, comment)
}

// GetArticleComments gets all comments for an article with pagination
func (s *CommentService) GetArticleComments(ctx context.Context, articleID primitive.ObjectID, sortBy string, page, limit int) ([]*model.Comment, int64, error) {
	return s.repo.FindByArticleID(ctx, articleID, sortBy, page, limit)
}

// GetCommentReplies gets all replies to a comment
func (s *CommentService) GetCommentReplies(ctx context.Context, parentID primitive.ObjectID, sortBy string) ([]*model.Comment, error) {
	return s.repo.FindRepliesByParentID(ctx, parentID, sortBy)
}

// GetCommentThread gets a comment with all its nested replies
func (s *CommentService) GetCommentThread(ctx context.Context, commentID primitive.ObjectID, sortBy string) (*model.Comment, []*model.Comment, error) {
	comment, err := s.repo.FindByID(ctx, commentID)
	if err != nil {
		return nil, nil, err
	}
	
	replies, err := s.repo.FindRepliesByParentID(ctx, commentID, sortBy)
	if err != nil {
		return nil, nil, err
	}
	
	return comment, replies, nil
}

// ModerateComment moderates a comment (approve/reject)
func (s *CommentService) ModerateComment(ctx context.Context, commentID primitive.ObjectID, status model.CommentStatus, moderatorID string, note string, userRole model.Role) error {
	// Only editors and moderators can moderate comments
	if userRole != model.RoleEditor && userRole != model.RoleModerator {
		return fmt.Errorf("insufficient permissions: only editors and moderators can moderate comments")
	}
	
	return s.repo.UpdateStatus(ctx, commentID, status, moderatorID, note)
}

// LikeComment adds a like to a comment
func (s *CommentService) LikeComment(ctx context.Context, commentID primitive.ObjectID, userID string) error {
	return s.repo.LikeComment(ctx, commentID, userID)
}

// UnlikeComment removes a like from a comment
func (s *CommentService) UnlikeComment(ctx context.Context, commentID primitive.ObjectID, userID string) error {
	return s.repo.UnlikeComment(ctx, commentID, userID)
}

// HasUserLiked checks if a user has liked a comment
func (s *CommentService) HasUserLiked(ctx context.Context, commentID primitive.ObjectID, userID string) (bool, error) {
	return s.repo.HasUserLiked(ctx, commentID, userID)
}

// ReportComment reports a comment for violation
func (s *CommentService) ReportComment(ctx context.Context, commentID primitive.ObjectID, userID string, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("report reason cannot be empty")
	}
	
	report := &model.CommentReport{
		CommentID:  commentID,
		ReporterID: userID,
		Reason:     reason,
	}
	
	return s.repo.ReportComment(ctx, report)
}

// AddFavorite adds an article to user's favorites
func (s *CommentService) AddFavorite(ctx context.Context, userID string, articleID primitive.ObjectID) error {
	return s.repo.AddFavorite(ctx, userID, articleID)
}

// RemoveFavorite removes an article from user's favorites
func (s *CommentService) RemoveFavorite(ctx context.Context, userID string, articleID primitive.ObjectID) error {
	return s.repo.RemoveFavorite(ctx, userID, articleID)
}

// GetUserFavorites gets all favorited articles for a user
func (s *CommentService) GetUserFavorites(ctx context.Context, userID string, page, limit int) ([]primitive.ObjectID, int64, error) {
	return s.repo.GetUserFavorites(ctx, userID, page, limit)
}

// IsFavorited checks if a user has favorited an article
func (s *CommentService) IsFavorited(ctx context.Context, userID string, articleID primitive.ObjectID) (bool, error) {
	return s.repo.IsFavorited(ctx, userID, articleID)
}

// GetPendingComments gets all comments pending moderation
func (s *CommentService) GetPendingComments(ctx context.Context, page, limit int) ([]*model.Comment, int64, error) {
	return s.repo.FindPendingComments(ctx, page, limit)
}
