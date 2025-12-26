package util

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/vhvplatform/go-cms-service/services/cms-service/internal/model"
)

// SocialShareURLs represents social media sharing URLs
type SocialShareURLs struct {
	Facebook string `json:"facebook"`
	Twitter  string `json:"twitter"`
	LinkedIn string `json:"linkedin"`
}

// SocialMediaSharer generates social media sharing URLs
type SocialMediaSharer struct {
	baseURL string
}

// NewSocialMediaSharer creates a new social media sharer
func NewSocialMediaSharer(baseURL string) *SocialMediaSharer {
	return &SocialMediaSharer{
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// GenerateShareURLs generates sharing URLs for all supported platforms
func (s *SocialMediaSharer) GenerateShareURLs(article *model.Article) *SocialShareURLs {
	articleURL := s.getArticleURL(article)
	title := article.Title
	summary := article.Summary
	
	return &SocialShareURLs{
		Facebook: s.GenerateFacebookURL(articleURL, title),
		Twitter:  s.GenerateTwitterURL(articleURL, title, article.Tags),
		LinkedIn: s.GenerateLinkedInURL(articleURL, title, summary),
	}
}

// GenerateFacebookURL generates a Facebook share URL
func (s *SocialMediaSharer) GenerateFacebookURL(articleURL, title string) string {
	params := url.Values{}
	params.Add("u", articleURL)
	params.Add("quote", title)
	
	return fmt.Sprintf("https://www.facebook.com/sharer/sharer.php?%s", params.Encode())
}

// GenerateTwitterURL generates a Twitter share URL
func (s *SocialMediaSharer) GenerateTwitterURL(articleURL, title string, tags []string) string {
	params := url.Values{}
	params.Add("url", articleURL)
	params.Add("text", title)
	
	// Add hashtags (max 3 for better UX)
	if len(tags) > 0 {
		hashtags := tags
		if len(hashtags) > 3 {
			hashtags = hashtags[:3]
		}
		params.Add("hashtags", strings.Join(hashtags, ","))
	}
	
	return fmt.Sprintf("https://twitter.com/intent/tweet?%s", params.Encode())
}

// GenerateLinkedInURL generates a LinkedIn share URL
func (s *SocialMediaSharer) GenerateLinkedInURL(articleURL, title, summary string) string {
	params := url.Values{}
	params.Add("url", articleURL)
	params.Add("title", title)
	if summary != "" {
		params.Add("summary", summary)
	}
	
	return fmt.Sprintf("https://www.linkedin.com/sharing/share-offsite/?%s", params.Encode())
}

// getArticleURL constructs the full URL for an article
func (s *SocialMediaSharer) getArticleURL(article *model.Article) string {
	if article.Slug != "" {
		return fmt.Sprintf("%s/articles/%s", s.baseURL, article.Slug)
	}
	return fmt.Sprintf("%s/articles/%s", s.baseURL, article.ID.Hex())
}
