package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AIConfiguration represents AI service configuration for a tenant
type AIConfiguration struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID string             `bson:"tenant_id" json:"tenantId"`

	// Feature flags
	SpellCheckEnabled  bool `bson:"spell_check_enabled" json:"spellCheckEnabled"`
	TranslationEnabled bool `bson:"translation_enabled" json:"translationEnabled"`
	ContentEditEnabled bool `bson:"content_edit_enabled" json:"contentEditEnabled"`
	ViolationDetection bool `bson:"violation_detection" json:"violationDetection"`

	// Provider configurations
	SpellCheckProvider  string `bson:"spell_check_provider,omitempty" json:"spellCheckProvider,omitempty"`  // openai, languagetool, etc
	TranslationProvider string `bson:"translation_provider,omitempty" json:"translationProvider,omitempty"` // google, deepl, openai
	AIProvider          string `bson:"ai_provider,omitempty" json:"aiProvider,omitempty"`                   // openai, anthropic, etc

	// API configurations
	APIKey      string `bson:"api_key,omitempty" json:"apiKey,omitempty"`
	APIEndpoint string `bson:"api_endpoint,omitempty" json:"apiEndpoint,omitempty"`
	Model       string `bson:"model,omitempty" json:"model,omitempty"` // e.g., gpt-4, claude-3

	// Advanced settings
	MaxTokens          int      `bson:"max_tokens,omitempty" json:"maxTokens,omitempty"`
	Temperature        float64  `bson:"temperature,omitempty" json:"temperature,omitempty"`
	SupportedLanguages []string `bson:"supported_languages,omitempty" json:"supportedLanguages,omitempty"`

	// Usage limits
	DailyLimit   int `bson:"daily_limit,omitempty" json:"dailyLimit,omitempty"`
	MonthlyLimit int `bson:"monthly_limit,omitempty" json:"monthlyLimit,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

// AIOperationLog represents a log of AI operations
type AIOperationLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID   string             `bson:"tenant_id" json:"tenantId"`
	UserID     string             `bson:"user_id" json:"userId"`
	ArticleID  string             `bson:"article_id,omitempty" json:"articleId,omitempty"`
	Operation  string             `bson:"operation" json:"operation"` // spell_check, translate, edit, detect_violation
	Provider   string             `bson:"provider" json:"provider"`
	InputText  string             `bson:"input_text" json:"inputText"`
	OutputText string             `bson:"output_text,omitempty" json:"outputText,omitempty"`
	SourceLang string             `bson:"source_lang,omitempty" json:"sourceLang,omitempty"`
	TargetLang string             `bson:"target_lang,omitempty" json:"targetLang,omitempty"`
	TokensUsed int                `bson:"tokens_used,omitempty" json:"tokensUsed,omitempty"`
	Cost       float64            `bson:"cost,omitempty" json:"cost,omitempty"`
	Success    bool               `bson:"success" json:"success"`
	Error      string             `bson:"error,omitempty" json:"error,omitempty"`
	Duration   int64              `bson:"duration_ms" json:"durationMs"` // milliseconds
	CreatedAt  time.Time          `bson:"created_at" json:"createdAt"`
}

// SpellCheckResult represents spell checking result
type SpellCheckResult struct {
	Original    string            `json:"original"`
	Corrections []SpellCorrection `json:"corrections"`
	HasErrors   bool              `json:"hasErrors"`
}

// SpellCorrection represents a single correction
type SpellCorrection struct {
	Word        string   `json:"word"`
	Suggestions []string `json:"suggestions"`
	Position    int      `json:"position"`
	Length      int      `json:"length"`
	Type        string   `json:"type"` // spelling, grammar
}

// TranslationResult represents translation result
type TranslationResult struct {
	SourceText     string `json:"sourceText"`
	TranslatedText string `json:"translatedText"`
	SourceLang     string `json:"sourceLang"`
	TargetLang     string `json:"targetLang"`
	Provider       string `json:"provider"`
}

// ContentEditSuggestion represents AI-generated content improvement
type ContentEditSuggestion struct {
	Original   string   `json:"original"`
	Improved   string   `json:"improved"`
	Changes    []string `json:"changes"`    // List of improvements made
	Confidence float64  `json:"confidence"` // 0-1
}

// ViolationDetectionResult represents content violation detection
type ViolationDetectionResult struct {
	HasViolation  bool     `json:"hasViolation"`
	ViolationType string   `json:"violationType"` // hate_speech, violence, sexual_content, etc
	Confidence    float64  `json:"confidence"`
	Explanation   string   `json:"explanation"`
	Suggestions   []string `json:"suggestions"` // How to fix
}
