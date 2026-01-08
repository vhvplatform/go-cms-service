package service_test

import (
	"testing"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestArticleValidator_ValidateCreate(t *testing.T) {
	v := validator.NewArticleValidator()

	tests := []struct {
		name    string
		article *model.Article
		wantErr bool
	}{
		{
			name: "Valid article",
			article: &model.Article{
				Title:       "Test Article",
				ArticleType: model.ArticleTypeNews,
				CategoryID:  primitive.NewObjectID(),
				Content:     "Test content",
			},
			wantErr: false,
		},
		{
			name: "Missing title",
			article: &model.Article{
				ArticleType: model.ArticleTypeNews,
				CategoryID:  primitive.NewObjectID(),
				Content:     "Test content",
			},
			wantErr: true,
		},
		{
			name: "Missing article type",
			article: &model.Article{
				Title:      "Test Article",
				CategoryID: primitive.NewObjectID(),
				Content:    "Test content",
			},
			wantErr: true,
		},
		{
			name: "Invalid article type",
			article: &model.Article{
				Title:       "Test Article",
				ArticleType: model.ArticleType("InvalidType"),
				CategoryID:  primitive.NewObjectID(),
				Content:     "Test content",
			},
			wantErr: true,
		},
		{
			name: "Missing category ID",
			article: &model.Article{
				Title:       "Test Article",
				ArticleType: model.ArticleTypeNews,
				Content:     "Test content",
			},
			wantErr: true,
		},
		{
			name: "Missing content",
			article: &model.Article{
				Title:       "Test Article",
				ArticleType: model.ArticleTypeNews,
				CategoryID:  primitive.NewObjectID(),
			},
			wantErr: true,
		},
		{
			name: "Video without URL",
			article: &model.Article{
				Title:       "Test Video",
				ArticleType: model.ArticleTypeVideo,
				CategoryID:  primitive.NewObjectID(),
				Content:     "Test content",
			},
			wantErr: true,
		},
		{
			name: "Valid video article",
			article: &model.Article{
				Title:       "Test Video",
				ArticleType: model.ArticleTypeVideo,
				CategoryID:  primitive.NewObjectID(),
				Content:     "Test content",
				VideoURL:    "https://example.com/video.mp4",
			},
			wantErr: false,
		},
		{
			name: "Title too long",
			article: &model.Article{
				Title:       string(make([]byte, 501)),
				ArticleType: model.ArticleTypeNews,
				CategoryID:  primitive.NewObjectID(),
				Content:     "Test content",
			},
			wantErr: true,
		},
		{
			name: "Too many tags",
			article: &model.Article{
				Title:       "Test Article",
				ArticleType: model.ArticleTypeNews,
				CategoryID:  primitive.NewObjectID(),
				Content:     "Test content",
				Tags:        make([]string, 51),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateCreate(tt.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArticleValidator_ValidateReject(t *testing.T) {
	v := validator.NewArticleValidator()

	tests := []struct {
		name      string
		articleID primitive.ObjectID
		note      string
		wantErr   bool
	}{
		{
			name:      "Valid rejection",
			articleID: primitive.NewObjectID(),
			note:      "This article needs more details",
			wantErr:   false,
		},
		{
			name:      "Empty note",
			articleID: primitive.NewObjectID(),
			note:      "",
			wantErr:   true,
		},
		{
			name:      "Note too long",
			articleID: primitive.NewObjectID(),
			note:      string(make([]byte, 2001)),
			wantErr:   true,
		},
		{
			name:      "Zero article ID",
			articleID: primitive.NilObjectID,
			note:      "Test note",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateReject(tt.articleID, tt.note)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateReject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArticleValidator_ValidateID(t *testing.T) {
	v := validator.NewArticleValidator()

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "Valid ObjectID",
			id:      "507f1f77bcf86cd799439011",
			wantErr: false,
		},
		{
			name:    "Empty ID",
			id:      "",
			wantErr: true,
		},
		{
			name:    "Invalid ID format",
			id:      "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := v.ValidateID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
