package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SensitiveKeywordService struct {
	repo *repository.SensitiveKeywordRepository
}

func NewSensitiveKeywordService(repo *repository.SensitiveKeywordRepository) *SensitiveKeywordService {
	return &SensitiveKeywordService{repo: repo}
}

// Create adds a new sensitive keyword
func (s *SensitiveKeywordService) Create(ctx context.Context, keyword *model.SensitiveKeyword) error {
	// Validate regex if isRegex is true
	if keyword.IsRegex {
		if _, err := regexp.Compile(keyword.Keyword); err != nil {
			return err
		}
	}
	return s.repo.Create(ctx, keyword)
}

// GetByID retrieves a keyword by ID
func (s *SensitiveKeywordService) GetByID(ctx context.Context, id primitive.ObjectID) (*model.SensitiveKeyword, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByTenant retrieves all keywords for a tenant
func (s *SensitiveKeywordService) GetByTenant(ctx context.Context, tenantID string, activeOnly bool) ([]*model.SensitiveKeyword, error) {
	return s.repo.GetByTenant(ctx, tenantID, activeOnly)
}

// Update updates a keyword
func (s *SensitiveKeywordService) Update(ctx context.Context, keyword *model.SensitiveKeyword) error {
	if keyword.IsRegex {
		if _, err := regexp.Compile(keyword.Keyword); err != nil {
			return err
		}
	}
	return s.repo.Update(ctx, keyword)
}

// Delete deletes a keyword
func (s *SensitiveKeywordService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

// ScanContent scans content for sensitive keywords
func (s *SensitiveKeywordService) ScanContent(ctx context.Context, tenantID, content string) (*model.ContentScanResult, error) {
	keywords, err := s.repo.GetByTenant(ctx, tenantID, true)
	if err != nil {
		return nil, err
	}

	result := &model.ContentScanResult{
		HasViolations:     false,
		Results:           []model.KeywordDetectionResult{},
		HighestSeverity:   "none",
		RecommendedAction: "none",
		ScannedAt:         time.Now(),
	}

	severityLevels := map[string]int{"low": 1, "medium": 2, "high": 3, "critical": 4}
	highestLevel := 0

	contentLower := strings.ToLower(content)

	for _, kw := range keywords {
		var matches []string
		var positions []int

		if kw.IsRegex {
			// Regex matching
			re, err := regexp.Compile("(?i)" + kw.Keyword)
			if err != nil {
				continue
			}
			found := re.FindAllStringIndex(content, -1)
			if len(found) > 0 {
				for _, match := range found {
					matches = append(matches, content[match[0]:match[1]])
					positions = append(positions, match[0])
				}
			}
		} else {
			// Simple keyword matching
			keywordLower := strings.ToLower(kw.Keyword)
			keywordLen := len(keywordLower)
			index := 0
			for {
				pos := strings.Index(contentLower[index:], keywordLower)
				if pos == -1 {
					break
				}
				actualPos := index + pos
				// Ensure we don't go out of bounds
				endPos := actualPos + keywordLen
				if endPos > len(content) {
					endPos = len(content)
				}
				matches = append(matches, content[actualPos:endPos])
				positions = append(positions, actualPos)
				index = actualPos + keywordLen
			}
		}

		if len(matches) > 0 {
			result.HasViolations = true
			result.Results = append(result.Results, model.KeywordDetectionResult{
				Keyword:     kw.Keyword,
				Matches:     matches,
				Positions:   positions,
				Severity:    kw.Severity,
				Action:      kw.Action,
				Category:    kw.Category,
				Description: kw.Description,
			})

			// Track highest severity
			if level, ok := severityLevels[kw.Severity]; ok && level > highestLevel {
				highestLevel = level
				result.HighestSeverity = kw.Severity
				result.RecommendedAction = kw.Action
			}
		}
	}

	return result, nil
}

// BulkImport imports multiple keywords at once
func (s *SensitiveKeywordService) BulkImport(ctx context.Context, keywords []*model.SensitiveKeyword) error {
	// Validate all regex patterns first
	for _, kw := range keywords {
		if kw.IsRegex {
			if _, err := regexp.Compile(kw.Keyword); err != nil {
				return err
			}
		}
	}
	return s.repo.BulkCreate(ctx, keywords)
}
