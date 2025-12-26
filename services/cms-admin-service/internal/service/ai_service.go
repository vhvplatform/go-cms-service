package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
)

type AIService struct {
	configRepo *repository.AIConfigRepository
	logRepo    *repository.AIOperationLogRepository
	httpClient *http.Client
}

func NewAIService(configRepo *repository.AIConfigRepository, logRepo *repository.AIOperationLogRepository) *AIService {
	return &AIService{
		configRepo: configRepo,
		logRepo:    logRepo,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CheckSpelling checks spelling and grammar
func (s *AIService) CheckSpelling(ctx context.Context, tenantID, userID, text string) (*model.SpellCheckResult, error) {
	config, err := s.configRepo.GetByTenant(ctx, tenantID)
	if err != nil || config == nil || !config.SpellCheckEnabled {
		return nil, fmt.Errorf("spell check not enabled for tenant")
	}

	startTime := time.Now()

	var result *model.SpellCheckResult
	var opErr error

	switch config.SpellCheckProvider {
	case "openai":
		result, opErr = s.spellCheckOpenAI(ctx, config, text)
	case "languagetool":
		result, opErr = s.spellCheckLanguageTool(ctx, config, text)
	default:
		opErr = fmt.Errorf("unsupported spell check provider: %s", config.SpellCheckProvider)
	}

	duration := time.Since(startTime).Milliseconds()

	// Log operation
	log := &model.AIOperationLog{
		TenantID:  tenantID,
		UserID:    userID,
		Operation: "spell_check",
		Provider:  config.SpellCheckProvider,
		InputText: text,
		Success:   opErr == nil,
		Duration:  duration,
		CreatedAt: time.Now(),
	}
	if opErr != nil {
		log.Error = opErr.Error()
	}
	s.logRepo.Create(ctx, log)

	return result, opErr
}

// Translate translates text to target language
func (s *AIService) Translate(ctx context.Context, tenantID, userID, text, targetLang string) (*model.TranslationResult, error) {
	config, err := s.configRepo.GetByTenant(ctx, tenantID)
	if err != nil || config == nil || !config.TranslationEnabled {
		return nil, fmt.Errorf("translation not enabled for tenant")
	}

	startTime := time.Now()

	var result *model.TranslationResult
	var opErr error

	switch config.TranslationProvider {
	case "openai":
		result, opErr = s.translateOpenAI(ctx, config, text, targetLang)
	case "google":
		result, opErr = s.translateGoogle(ctx, config, text, targetLang)
	case "deepl":
		result, opErr = s.translateDeepL(ctx, config, text, targetLang)
	default:
		opErr = fmt.Errorf("unsupported translation provider: %s", config.TranslationProvider)
	}

	duration := time.Since(startTime).Milliseconds()

	// Log operation
	log := &model.AIOperationLog{
		TenantID:   tenantID,
		UserID:     userID,
		Operation:  "translate",
		Provider:   config.TranslationProvider,
		InputText:  text,
		TargetLang: targetLang,
		Success:    opErr == nil,
		Duration:   duration,
		CreatedAt:  time.Now(),
	}
	if result != nil {
		log.OutputText = result.TranslatedText
		log.SourceLang = result.SourceLang
	}
	if opErr != nil {
		log.Error = opErr.Error()
	}
	s.logRepo.Create(ctx, log)

	return result, opErr
}

// ImproveContent suggests content improvements
func (s *AIService) ImproveContent(ctx context.Context, tenantID, userID, articleID, content string) (*model.ContentEditSuggestion, error) {
	config, err := s.configRepo.GetByTenant(ctx, tenantID)
	if err != nil || config == nil || !config.ContentEditEnabled {
		return nil, fmt.Errorf("content editing not enabled for tenant")
	}

	startTime := time.Now()

	var result *model.ContentEditSuggestion
	var opErr error

	switch config.AIProvider {
	case "openai":
		result, opErr = s.improveContentOpenAI(ctx, config, content)
	case "anthropic":
		result, opErr = s.improveContentAnthropic(ctx, config, content)
	default:
		opErr = fmt.Errorf("unsupported AI provider: %s", config.AIProvider)
	}

	duration := time.Since(startTime).Milliseconds()

	// Log operation
	log := &model.AIOperationLog{
		TenantID:  tenantID,
		UserID:    userID,
		ArticleID: articleID,
		Operation: "improve_content",
		Provider:  config.AIProvider,
		InputText: content,
		Success:   opErr == nil,
		Duration:  duration,
		CreatedAt: time.Now(),
	}
	if result != nil {
		log.OutputText = result.Improved
	}
	if opErr != nil {
		log.Error = opErr.Error()
	}
	s.logRepo.Create(ctx, log)

	return result, opErr
}

// DetectViolation detects content violations using AI
func (s *AIService) DetectViolation(ctx context.Context, tenantID, userID, content string) (*model.ViolationDetectionResult, error) {
	config, err := s.configRepo.GetByTenant(ctx, tenantID)
	if err != nil || config == nil || !config.ViolationDetection {
		return nil, fmt.Errorf("violation detection not enabled for tenant")
	}

	startTime := time.Now()

	var result *model.ViolationDetectionResult
	var opErr error

	switch config.AIProvider {
	case "openai":
		result, opErr = s.detectViolationOpenAI(ctx, config, content)
	case "anthropic":
		result, opErr = s.detectViolationAnthropic(ctx, config, content)
	default:
		opErr = fmt.Errorf("unsupported AI provider: %s", config.AIProvider)
	}

	duration := time.Since(startTime).Milliseconds()

	// Log operation
	log := &model.AIOperationLog{
		TenantID:  tenantID,
		UserID:    userID,
		Operation: "detect_violation",
		Provider:  config.AIProvider,
		InputText: content,
		Success:   opErr == nil,
		Duration:  duration,
		CreatedAt: time.Now(),
	}
	if opErr != nil {
		log.Error = opErr.Error()
	}
	s.logRepo.Create(ctx, log)

	return result, opErr
}

// Provider-specific implementations (simplified - would need actual API integration)

func (s *AIService) spellCheckOpenAI(ctx context.Context, config *model.AIConfiguration, text string) (*model.SpellCheckResult, error) {
	// Sanitize input to prevent prompt injection
	sanitizedText := strings.ReplaceAll(text, "\n", " ")
	sanitizedText = strings.ReplaceAll(sanitizedText, "\"", "'")
	if len(sanitizedText) > 5000 {
		sanitizedText = sanitizedText[:5000] // Limit length
	}

	_ = "Check the following text for spelling and grammar errors. Return a JSON array of corrections."

	// Mock response for now - in production, integrate with actual OpenAI API
	return &model.SpellCheckResult{
		Original:    text,
		Corrections: []model.SpellCorrection{},
		HasErrors:   false,
	}, nil
}

func (s *AIService) spellCheckLanguageTool(ctx context.Context, config *model.AIConfiguration, text string) (*model.SpellCheckResult, error) {
	// Integration with LanguageTool API
	return &model.SpellCheckResult{
		Original:    text,
		Corrections: []model.SpellCorrection{},
		HasErrors:   false,
	}, nil
}

func (s *AIService) translateOpenAI(ctx context.Context, config *model.AIConfiguration, text, targetLang string) (*model.TranslationResult, error) {
	type openAIRequest struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		MaxTokens   int     `json:"max_tokens,omitempty"`
		Temperature float64 `json:"temperature,omitempty"`
	}

	req := openAIRequest{
		Model: config.Model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "system", Content: fmt.Sprintf("You are a translator. Translate the following text to %s. Return only the translation without any explanation.", targetLang)},
			{Role: "user", Content: text},
		},
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
	}

	jsonData, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", config.APIEndpoint, bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+config.APIKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Parse OpenAI response
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no translation returned")
	}

	return &model.TranslationResult{
		SourceText:     text,
		TranslatedText: openAIResp.Choices[0].Message.Content,
		TargetLang:     targetLang,
		Provider:       "openai",
	}, nil
}

func (s *AIService) translateGoogle(ctx context.Context, config *model.AIConfiguration, text, targetLang string) (*model.TranslationResult, error) {
	// Google Translate API integration
	return &model.TranslationResult{
		SourceText:     text,
		TranslatedText: text, // Placeholder
		TargetLang:     targetLang,
		Provider:       "google",
	}, nil
}

func (s *AIService) translateDeepL(ctx context.Context, config *model.AIConfiguration, text, targetLang string) (*model.TranslationResult, error) {
	// DeepL API integration
	return &model.TranslationResult{
		SourceText:     text,
		TranslatedText: text, // Placeholder
		TargetLang:     targetLang,
		Provider:       "deepl",
	}, nil
}

func (s *AIService) improveContentOpenAI(ctx context.Context, config *model.AIConfiguration, content string) (*model.ContentEditSuggestion, error) {
	// OpenAI content improvement
	return &model.ContentEditSuggestion{
		Original:   content,
		Improved:   content,
		Changes:    []string{},
		Confidence: 0.8,
	}, nil
}

func (s *AIService) improveContentAnthropic(ctx context.Context, config *model.AIConfiguration, content string) (*model.ContentEditSuggestion, error) {
	// Anthropic content improvement
	return &model.ContentEditSuggestion{
		Original:   content,
		Improved:   content,
		Changes:    []string{},
		Confidence: 0.8,
	}, nil
}

func (s *AIService) detectViolationOpenAI(ctx context.Context, config *model.AIConfiguration, content string) (*model.ViolationDetectionResult, error) {
	// OpenAI violation detection
	return &model.ViolationDetectionResult{
		HasViolation:  false,
		ViolationType: "",
		Confidence:    0.0,
		Explanation:   "No violations detected",
		Suggestions:   []string{},
	}, nil
}

func (s *AIService) detectViolationAnthropic(ctx context.Context, config *model.AIConfiguration, content string) (*model.ViolationDetectionResult, error) {
	// Anthropic violation detection
	return &model.ViolationDetectionResult{
		HasViolation:  false,
		ViolationType: "",
		Confidence:    0.0,
		Explanation:   "No violations detected",
		Suggestions:   []string{},
	}, nil
}
