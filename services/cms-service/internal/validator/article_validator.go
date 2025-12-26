package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/vhvplatform/go-cms-service/services/cms-service/internal/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ValidationError represents a validation error with field information
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ArticleValidator validates article input
type ArticleValidator struct{}

// NewArticleValidator creates a new article validator
func NewArticleValidator() *ArticleValidator {
	return &ArticleValidator{}
}

// ValidateCreate validates article creation input
func (v *ArticleValidator) ValidateCreate(article *model.Article) error {
	var errs ValidationErrors

	// Required fields
	if strings.TrimSpace(article.Title) == "" {
		errs = append(errs, ValidationError{Field: "title", Message: "title is required"})
	} else if utf8.RuneCountInString(article.Title) > 500 {
		errs = append(errs, ValidationError{Field: "title", Message: "title must not exceed 500 characters"})
	}

	if article.ArticleType == "" {
		errs = append(errs, ValidationError{Field: "articleType", Message: "article type is required"})
	} else if !v.isValidArticleType(article.ArticleType) {
		errs = append(errs, ValidationError{Field: "articleType", Message: "invalid article type"})
	}

	if article.CategoryID.IsZero() {
		errs = append(errs, ValidationError{Field: "categoryId", Message: "category ID is required"})
	}

	// Content validation
	if strings.TrimSpace(article.Content) == "" && len(article.ContentBlocks) == 0 {
		errs = append(errs, ValidationError{Field: "content", Message: "content or content blocks are required"})
	}

	// Summary validation
	if strings.TrimSpace(article.Summary) != "" && utf8.RuneCountInString(article.Summary) > 1000 {
		errs = append(errs, ValidationError{Field: "summary", Message: "summary must not exceed 1000 characters"})
	}

	// Slug validation (if provided)
	if article.Slug != "" && !v.isValidSlug(article.Slug) {
		errs = append(errs, ValidationError{Field: "slug", Message: "slug must contain only lowercase letters, numbers, and hyphens"})
	}

	// Tags validation
	if len(article.Tags) > 50 {
		errs = append(errs, ValidationError{Field: "tags", Message: "cannot have more than 50 tags"})
	}
	for _, tag := range article.Tags {
		if utf8.RuneCountInString(tag) > 100 {
			errs = append(errs, ValidationError{Field: "tags", Message: "each tag must not exceed 100 characters"})
			break
		}
	}

	// SEO validation
	if article.SEO.Title != "" && utf8.RuneCountInString(article.SEO.Title) > 200 {
		errs = append(errs, ValidationError{Field: "seo.title", Message: "SEO title must not exceed 200 characters"})
	}
	if article.SEO.Description != "" && utf8.RuneCountInString(article.SEO.Description) > 500 {
		errs = append(errs, ValidationError{Field: "seo.description", Message: "SEO description must not exceed 500 characters"})
	}

	// Type-specific validation
	switch article.ArticleType {
	case model.ArticleTypeVideo:
		if article.VideoURL == "" {
			errs = append(errs, ValidationError{Field: "videoUrl", Message: "video URL is required for video articles"})
		} else if !v.isValidURL(article.VideoURL) {
			errs = append(errs, ValidationError{Field: "videoUrl", Message: "invalid video URL format"})
		}
	case model.ArticleTypePodcast:
		if article.AudioURL == "" {
			errs = append(errs, ValidationError{Field: "audioUrl", Message: "audio URL is required for podcast articles"})
		} else if !v.isValidURL(article.AudioURL) {
			errs = append(errs, ValidationError{Field: "audioUrl", Message: "invalid audio URL format"})
		}
	case model.ArticleTypePhotoGallery:
		if len(article.Images) == 0 {
			errs = append(errs, ValidationError{Field: "images", Message: "at least one image is required for photo gallery"})
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ValidateUpdate validates article update input
func (v *ArticleValidator) ValidateUpdate(article *model.Article) error {
	var errs ValidationErrors

	if article.ID.IsZero() {
		return errors.New("article ID is required for update")
	}

	// Same validation as create for fields that can be updated
	if strings.TrimSpace(article.Title) == "" {
		errs = append(errs, ValidationError{Field: "title", Message: "title is required"})
	} else if utf8.RuneCountInString(article.Title) > 500 {
		errs = append(errs, ValidationError{Field: "title", Message: "title must not exceed 500 characters"})
	}

	if article.ArticleType == "" {
		errs = append(errs, ValidationError{Field: "articleType", Message: "article type is required"})
	} else if !v.isValidArticleType(article.ArticleType) {
		errs = append(errs, ValidationError{Field: "articleType", Message: "invalid article type"})
	}

	if strings.TrimSpace(article.Summary) != "" && utf8.RuneCountInString(article.Summary) > 1000 {
		errs = append(errs, ValidationError{Field: "summary", Message: "summary must not exceed 1000 characters"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ValidateReject validates rejection input
func (v *ArticleValidator) ValidateReject(articleID primitive.ObjectID, note string) error {
	if articleID.IsZero() {
		return errors.New("article ID is required")
	}

	if strings.TrimSpace(note) == "" {
		return ValidationError{Field: "note", Message: "rejection note is required"}
	}

	if utf8.RuneCountInString(note) > 2000 {
		return ValidationError{Field: "note", Message: "rejection note must not exceed 2000 characters"}
	}

	return nil
}

// ValidateID validates that an ID is valid
func (v *ArticleValidator) ValidateID(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.NilObjectID, errors.New("ID is required")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid ID format")
	}

	return objectID, nil
}

// isValidArticleType checks if article type is valid
func (v *ArticleValidator) isValidArticleType(articleType model.ArticleType) bool {
	validTypes := []model.ArticleType{
		model.ArticleTypeNews,
		model.ArticleTypeVideo,
		model.ArticleTypePhotoGallery,
		model.ArticleTypeLegalDocument,
		model.ArticleTypeStaffProfile,
		model.ArticleTypeJob,
		model.ArticleTypeProcedure,
		model.ArticleTypeDownload,
		model.ArticleTypePodcast,
		model.ArticleTypeEventInfo,
		model.ArticleTypeInfographic,
		model.ArticleTypeDestination,
		model.ArticleTypePartner,
		model.ArticleTypePDF,
	}

	for _, vt := range validTypes {
		if articleType == vt {
			return true
		}
	}
	return false
}

// isValidSlug checks if slug format is valid
func (v *ArticleValidator) isValidSlug(slug string) bool {
	// Slug should contain only lowercase letters, numbers, and hyphens
	matched, _ := regexp.MatchString("^[a-z0-9-]+$", slug)
	return matched
}

// isValidURL checks if URL format is valid
func (v *ArticleValidator) isValidURL(url string) bool {
	// Basic URL validation
	matched, _ := regexp.MatchString(`^https?://[^\s/$.?#].[^\s]*$`, url)
	return matched
}
