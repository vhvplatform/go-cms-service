package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Tenant represents a tenant/organization in the system
type Tenant struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Code        string             `json:"code" bson:"code"` // Unique tenant code
	Description string             `json:"description" bson:"description"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   string             `json:"createdBy" bson:"createdBy"`
}

// TenantArticleTypeConfig represents the configuration of article types allowed for a tenant
type TenantArticleTypeConfig struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID          primitive.ObjectID `json:"tenantId" bson:"tenantId"`
	AllowedTypes      []ArticleType      `json:"allowedTypes" bson:"allowedTypes"`           // Article types this tenant can use
	DisallowedTypes   []ArticleType      `json:"disallowedTypes" bson:"disallowedTypes"`     // Explicitly disallowed types
	AllowAllTypes     bool               `json:"allowAllTypes" bson:"allowAllTypes"`         // If true, all types are allowed
	CustomTypeConfigs []CustomTypeConfig `json:"customTypeConfigs" bson:"customTypeConfigs"` // Custom configurations per type
	CreatedAt         time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy         string             `json:"createdBy" bson:"createdBy"`
}

// CustomTypeConfig represents custom configuration for a specific article type
type CustomTypeConfig struct {
	ArticleType    ArticleType            `json:"articleType" bson:"articleType"`
	MaxArticles    int                    `json:"maxArticles" bson:"maxArticles"`       // Max number of articles of this type (0 = unlimited)
	RequiredFields []string               `json:"requiredFields" bson:"requiredFields"` // Additional required fields for this type
	CustomSettings map[string]interface{} `json:"customSettings" bson:"customSettings"` // Type-specific settings
}
