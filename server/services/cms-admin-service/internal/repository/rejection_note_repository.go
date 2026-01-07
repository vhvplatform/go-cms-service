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

// RejectionNoteRepository handles rejection note data operations
type RejectionNoteRepository struct {
	collection *mongo.Collection
}

// NewRejectionNoteRepository creates a new rejection note repository
func NewRejectionNoteRepository(db *mongo.Database) *RejectionNoteRepository {
	return &RejectionNoteRepository{
		collection: db.Collection("rejection_notes"),
	}
}

// Create creates a new rejection note
func (r *RejectionNoteRepository) Create(ctx context.Context, note *model.RejectionNote) error {
	note.ID = primitive.NewObjectID()
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, note)
	return err
}

// FindByArticleID finds all rejection notes for a specific article
func (r *RejectionNoteRepository) FindByArticleID(ctx context.Context, articleID primitive.ObjectID) ([]*model.RejectionNote, error) {
	filter := bson.M{"articleId": articleID}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}) // Oldest first for conversation flow

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notes []*model.RejectionNote
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, err
	}

	return notes, nil
}

// FindByID finds a specific rejection note by ID
func (r *RejectionNoteRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.RejectionNote, error) {
	var note model.RejectionNote
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&note)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &note, nil
}

// Update updates a rejection note
func (r *RejectionNoteRepository) Update(ctx context.Context, note *model.RejectionNote) error {
	note.UpdatedAt = time.Now()

	filter := bson.M{"_id": note.ID}
	update := bson.M{"$set": note}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// MarkAsResolved marks all rejection notes for an article as resolved
func (r *RejectionNoteRepository) MarkAsResolved(ctx context.Context, articleID primitive.ObjectID) error {
	filter := bson.M{"articleId": articleID}
	update := bson.M{
		"$set": bson.M{
			"isResolved": true,
			"updatedAt":  time.Now(),
		},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

// CountUnresolvedByArticleID counts unresolved rejection notes for an article
func (r *RejectionNoteRepository) CountUnresolvedByArticleID(ctx context.Context, articleID primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"articleId":  articleID,
		"isResolved": false,
	}
	return r.collection.CountDocuments(ctx, filter)
}

// FindUnresolvedByUserID finds all unresolved rejection notes for articles created by a user
func (r *RejectionNoteRepository) FindUnresolvedByUserID(ctx context.Context, userID string, page, limit int) ([]*model.RejectionNote, int64, error) {
	// This requires joining with articles collection to filter by createdBy
	// For now, we'll return all unresolved notes
	filter := bson.M{"isResolved": false}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var notes []*model.RejectionNote
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, 0, err
	}

	return notes, total, nil
}

// DeleteByArticleID deletes all rejection notes for an article
func (r *RejectionNoteRepository) DeleteByArticleID(ctx context.Context, articleID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"articleId": articleID})
	return err
}
