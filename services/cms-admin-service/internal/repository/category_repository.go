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

// CategoryRepository handles category data operations
type CategoryRepository struct {
	collection *mongo.Collection
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *mongo.Database) *CategoryRepository {
	return &CategoryRepository{
		collection: db.Collection("categories"),
	}
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *model.Category) error {
	category.ID = primitive.NewObjectID()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, category)
	return err
}

// FindByID finds a category by ID
func (r *CategoryRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Category, error) {
	var category model.Category
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("category not found")
		}
		return nil, err
	}
	return &category, nil
}

// FindBySlug finds a category by slug
func (r *CategoryRepository) FindBySlug(ctx context.Context, slug string) (*model.Category, error) {
	var category model.Category
	err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("category not found")
		}
		return nil, err
	}
	return &category, nil
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, category *model.Category) error {
	category.UpdatedAt = time.Now()

	filter := bson.M{"_id": category.ID}
	update := bson.M{"$set": category}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// FindAll finds all categories
func (r *CategoryRepository) FindAll(ctx context.Context) ([]*model.Category, error) {
	opts := options.Find().SetSort(bson.M{"ordering": 1, "name": 1})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []*model.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

// FindByParentID finds categories by parent ID
func (r *CategoryRepository) FindByParentID(ctx context.Context, parentID *primitive.ObjectID) ([]*model.Category, error) {
	filter := bson.M{}
	if parentID == nil {
		filter["parentId"] = nil
	} else {
		filter["parentId"] = *parentID
	}

	opts := options.Find().SetSort(bson.M{"ordering": 1, "name": 1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []*model.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

// CreateIndexes creates necessary indexes for the categories collection
func (r *CategoryRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "parentId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "categoryType", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "articleType", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
