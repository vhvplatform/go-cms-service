package service

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
)

// RSS represents an RSS 2.0 feed
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel *Channel `xml:"channel"`
}

// Channel represents an RSS channel
type Channel struct {
	Title          string `xml:"title"`
	Link           string `xml:"link"`
	Description    string `xml:"description"`
	Language       string `xml:"language,omitempty"`
	Copyright      string `xml:"copyright,omitempty"`
	ManagingEditor string `xml:"managingEditor,omitempty"`
	WebMaster      string `xml:"webMaster,omitempty"`
	PubDate        string `xml:"pubDate,omitempty"`
	LastBuildDate  string `xml:"lastBuildDate"`
	Category       string `xml:"category,omitempty"`
	Generator      string `xml:"generator,omitempty"`
	Docs           string `xml:"docs,omitempty"`
	TTL            int    `xml:"ttl,omitempty"`
	Items          []Item `xml:"item"`
}

// Item represents an RSS item
type Item struct {
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Description string     `xml:"description"`
	Author      string     `xml:"author,omitempty"`
	Category    string     `xml:"category,omitempty"`
	Comments    string     `xml:"comments,omitempty"`
	Enclosure   *Enclosure `xml:"enclosure,omitempty"`
	GUID        string     `xml:"guid"`
	PubDate     string     `xml:"pubDate"`
	Source      string     `xml:"source,omitempty"`
}

// Enclosure represents an RSS enclosure (for media)
type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

// RSSService handles RSS feed generation
type RSSService struct {
	articleRepo *repository.ArticleRepository
	baseURL     string
}

// NewRSSService creates a new RSS service
func NewRSSService(articleRepo *repository.ArticleRepository, baseURL string) *RSSService {
	return &RSSService{
		articleRepo: articleRepo,
		baseURL:     baseURL,
	}
}

// GenerateFeed generates an RSS feed for published articles
func (s *RSSService) GenerateFeed(ctx context.Context, limit int, categoryID *string) (string, error) {
	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}

	// Build filter for published articles
	filter := map[string]interface{}{
		"status": model.ArticleStatusPublished,
	}

	if categoryID != nil && *categoryID != "" {
		filter["categoryId"] = *categoryID
	}

	// Get articles
	articles, _, err := s.articleRepo.FindAll(ctx, filter, 1, limit, map[string]int{
		"publishAt": -1, // Sort by publish date descending
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch articles: %w", err)
	}

	// Build RSS feed
	rss := &RSS{
		Version: "2.0",
		Channel: &Channel{
			Title:         "CMS Service - Latest Articles",
			Link:          s.baseURL,
			Description:   "Latest published articles from CMS Service",
			Language:      "en",
			Generator:     "CMS Service RSS Generator",
			LastBuildDate: formatRFC822(time.Now()),
			TTL:           60, // 60 minutes
			Items:         make([]Item, 0, len(articles)),
		},
	}

	// Convert articles to RSS items
	for _, article := range articles {
		item := Item{
			Title:       article.Title,
			Link:        s.getArticleURL(article),
			Description: s.formatDescription(article),
			Author:      article.Author.Name,
			GUID:        article.ID.Hex(),
			PubDate:     formatRFC822(article.PublishAt),
		}

		// Add enclosure for video/podcast articles
		if article.ArticleType == model.ArticleTypeVideo && article.VideoURL != "" {
			item.Enclosure = &Enclosure{
				URL:    article.VideoURL,
				Type:   "video/mp4",
				Length: 0, // Would need to be set if known
			}
		} else if article.ArticleType == model.ArticleTypePodcast && article.AudioURL != "" {
			item.Enclosure = &Enclosure{
				URL:    article.AudioURL,
				Type:   "audio/mpeg",
				Length: 0,
			}
		}

		rss.Channel.Items = append(rss.Channel.Items, item)
	}

	// Marshal to XML
	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(rss); err != nil {
		return "", fmt.Errorf("failed to encode RSS: %w", err)
	}

	return buf.String(), nil
}

// getArticleURL constructs the full URL for an article
func (s *RSSService) getArticleURL(article *model.Article) string {
	if article.Slug != "" {
		return fmt.Sprintf("%s/articles/%s", s.baseURL, article.Slug)
	}
	return fmt.Sprintf("%s/articles/%s", s.baseURL, article.ID.Hex())
}

// formatDescription formats the article description for RSS
func (s *RSSService) formatDescription(article *model.Article) string {
	description := article.Summary
	if description == "" {
		// Use first 200 characters of content
		description = article.Content
		if len(description) > 200 {
			description = description[:200] + "..."
		}
	}

	// Escape HTML
	return html.EscapeString(description)
}

// formatRFC822 formats time to RFC822 format for RSS
func formatRFC822(t time.Time) string {
	return t.Format(time.RFC1123Z)
}
