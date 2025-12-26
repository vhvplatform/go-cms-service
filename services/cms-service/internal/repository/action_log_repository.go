package repository

import (
	"context"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ActionLogRepository handles action log data operations
type ActionLogRepository struct {
	collection *mongo.Collection
}

// NewActionLogRepository creates a new action log repository
func NewActionLogRepository(db *mongo.Database) *ActionLogRepository {
	return &ActionLogRepository{
		collection: db.Collection("action_logs"),
	}
}

// Create creates a new action log entry
func (r *ActionLogRepository) Create(ctx context.Context, log *model.ActionLog) error {
	log.ID = primitive.NewObjectID()
	log.Timestamp = time.Now()
	
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

// FindByArticleID finds all action logs for a specific article
func (r *ActionLogRepository) FindByArticleID(ctx context.Context, articleID primitive.ObjectID, page, limit int) ([]*model.ActionLog, int64, error) {
	filter := bson.M{"articleId": articleID}
	
	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var logs []*model.ActionLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
}

// FindByUserID finds all action logs by a specific user
func (r *ActionLogRepository) FindByUserID(ctx context.Context, userID string, page, limit int) ([]*model.ActionLog, int64, error) {
	filter := bson.M{"userId": userID}
	
	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var logs []*model.ActionLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
}

// FindByActionType finds all action logs of a specific type
func (r *ActionLogRepository) FindByActionType(ctx context.Context, actionType model.ActionType, page, limit int) ([]*model.ActionLog, int64, error) {
	filter := bson.M{"actionType": actionType}
	
	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var logs []*model.ActionLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
}

// FindAll finds all action logs with filters
func (r *ActionLogRepository) FindAll(ctx context.Context, filter map[string]interface{}, page, limit int) ([]*model.ActionLog, int64, error) {
	bsonFilter := bson.M{}
	for k, v := range filter {
		bsonFilter[k] = v
	}
	
	// Count total
	total, err := r.collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		return nil, 0, err
	}
	
	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))
	
	cursor, err := r.collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var logs []*model.ActionLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
}
