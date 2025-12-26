package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CategoryService handles category business logic
type CategoryService struct {
	repo *repository.CategoryRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		repo: repo,
	}
}

// Create creates a new category
func (s *CategoryService) Create(ctx context.Context, category *model.Category, userID string) error {
	// Generate slug if not provided
	if category.Slug == "" {
		category.Slug = s.generateSlug(category.Name)
	}

	// Validate category type
	if category.CategoryType == model.CategoryTypeArticle && category.ArticleType == "" {
		return fmt.Errorf("articleType is required for Article category type")
	}

	if category.CategoryType == model.CategoryTypeLink && category.CategoryLink == "" {
		return fmt.Errorf("categoryLink is required for Link category type")
	}

	category.CreatedBy = userID
	return s.repo.Create(ctx, category)
}

// FindByID finds a category by ID
func (s *CategoryService) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Category, error) {
	return s.repo.FindByID(ctx, id)
}

// FindBySlug finds a category by slug
func (s *CategoryService) FindBySlug(ctx context.Context, slug string) (*model.Category, error) {
	return s.repo.FindBySlug(ctx, slug)
}

// Update updates a category
func (s *CategoryService) Update(ctx context.Context, category *model.Category) error {
	return s.repo.Update(ctx, category)
}

// Delete deletes a category
func (s *CategoryService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

// GetTree gets the category tree structure
func (s *CategoryService) GetTree(ctx context.Context) ([]*CategoryNode, error) {
	allCategories, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Build tree structure
	categoryMap := make(map[string]*CategoryNode)
	var roots []*CategoryNode

	// First pass: create nodes
	for _, cat := range allCategories {
		node := &CategoryNode{
			Category: cat,
			Children: []*CategoryNode{},
		}
		categoryMap[cat.ID.Hex()] = node
	}

	// Second pass: build tree
	for _, node := range categoryMap {
		if node.Category.ParentID == nil {
			roots = append(roots, node)
		} else {
			parentID := node.Category.ParentID.Hex()
			if parent, ok := categoryMap[parentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return roots, nil
}

// GetChildren gets direct children of a category
func (s *CategoryService) GetChildren(ctx context.Context, parentID *primitive.ObjectID) ([]*model.Category, error) {
	return s.repo.FindByParentID(ctx, parentID)
}

// generateSlug generates a URL-friendly slug from a name
func (s *CategoryService) generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

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

// CategoryNode represents a node in the category tree
type CategoryNode struct {
	Category *model.Category `json:"category"`
	Children []*CategoryNode `json:"children"`
}
