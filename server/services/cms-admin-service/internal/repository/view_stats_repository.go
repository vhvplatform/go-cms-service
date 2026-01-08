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

// ViewStatsRepository handles view statistics operations
type ViewStatsRepository struct {
	collection *mongo.Collection
}

// NewViewStatsRepository creates a new view stats repository
func NewViewStatsRepository(db *mongo.Database) *ViewStatsRepository {
	return &ViewStatsRepository{
		collection: db.Collection("article_views"),
	}
}

// RecordView records a view for an article on a specific date
func (r *ViewStatsRepository) RecordView(ctx context.Context, articleID primitive.ObjectID, date time.Time) error {
	// Normalize date to start of day
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	filter := bson.M{
		"articleId": articleID,
		"date":      date,
	}

	update := bson.M{
		"$inc": bson.M{"views": 1},
		"$set": bson.M{"updatedAt": time.Now()},
		"$setOnInsert": bson.M{
			"articleId": articleID,
			"date":      date,
			"createdAt": time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetArticleStats gets view statistics for a specific article
func (r *ViewStatsRepository) GetArticleStats(ctx context.Context, articleID primitive.ObjectID, startDate, endDate time.Time) ([]*model.ArticleView, error) {
	filter := bson.M{
		"articleId": articleID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	opts := options.Find().SetSort(bson.M{"date": 1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stats []*model.ArticleView
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// GetCategoryStats gets aggregated view statistics for a category
func (r *ViewStatsRepository) GetCategoryStats(ctx context.Context, categoryID primitive.ObjectID, startDate, endDate time.Time) (int, error) {
	// This would require joining with articles collection
	// For now, returning a simple aggregation
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"date": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":        nil,
			"totalViews": bson.M{"$sum": "$views"},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return 0, err
	}

	if len(result) > 0 {
		if total, ok := result[0]["totalViews"].(int32); ok {
			return int(total), nil
		}
		if total, ok := result[0]["totalViews"].(int64); ok {
			return int(total), nil
		}
	}

	return 0, nil
}

// CreateIndexes creates necessary indexes for the article_views collection
func (r *ViewStatsRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "articleId", Value: 1},
				{Key: "date", Value: -1},
			},
		},
		{
			Keys: bson.D{{Key: "date", Value: -1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
