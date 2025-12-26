package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/vhvplatform/go-cms-service/services/article-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/article-service/internal/repository"
	"github.com/vhvplatform/go-cms-service/services/article-service/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockArticleRepository is a mock implementation of ArticleRepository
type MockArticleRepository struct {
	articles map[string]*model.Article
}

func NewMockArticleRepository() *MockArticleRepository {
	return &MockArticleRepository{
		articles: make(map[string]*model.Article),
	}
}

func (m *MockArticleRepository) Create(ctx context.Context, article *model.Article) error {
	if article.ID.IsZero() {
		article.ID = primitive.NewObjectID()
	}
	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()
	m.articles[article.ID.Hex()] = article
	return nil
}

func (m *MockArticleRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Article, error) {
	article, exists := m.articles[id.Hex()]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return article, nil
}

func (m *MockArticleRepository) Update(ctx context.Context, article *model.Article) error {
	article.UpdatedAt = time.Now()
	m.articles[article.ID.Hex()] = article
	return nil
}

// MockViewQueue is a mock implementation of ViewQueue
type MockViewQueue struct {
	views []primitive.ObjectID
}

func (m *MockViewQueue) Enqueue(articleID primitive.ObjectID) error {
	m.views = append(m.views, articleID)
	return nil
}

// TestArticleService_Create tests article creation
func TestArticleService_Create(t *testing.T) {
	// Arrange
	repo := NewMockArticleRepository()
	service := service.NewArticleService(repo, nil, nil, nil)
	
	article := &model.Article{
		Title:       "Test Article",
		ArticleType: model.ArticleTypeNews,
		Content:     "This is test content",
		Status:      model.ArticleStatusDraft,
	}
	
	// Act
	err := service.Create(context.Background(), article, "user123")
	
	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if article.ID.IsZero() {
		t.Error("Expected article ID to be set")
	}
	
	if article.Slug == "" {
		t.Error("Expected slug to be generated")
	}
	
	if article.CharCount == 0 {
		t.Error("Expected character count to be calculated")
	}
	
	if article.CreatedBy != "user123" {
		t.Errorf("Expected createdBy to be 'user123', got '%s'", article.CreatedBy)
	}
}

// TestArticleService_GenerateSlug tests slug generation
func TestArticleService_GenerateSlug(t *testing.T) {
	repo := NewMockArticleRepository()
	service := service.NewArticleService(repo, nil, nil, nil)
	
	testCases := []struct {
		name     string
		title    string
		wantSlug bool
	}{
		{
			name:     "Simple title",
			title:    "Test Article",
			wantSlug: true,
		},
		{
			name:     "Title with special characters",
			title:    "Test Article!!! @#$%",
			wantSlug: true,
		},
		{
			name:     "Unicode title",
			title:    "Bài viết tiếng Việt",
			wantSlug: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			article := &model.Article{
				Title:       tc.title,
				ArticleType: model.ArticleTypeNews,
				Status:      model.ArticleStatusDraft,
			}
			
			err := service.Create(context.Background(), article, "user123")
			
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			
			if tc.wantSlug && article.Slug == "" {
				t.Error("Expected slug to be generated")
			}
		})
	}
}

// TestArticleService_Update_PermissionCheck tests permission checking during update
func TestArticleService_Update_PermissionCheck(t *testing.T) {
	repo := NewMockArticleRepository()
	service := service.NewArticleService(repo, nil, nil, nil)
	
	// Create a published article
	article := &model.Article{
		ID:          primitive.NewObjectID(),
		Title:       "Published Article",
		ArticleType: model.ArticleTypeNews,
		Status:      model.ArticleStatusPublished,
		CreatedBy:   "user123",
	}
	repo.Create(context.Background(), article)
	
	// Try to update as a writer
	article.Title = "Updated Title"
	err := service.Update(context.Background(), article, "user123", model.RoleWriter)
	
	if err == nil {
		t.Error("Expected error when writer tries to edit published article, got nil")
	}
	
	// Try to update as an editor (should succeed)
	err = service.Update(context.Background(), article, "editor456", model.RoleEditor)
	
	if err != nil {
		t.Errorf("Expected no error for editor, got %v", err)
	}
}

// TestArticleService_CharCount tests character counting
func TestArticleService_CharCount(t *testing.T) {
	repo := NewMockArticleRepository()
	service := service.NewArticleService(repo, nil, nil, nil)
	
	article := &model.Article{
		Title:       "Test",
		ArticleType: model.ArticleTypeNews,
		Content:     "This is a test article with some content.",
		Status:      model.ArticleStatusDraft,
	}
	
	err := service.Create(context.Background(), article, "user123")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if article.CharCount == 0 {
		t.Error("Expected character count to be greater than 0")
	}
}

// BenchmarkArticleService_Create benchmarks article creation
func BenchmarkArticleService_Create(b *testing.B) {
	repo := NewMockArticleRepository()
	service := service.NewArticleService(repo, nil, nil, nil)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		article := &model.Article{
			Title:       "Benchmark Article",
			ArticleType: model.ArticleTypeNews,
			Content:     "Benchmark content",
			Status:      model.ArticleStatusDraft,
		}
		
		service.Create(context.Background(), article, "user123")
	}
}
