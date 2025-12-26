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

// ArticleVersionRepository handles article version data operations
type ArticleVersionRepository struct {
	collection *mongo.Collection
}

// NewArticleVersionRepository creates a new article version repository
func NewArticleVersionRepository(db *mongo.Database) *ArticleVersionRepository {
	return &ArticleVersionRepository{
		collection: db.Collection("article_versions"),
	}
}

// Create creates a new article version
func (r *ArticleVersionRepository) Create(ctx context.Context, version *model.ArticleVersion) error {
	version.ID = primitive.NewObjectID()
	version.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, version)
	return err
}

// FindByArticleID finds all versions for a specific article
func (r *ArticleVersionRepository) FindByArticleID(ctx context.Context, articleID primitive.ObjectID) ([]*model.ArticleVersion, error) {
	filter := bson.M{"articleId": articleID}
	opts := options.Find().SetSort(bson.D{{Key: "versionNum", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var versions []*model.ArticleVersion
	if err := cursor.All(ctx, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// FindByID finds a specific version by ID
func (r *ArticleVersionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.ArticleVersion, error) {
	var version model.ArticleVersion
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&version)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &version, nil
}

// FindByVersionNumber finds a specific version by article ID and version number
func (r *ArticleVersionRepository) FindByVersionNumber(ctx context.Context, articleID primitive.ObjectID, versionNum int) (*model.ArticleVersion, error) {
	var version model.ArticleVersion
	filter := bson.M{
		"articleId":  articleID,
		"versionNum": versionNum,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&version)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &version, nil
}

// GetLatestVersionNumber gets the latest version number for an article
func (r *ArticleVersionRepository) GetLatestVersionNumber(ctx context.Context, articleID primitive.ObjectID) (int, error) {
	filter := bson.M{"articleId": articleID}
	opts := options.FindOne().SetSort(bson.D{{Key: "versionNum", Value: -1}})

	var version model.ArticleVersion
	err := r.collection.FindOne(ctx, filter, opts).Decode(&version)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil // No versions yet
		}
		return 0, err
	}

	return version.VersionNum, nil
}

// DeleteByArticleID deletes all versions for a specific article
func (r *ArticleVersionRepository) DeleteByArticleID(ctx context.Context, articleID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"articleId": articleID})
	return err
}

// CountByArticleID counts versions for a specific article
func (r *ArticleVersionRepository) CountByArticleID(ctx context.Context, articleID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"articleId": articleID})
}
