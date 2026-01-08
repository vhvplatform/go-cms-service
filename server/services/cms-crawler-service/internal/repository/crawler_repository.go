package repository

import (
	"context"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-crawler-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CrawlerArticleRepository struct {
	collection *mongo.Collection
}

func NewCrawlerArticleRepository(db *mongo.Database) *CrawlerArticleRepository {
	return &CrawlerArticleRepository{
		collection: db.Collection("crawler_articles"),
	}
}

func (r *CrawlerArticleRepository) Create(ctx context.Context, article *model.CrawlerArticle) error {
	article.CrawledAt = time.Now()
	article.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, article)
	if err != nil {
		return err
	}
	article.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *CrawlerArticleRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.CrawlerArticle, error) {
	var article model.CrawlerArticle
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&article)
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *CrawlerArticleRepository) GetByTenant(ctx context.Context, tenantID string, status string, limit, skip int) ([]*model.CrawlerArticle, error) {
	filter := bson.M{"tenant_id": tenantID}
	if status != "" {
		filter["status"] = status
	}

	opts := options.Find().
		SetSort(bson.M{"crawled_at": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*model.CrawlerArticle
	if err := cursor.All(ctx, &articles); err != nil {
		return nil, err
	}
	return articles, nil
}

func (r *CrawlerArticleRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status, userID string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if status == "approved" {
		update["$set"].(bson.M)["approved_by"] = userID
		update["$set"].(bson.M)["approved_at"] = time.Now()
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *CrawlerArticleRepository) FindDuplicates(ctx context.Context, contentHash string) ([]*model.CrawlerArticle, error) {
	filter := bson.M{"content_hash": contentHash}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*model.CrawlerArticle
	if err := cursor.All(ctx, &articles); err != nil {
		return nil, err
	}
	return articles, nil
}

func (r *CrawlerArticleRepository) DeleteOldArticles(ctx context.Context, tenantID string, beforeDate time.Time) (int64, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"crawled_at": bson.M{"$lt": beforeDate},
		"status":     bson.M{"$in": []string{"pending", "rejected"}},
	}

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (r *CrawlerArticleRepository) GetStats(ctx context.Context, tenantID string) (*model.CrawlerStats, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"tenant_id": tenantID}},
		{"$group": bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := &model.CrawlerStats{TenantID: tenantID}
	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		status := result["_id"].(string)
		count := int(result["count"].(int32))

		switch status {
		case "pending":
			stats.PendingReview = count
		case "approved":
			stats.Approved = count
		case "rejected":
			stats.Rejected = count
		case "converted":
			stats.Converted = count
		}
		stats.TotalCrawled += count
	}

	return stats, nil
}

type CrawlerSourceRepository struct {
	collection *mongo.Collection
}

func NewCrawlerSourceRepository(db *mongo.Database) *CrawlerSourceRepository {
	return &CrawlerSourceRepository{
		collection: db.Collection("crawler_sources"),
	}
}

func (r *CrawlerSourceRepository) Create(ctx context.Context, source *model.CrawlerSource) error {
	source.CreatedAt = time.Now()
	source.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, source)
	if err != nil {
		return err
	}
	source.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *CrawlerSourceRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.CrawlerSource, error) {
	var source model.CrawlerSource
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&source)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *CrawlerSourceRepository) GetByTenant(ctx context.Context, tenantID string, activeOnly bool) ([]*model.CrawlerSource, error) {
	filter := bson.M{"tenant_id": tenantID}
	if activeOnly {
		filter["is_active"] = true
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sources []*model.CrawlerSource
	if err := cursor.All(ctx, &sources); err != nil {
		return nil, err
	}
	return sources, nil
}

func (r *CrawlerSourceRepository) Update(ctx context.Context, source *model.CrawlerSource) error {
	source.UpdatedAt = time.Now()
	update := bson.M{"$set": source}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": source.ID}, update)
	return err
}

func (r *CrawlerSourceRepository) UpdateLastCrawled(ctx context.Context, id primitive.ObjectID, success bool) error {
	update := bson.M{
		"$set": bson.M{"last_crawled_at": time.Now()},
		"$inc": bson.M{"total_crawled": 1},
	}

	if !success {
		update["$inc"].(bson.M)["total_errors"] = 1
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

type CrawlerCampaignRepository struct {
	collection *mongo.Collection
}

func NewCrawlerCampaignRepository(db *mongo.Database) *CrawlerCampaignRepository {
	return &CrawlerCampaignRepository{
		collection: db.Collection("crawler_campaigns"),
	}
}

func (r *CrawlerCampaignRepository) Create(ctx context.Context, campaign *model.CrawlerCampaign) error {
	campaign.CreatedAt = time.Now()
	campaign.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, campaign)
	if err != nil {
		return err
	}
	campaign.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *CrawlerCampaignRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.CrawlerCampaign, error) {
	var campaign model.CrawlerCampaign
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&campaign)
	if err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (r *CrawlerCampaignRepository) GetActiveCampaigns(ctx context.Context, tenantID string) ([]*model.CrawlerCampaign, error) {
	filter := bson.M{
		"tenant_id": tenantID,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var campaigns []*model.CrawlerCampaign
	if err := cursor.All(ctx, &campaigns); err != nil {
		return nil, err
	}
	return campaigns, nil
}

func (r *CrawlerCampaignRepository) UpdateLastRun(ctx context.Context, id primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"last_run_at": time.Now(),
			"updated_at":  time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}
