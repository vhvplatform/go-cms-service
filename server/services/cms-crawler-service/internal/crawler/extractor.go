package crawler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/vhvplatform/go-cms-service/services/cms-crawler-service/internal/model"
)

type ContentExtractor struct {
	httpClient *http.Client
}

func NewContentExtractor() *ContentExtractor {
	return &ContentExtractor{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Extract extracts content from a URL using the provided configuration
func (e *ContentExtractor) Extract(ctx context.Context, sourceURL string, config model.ExtractionConfig, source *model.CrawlerSource) (*model.CrawlerArticle, error) {
	// Create request with custom headers and user agent
	req, err := http.NewRequestWithContext(ctx, "GET", sourceURL, nil)
	if err != nil {
		return nil, err
	}

	// Set user agent
	if len(source.UserAgents) > 0 {
		req.Header.Set("User-Agent", source.UserAgents[0])
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	}

	// Set custom headers
	for key, value := range source.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch URL: status %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	article := &model.CrawlerArticle{
		SourceURL: sourceURL,
		SourceID:  source.ID,
		TenantID:  source.TenantID,
		Status:    "pending",
	}

	// Extract title
	if config.TitleSelector != "" {
		article.Title = strings.TrimSpace(doc.Find(config.TitleSelector).First().Text())
	} else if config.TitleXPath != "" {
		// XPath extraction requires additional library like htmlquery or xmlpath
		// For now, fallback to basic extraction
		article.Title = strings.TrimSpace(doc.Find("title").First().Text())
	}

	// Extract content
	if config.ContentSelector != "" {
		// Remove unwanted elements first
		for _, selector := range config.RemoveSelectors {
			doc.Find(selector).Remove()
		}

		contentNode := doc.Find(config.ContentSelector).First()
		if config.UseReadability {
			// Use readability algorithm
			article.Content = e.extractReadableContent(contentNode)
		} else {
			html, _ := contentNode.Html()
			article.Content = strings.TrimSpace(html)
		}
	}

	// Extract image
	if config.ImageSelector != "" {
		imgNode := doc.Find(config.ImageSelector).First()
		if imgSrc, exists := imgNode.Attr("src"); exists {
			article.ImageURL = e.resolveURL(sourceURL, imgSrc)
		}
	}

	// Extract author
	if config.AuthorSelector != "" {
		article.Author = strings.TrimSpace(doc.Find(config.AuthorSelector).First().Text())
	}

	// Extract tags
	if config.TagsSelector != "" {
		doc.Find(config.TagsSelector).Each(func(i int, s *goquery.Selection) {
			tag := strings.TrimSpace(s.Text())
			if tag != "" {
				article.Tags = append(article.Tags, tag)
			}
		})
	}

	// Generate content hash for duplicate detection
	article.ContentHash = e.generateContentHash(article.Title + article.Content)

	// Store raw HTML if needed
	rawHTML, _ := doc.Html()
	article.RawHTML = rawHTML

	return article, nil
}

// extractReadableContent uses a simplified readability algorithm
func (e *ContentExtractor) extractReadableContent(node *goquery.Selection) string {
	// Remove script and style tags
	node.Find("script, style").Remove()

	// Get text content
	text := node.Text()

	// Clean up whitespace
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// resolveURL resolves relative URLs to absolute
func (e *ContentExtractor) resolveURL(baseURL, relativeURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return relativeURL
	}

	rel, err := url.Parse(relativeURL)
	if err != nil {
		return relativeURL
	}

	return base.ResolveReference(rel).String()
}

// generateContentHash generates SHA-256 hash of content for duplicate detection
func (e *ContentExtractor) generateContentHash(content string) string {
	hash := sha256.New()
	io.WriteString(hash, content)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// ExtractFromRSS extracts articles from RSS feed
func (e *ContentExtractor) ExtractFromRSS(ctx context.Context, feedURL string, source *model.CrawlerSource) ([]*model.CrawlerArticle, error) {
	// Create request with custom headers
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	// Set user agent
	if len(source.UserAgents) > 0 {
		req.Header.Set("User-Agent", source.UserAgents[0])
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch RSS: status %d", resp.StatusCode)
	}

	// Parse XML/RSS feed using goquery for basic extraction
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var articles []*model.CrawlerArticle

	// Extract items from RSS feed
	doc.Find("item").Each(func(i int, s *goquery.Selection) {
		article := &model.CrawlerArticle{
			SourceID: source.ID,
			TenantID: source.TenantID,
			Status:   "pending",
		}

		// Extract title
		article.Title = strings.TrimSpace(s.Find("title").Text())

		// Extract description/content
		content := s.Find("description").Text()
		if content == "" {
			// Try content:encoded without namespace (common in RSS)
			content = s.Find("encoded").Text()
		}
		article.Content = strings.TrimSpace(content)

		// Extract link as source URL
		article.SourceURL = strings.TrimSpace(s.Find("link").Text())

		// Extract author
		article.Author = strings.TrimSpace(s.Find("author").Text())
		if article.Author == "" {
			article.Author = strings.TrimSpace(s.Find("dc\\:creator").Text())
		}

		// Extract categories as tags
		s.Find("category").Each(func(j int, cat *goquery.Selection) {
			tag := strings.TrimSpace(cat.Text())
			if tag != "" {
				article.Tags = append(article.Tags, tag)
			}
		})

		// Generate content hash
		article.ContentHash = e.generateContentHash(article.Title + article.Content)

		if article.Title != "" && article.Content != "" {
			articles = append(articles, article)
		}
	})

	return articles, nil
}

// SimilarityCalculator calculates content similarity
type SimilarityCalculator struct{}

func NewSimilarityCalculator() *SimilarityCalculator {
	return &SimilarityCalculator{}
}

// CalculateSimilarity calculates similarity between two texts (0-1)
func (s *SimilarityCalculator) CalculateSimilarity(text1, text2 string) float64 {
	// Simplified Jaccard similarity
	words1 := s.tokenize(text1)
	words2 := s.tokenize(text2)

	intersection := 0
	union := make(map[string]bool)

	for word := range words1 {
		union[word] = true
		if words2[word] {
			intersection++
		}
	}

	for word := range words2 {
		union[word] = true
	}

	if len(union) == 0 {
		return 0
	}

	return float64(intersection) / float64(len(union))
}

func (s *SimilarityCalculator) tokenize(text string) map[string]bool {
	words := make(map[string]bool)
	text = strings.ToLower(text)
	tokens := strings.Fields(text)

	for _, token := range tokens {
		// Remove punctuation
		token = strings.Trim(token, ".,!?;:()[]{}\"'")
		if len(token) > 2 { // Ignore very short words
			words[token] = true
		}
	}

	return words
}
