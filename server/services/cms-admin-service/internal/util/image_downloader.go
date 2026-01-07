package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ImageDownloader handles downloading external images
type ImageDownloader struct {
	uploadDir string
	baseURL   string
	client    *http.Client
}

// NewImageDownloader creates a new image downloader
func NewImageDownloader(uploadDir, baseURL string) *ImageDownloader {
	return &ImageDownloader{
		uploadDir: uploadDir,
		baseURL:   baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessArticleContent processes article content and downloads external images
func (d *ImageDownloader) ProcessArticleContent(content string) (string, error) {
	// Find all image URLs in content
	imageRegex := regexp.MustCompile(`<img[^>]+src="([^"]+)"`)
	matches := imageRegex.FindAllStringSubmatch(content, -1)

	processedContent := content

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		originalURL := match[1]

		// Skip if already a local image
		if strings.HasPrefix(originalURL, d.baseURL) || strings.HasPrefix(originalURL, "/") {
			continue
		}

		// Skip data URLs
		if strings.HasPrefix(originalURL, "data:") {
			continue
		}

		// Download and save the image
		localPath, err := d.DownloadImage(originalURL)
		if err != nil {
			// Log error but continue processing
			fmt.Printf("Failed to download image %s: %v\n", originalURL, err)
			continue
		}

		// Replace URL in content
		localURL := d.baseURL + "/uploads/" + localPath
		processedContent = strings.ReplaceAll(processedContent, originalURL, localURL)
	}

	return processedContent, nil
}

// DownloadImage downloads an image from URL and saves it locally
func (d *ImageDownloader) DownloadImage(url string) (string, error) {
	// Download image
	resp, err := d.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	// Validate content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("not an image: %s", contentType)
	}

	// Generate filename
	filename := d.generateFilename(url, contentType)

	// Create directory if not exists
	yearMonth := time.Now().Format("2006/01")
	targetDir := filepath.Join(d.uploadDir, yearMonth)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Full path
	fullPath := filepath.Join(targetDir, filename)

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		// File exists, return existing path
		return filepath.Join(yearMonth, filename), nil
	}

	// Save file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		os.Remove(fullPath) // Clean up on error
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filepath.Join(yearMonth, filename), nil
}

// generateFilename generates a unique filename for an image
func (d *ImageDownloader) generateFilename(url, contentType string) string {
	// Generate hash from URL
	hash := md5.Sum([]byte(url))
	hashStr := fmt.Sprintf("%x", hash)[:12]

	// Get extension from content type
	ext := d.getExtensionFromContentType(contentType)

	// Generate filename: timestamp-hash.ext
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d-%s%s", timestamp, hashStr, ext)
}

// getExtensionFromContentType returns file extension from content type
func (d *ImageDownloader) getExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	default:
		return ".jpg" // Default to jpg
	}
}

// ProcessHTMLImages finds and processes images in HTML content
func (d *ImageDownloader) ProcessHTMLImages(html string) (string, []string, error) {
	downloadedImages := []string{}
	processedHTML := html

	// Find all image tags
	imgRegex := regexp.MustCompile(`<img[^>]*src="([^"]+)"[^>]*>`)
	matches := imgRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		originalURL := match[1]

		// Skip local and data URLs
		if strings.HasPrefix(originalURL, d.baseURL) ||
			strings.HasPrefix(originalURL, "/") ||
			strings.HasPrefix(originalURL, "data:") {
			continue
		}

		// Download image
		localPath, err := d.DownloadImage(originalURL)
		if err != nil {
			fmt.Printf("Warning: Failed to download %s: %v\n", originalURL, err)
			continue
		}

		// Build local URL
		localURL := d.baseURL + "/uploads/" + localPath

		// Replace in HTML
		processedHTML = strings.ReplaceAll(processedHTML, originalURL, localURL)
		downloadedImages = append(downloadedImages, localPath)
	}

	return processedHTML, downloadedImages, nil
}
