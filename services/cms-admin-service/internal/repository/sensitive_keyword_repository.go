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

type SensitiveKeywordRepository struct {
	collection *mongo.Collection
}

func NewSensitiveKeywordRepository(db *mongo.Database) *SensitiveKeywordRepository {
	return &SensitiveKeywordRepository{
		collection: db.Collection("sensitive_keywords"),
	}
}

// Create adds a new sensitive keyword
func (r *SensitiveKeywordRepository) Create(ctx context.Context, keyword *model.SensitiveKeyword) error {
	keyword.CreatedAt = time.Now()
	keyword.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, keyword)
	if err != nil {
		return err
	}
	keyword.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a keyword by ID
func (r *SensitiveKeywordRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.SensitiveKeyword, error) {
	var keyword model.SensitiveKeyword
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&keyword)
	if err != nil {
		return nil, err
	}
	return &keyword, nil
}

// GetByTenant retrieves all keywords for a tenant
func (r *SensitiveKeywordRepository) GetByTenant(ctx context.Context, tenantID string, activeOnly bool) ([]*model.SensitiveKeyword, error) {
	filter := bson.M{"tenant_id": tenantID}
	if activeOnly {
		filter["is_active"] = true
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.M{"severity": -1, "created_at": -1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var keywords []*model.SensitiveKeyword
	if err := cursor.All(ctx, &keywords); err != nil {
		return nil, err
	}
	return keywords, nil
}

// GetByCategory retrieves keywords by category
func (r *SensitiveKeywordRepository) GetByCategory(ctx context.Context, tenantID, category string) ([]*model.SensitiveKeyword, error) {
	filter := bson.M{
		"tenant_id": tenantID,
		"category":  category,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var keywords []*model.SensitiveKeyword
	if err := cursor.All(ctx, &keywords); err != nil {
		return nil, err
	}
	return keywords, nil
}

// Update updates a keyword
func (r *SensitiveKeywordRepository) Update(ctx context.Context, keyword *model.SensitiveKeyword) error {
	keyword.UpdatedAt = time.Now()

	update := bson.M{
		"$set": keyword,
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": keyword.ID}, update)
	return err
}

// Delete deletes a keyword
func (r *SensitiveKeywordRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// BulkCreate creates multiple keywords at once
func (r *SensitiveKeywordRepository) BulkCreate(ctx context.Context, keywords []*model.SensitiveKeyword) error {
	if len(keywords) == 0 {
		return nil
	}

	now := time.Now()
	docs := make([]interface{}, len(keywords))
	for i, k := range keywords {
		k.CreatedAt = now
		k.UpdatedAt = now
		docs[i] = k
	}

	_, err := r.collection.InsertMany(ctx, docs)
	return err
}
