package repository

import (
	"context"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AIConfigRepository struct {
	collection *mongo.Collection
}

func NewAIConfigRepository(db *mongo.Database) *AIConfigRepository {
	return &AIConfigRepository{
		collection: db.Collection("ai_configurations"),
	}
}

// GetByTenant retrieves AI configuration for a tenant
func (r *AIConfigRepository) GetByTenant(ctx context.Context, tenantID string) (*model.AIConfiguration, error) {
	var config model.AIConfiguration
	err := r.collection.FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// Upsert creates or updates AI configuration
func (r *AIConfigRepository) Upsert(ctx context.Context, config *model.AIConfiguration) error {
	now := time.Now()
	config.UpdatedAt = now
	if config.CreatedAt.IsZero() {
		config.CreatedAt = now
	}

	filter := bson.M{"tenant_id": config.TenantID}
	update := bson.M{"$set": config}
	opts := options.Update().SetUpsert(true)

	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	if result.UpsertedID != nil {
		config.ID = result.UpsertedID.(primitive.ObjectID)
	}
	return nil
}

type AIOperationLogRepository struct {
	collection *mongo.Collection
}

func NewAIOperationLogRepository(db *mongo.Database) *AIOperationLogRepository {
	return &AIOperationLogRepository{
		collection: db.Collection("ai_operation_logs"),
	}
}

// Create adds a new operation log
func (r *AIOperationLogRepository) Create(ctx context.Context, log *model.AIOperationLog) error {
	log.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		return err
	}
	log.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByTenant retrieves logs for a tenant with pagination
func (r *AIOperationLogRepository) GetByTenant(ctx context.Context, tenantID string, limit, skip int) ([]*model.AIOperationLog, error) {
	filter := bson.M{"tenant_id": tenantID}
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*model.AIOperationLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// GetUsageStats retrieves usage statistics for a tenant
func (r *AIOperationLogRepository) GetUsageStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	pipeline := []bson.M{
		{"$match": bson.M{
			"tenant_id": tenantID,
			"created_at": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
			"success": true,
		}},
		{"$group": bson.M{
			"_id":          "$operation",
			"count":        bson.M{"$sum": 1},
			"total_tokens": bson.M{"$sum": "$tokens_used"},
			"total_cost":   bson.M{"$sum": "$cost"},
			"avg_duration": bson.M{"$avg": "$duration_ms"},
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["operations"] = results

	return stats, nil
}
