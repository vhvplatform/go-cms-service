package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SensitiveKeyword represents a sensitive keyword or pattern to detect in content
type SensitiveKeyword struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID    string             `bson:"tenant_id" json:"tenantId"`
	Keyword     string             `bson:"keyword" json:"keyword"`               // The keyword or regex pattern
	IsRegex     bool               `bson:"is_regex" json:"isRegex"`              // Whether it's a regex pattern
	Severity    string             `bson:"severity" json:"severity"`             // low, medium, high, critical
	Action      string             `bson:"action" json:"action"`                 // warn, block, review
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Category    string             `bson:"category,omitempty" json:"category,omitempty"` // e.g., profanity, violence, political
	IsActive    bool               `bson:"is_active" json:"isActive"`
	CreatedBy   string             `bson:"created_by" json:"createdBy"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}

// KeywordDetectionResult represents the result of keyword detection in content
type KeywordDetectionResult struct {
	Keyword      string   `json:"keyword"`
	Matches      []string `json:"matches"`      // Actual matched text
	Positions    []int    `json:"positions"`    // Position in content
	Severity     string   `json:"severity"`
	Action       string   `json:"action"`
	Category     string   `json:"category"`
	Description  string   `json:"description"`
}

// ContentScanResult represents the complete scan result
type ContentScanResult struct {
	HasViolations bool                     `json:"hasViolations"`
	Results       []KeywordDetectionResult `json:"results"`
	HighestSeverity string                 `json:"highestSeverity"`
	RecommendedAction string               `json:"recommendedAction"` // warn, block, review
	ScannedAt       time.Time              `json:"scannedAt"`
}
