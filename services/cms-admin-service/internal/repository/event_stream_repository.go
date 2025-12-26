package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventStreamRepository handles event stream data operations
type EventStreamRepository struct {
	collection *mongo.Collection
}

// NewEventStreamRepository creates a new event stream repository
func NewEventStreamRepository(db *mongo.Database) *EventStreamRepository {
	return &EventStreamRepository{
		collection: db.Collection("event_streams"),
	}
}

// Create creates a new event stream
func (r *EventStreamRepository) Create(ctx context.Context, eventStream *model.EventStream) error {
	eventStream.ID = primitive.NewObjectID()
	eventStream.CreatedAt = time.Now()
	eventStream.UpdatedAt = time.Now()
	
	_, err := r.collection.InsertOne(ctx, eventStream)
	return err
}

// FindByID finds an event stream by ID
func (r *EventStreamRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.EventStream, error) {
	var eventStream model.EventStream
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&eventStream)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("event stream not found")
		}
		return nil, err
	}
	return &eventStream, nil
}

// FindBySlug finds an event stream by slug
func (r *EventStreamRepository) FindBySlug(ctx context.Context, slug string) (*model.EventStream, error) {
	var eventStream model.EventStream
	err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&eventStream)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("event stream not found")
		}
		return nil, err
	}
	return &eventStream, nil
}

// Update updates an event stream
func (r *EventStreamRepository) Update(ctx context.Context, eventStream *model.EventStream) error {
	eventStream.UpdatedAt = time.Now()
	
	filter := bson.M{"_id": eventStream.ID}
	update := bson.M{"$set": eventStream}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete deletes an event stream
func (r *EventStreamRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// FindAll finds all event streams
func (r *EventStreamRepository) FindAll(ctx context.Context, page, limit int) ([]*model.EventStream, int64, error) {
	// Count total
	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	// Setup pagination
	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"startAt": -1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var eventStreams []*model.EventStream
	if err := cursor.All(ctx, &eventStreams); err != nil {
		return nil, 0, err
	}

	return eventStreams, total, nil
}

// CreateIndexes creates necessary indexes for the event_streams collection
func (r *EventStreamRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "startAt", Value: -1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
