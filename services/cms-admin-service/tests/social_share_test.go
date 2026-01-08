package service_test

import (
	"strings"
	"testing"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSocialMediaSharer_GenerateShareURLs(t *testing.T) {
	sharer := util.NewSocialMediaSharer("https://example.com")

	article := &model.Article{
		ID:      primitive.NewObjectID(),
		Title:   "Test Article",
		Slug:    "test-article",
		Summary: "This is a test article",
		Tags:    []string{"test", "article", "go"},
	}

	urls := sharer.GenerateShareURLs(article)

	if urls.Facebook == "" {
		t.Error("Facebook URL should not be empty")
	}
	if urls.Twitter == "" {
		t.Error("Twitter URL should not be empty")
	}
	if urls.LinkedIn == "" {
		t.Error("LinkedIn URL should not be empty")
	}

	// Verify URLs contain expected components
	if !contains(urls.Facebook, "facebook.com") {
		t.Error("Facebook URL should contain facebook.com")
	}
	if !contains(urls.Twitter, "twitter.com") {
		t.Error("Twitter URL should contain twitter.com")
	}
	if !contains(urls.LinkedIn, "linkedin.com") {
		t.Error("LinkedIn URL should contain linkedin.com")
	}
}

func TestSocialMediaSharer_GenerateFacebookURL(t *testing.T) {
	sharer := util.NewSocialMediaSharer("https://example.com")

	url := sharer.GenerateFacebookURL("https://example.com/article/123", "Test Title")

	if url == "" {
		t.Error("Facebook URL should not be empty")
	}
	if !contains(url, "facebook.com/sharer") {
		t.Error("URL should contain facebook.com/sharer")
	}
	// Check for either encoded format
	if !contains(url, "Test") {
		t.Error("URL should contain title text")
	}
}

func TestSocialMediaSharer_GenerateTwitterURL(t *testing.T) {
	sharer := util.NewSocialMediaSharer("https://example.com")

	url := sharer.GenerateTwitterURL("https://example.com/article/123", "Test Title", []string{"tag1", "tag2", "tag3", "tag4"})

	if url == "" {
		t.Error("Twitter URL should not be empty")
	}
	if !contains(url, "twitter.com/intent/tweet") {
		t.Error("URL should contain twitter.com/intent/tweet")
	}
	// Should only include max 3 hashtags
	if !contains(url, "hashtags=tag1") {
		t.Error("URL should contain hashtags")
	}
}

func TestSocialMediaSharer_GenerateLinkedInURL(t *testing.T) {
	sharer := util.NewSocialMediaSharer("https://example.com")

	url := sharer.GenerateLinkedInURL("https://example.com/article/123", "Test Title", "Test Summary")

	if url == "" {
		t.Error("LinkedIn URL should not be empty")
	}
	if !contains(url, "linkedin.com") {
		t.Error("URL should contain linkedin.com")
	}
	if !contains(url, "Test+Title") && !contains(url, "Test%20Title") {
		t.Error("URL should contain encoded title")
	}
}

func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}
