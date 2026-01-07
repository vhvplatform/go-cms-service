package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CMSClient handles communication with CMS Service
type CMSClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewCMSClient creates a new CMS client
func NewCMSClient(baseURL string) *CMSClient {
	return &CMSClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetArticle fetches an article from CMS service
func (c *CMSClient) GetArticle(ctx context.Context, articleID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/public/articles/%s", c.baseURL, articleID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get article: status %d", resp.StatusCode)
	}

	var article map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&article); err != nil {
		return nil, err
	}

	return article, nil
}

// ListArticles fetches list of articles from CMS service
func (c *CMSClient) ListArticles(ctx context.Context, page, limit int, filters map[string]string) ([]map[string]interface{}, int64, error) {
	url := fmt.Sprintf("%s/api/v1/public/articles?page=%d&limit=%d", c.baseURL, page, limit)

	// Add filters to query
	for k, v := range filters {
		url += fmt.Sprintf("&%s=%s", k, v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("failed to list articles: status %d", resp.StatusCode)
	}

	var result struct {
		Articles []map[string]interface{} `json:"articles"`
		Total    int64                    `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	return result.Articles, result.Total, nil
}

// RecordView records a view on an article
func (c *CMSClient) RecordView(ctx context.Context, articleID string) error {
	url := fmt.Sprintf("%s/api/v1/public/articles/%s/view", c.baseURL, articleID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to record view: status %d", resp.StatusCode)
	}

	return nil
}

// StatsClient handles communication with Stats Service
type StatsClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewStatsClient creates a new stats client
func NewStatsClient(baseURL string) *StatsClient {
	return &StatsClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetComments fetches comments for an article
func (c *StatsClient) GetComments(ctx context.Context, articleID string, page, limit int, sortBy string) ([]map[string]interface{}, int64, error) {
	url := fmt.Sprintf("%s/api/v1/articles/%s/comments?page=%d&limit=%d&sortBy=%s",
		c.baseURL, articleID, page, limit, sortBy)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("failed to get comments: status %d", resp.StatusCode)
	}

	var result struct {
		Comments []map[string]interface{} `json:"comments"`
		Total    int64                    `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	return result.Comments, result.Total, nil
}

// CreateComment posts a comment
func (c *StatsClient) CreateComment(ctx context.Context, articleID, userID, userName, content string, parentID *string, authToken string) error {
	url := fmt.Sprintf("%s/api/v1/articles/%s/comments", c.baseURL, articleID)

	payload := map[string]interface{}{
		"content": content,
	}
	if parentID != nil {
		payload["parentId"] = *parentID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create comment: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// LikeComment likes a comment
func (c *StatsClient) LikeComment(ctx context.Context, commentID, authToken string) error {
	url := fmt.Sprintf("%s/api/v1/comments/%s/like", c.baseURL, commentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to like comment: status %d", resp.StatusCode)
	}

	return nil
}

// AddFavorite adds article to favorites
func (c *StatsClient) AddFavorite(ctx context.Context, articleID, authToken string) error {
	url := fmt.Sprintf("%s/api/v1/articles/%s/favorite", c.baseURL, articleID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add favorite: status %d", resp.StatusCode)
	}

	return nil
}

// VoteOnPoll votes on a poll
func (c *StatsClient) VoteOnPoll(ctx context.Context, pollID string, optionIDs []string, authToken string) error {
	url := fmt.Sprintf("%s/api/v1/polls/%s/vote", c.baseURL, pollID)

	payload := map[string]interface{}{
		"optionIds": optionIDs,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to vote: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
