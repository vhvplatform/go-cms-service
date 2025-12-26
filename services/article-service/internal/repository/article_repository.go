package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vhvplatform/go-cms-service/services/article-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ArticleRepository handles article data operations
type ArticleRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

// NewArticleRepository creates a new article repository
func NewArticleRepository(db *mongo.Database) *ArticleRepository {
	return &ArticleRepository{
		collection: db.Collection("articles"),
		db:         db,
	}
}

// Create creates a new article
func (r *ArticleRepository) Create(ctx context.Context, article *model.Article) error {
	article.ID = primitive.NewObjectID()
	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()
	
	_, err := r.collection.InsertOne(ctx, article)
	return err
}

// FindByID finds an article by ID
func (r *ArticleRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Article, error) {
	var article model.Article
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&article)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("article not found")
		}
		return nil, err
	}
	return &article, nil
}

// FindBySlug finds an article by slug
func (r *ArticleRepository) FindBySlug(ctx context.Context, slug string) (*model.Article, error) {
	var article model.Article
	err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&article)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("article not found")
		}
		return nil, err
	}
	return &article, nil
}

// Update updates an article
func (r *ArticleRepository) Update(ctx context.Context, article *model.Article) error {
	article.UpdatedAt = time.Now()
	
	filter := bson.M{"_id": article.ID}
	update := bson.M{"$set": article}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete soft deletes an article
func (r *ArticleRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":    model.ArticleStatusDeleted,
			"updatedAt": time.Now(),
		},
	}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// FindAll finds articles with filters and pagination
func (r *ArticleRepository) FindAll(ctx context.Context, filter map[string]interface{}, page, limit int, sort map[string]int) ([]*model.Article, int64, error) {
	// Build filter
	bsonFilter := bson.M{}
	for k, v := range filter {
		if k == "q" {
			// Full-text search
			bsonFilter["$text"] = bson.M{"$search": v}
		} else {
			bsonFilter[k] = v
		}
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		return nil, 0, err
	}

	// Setup pagination
	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	// Setup sorting
	if len(sort) > 0 {
		opts.SetSort(sort)
	} else {
		opts.SetSort(bson.M{"publishAt": -1}) // Default sort
	}

	// Execute query
	cursor, err := r.collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var articles []*model.Article
	if err := cursor.All(ctx, &articles); err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// UpdateStatus updates article status
func (r *ArticleRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status model.ArticleStatus, publishedBy string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":      status,
			"publishedBy": publishedBy,
			"updatedAt":   time.Now(),
		},
	}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// UpdateOrdering updates article ordering
func (r *ArticleRepository) UpdateOrdering(ctx context.Context, id primitive.ObjectID, ordering int) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"ordering":  ordering,
			"updatedAt": time.Now(),
		},
	}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// IncrementViewCount increments the view count for an article
func (r *ArticleRepository) IncrementViewCount(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$inc": bson.M{"viewCount": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// FindArticlesToPublish finds articles that should be published
func (r *ArticleRepository) FindArticlesToPublish(ctx context.Context) ([]*model.Article, error) {
	now := time.Now()
	filter := bson.M{
		"status":    model.ArticleStatusPendingReview,
		"publishAt": bson.M{"$lte": now},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*model.Article
	if err := cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

// FindArticlesToExpire finds articles that should be expired
func (r *ArticleRepository) FindArticlesToExpire(ctx context.Context) ([]*model.Article, error) {
	now := time.Now()
	filter := bson.M{
		"status":    model.ArticleStatusPublished,
		"expiredAt": bson.M{"$lte": now, "$ne": nil},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*model.Article
	if err := cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

// CalculateCharCount calculates character count from content
func (r *ArticleRepository) CalculateCharCount(content string) int {
	// Remove HTML tags and count characters
	cleaned := strings.ReplaceAll(content, "<", " <")
	cleaned = strings.ReplaceAll(cleaned, ">", "> ")
	
	// Simple character count (can be enhanced)
	return len(strings.TrimSpace(cleaned))
}

// CalculateImageCount calculates image count from content blocks
func (r *ArticleRepository) CalculateImageCount(blocks []model.ContentBlock) int {
	count := 0
	for _, block := range blocks {
		if block.Type == "image" {
			count++
		}
	}
	return count
}

// CreateIndexes creates necessary indexes for the articles collection
func (r *ArticleRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "title", Value: "text"},
				{Key: "summary", Value: "text"},
				{Key: "content", Value: "text"},
			},
		},
		{
			Keys: bson.D{{Key: "articleType", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "categoryId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "eventLineId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "publishAt", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tags", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "publishAt", Value: -1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
