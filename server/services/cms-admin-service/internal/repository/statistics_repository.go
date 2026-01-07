package repository

import (
	"context"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// StatisticsRepository handles statistics operations
type StatisticsRepository struct {
	db                *mongo.Database
	articleCollection *mongo.Collection
	viewCollection    *mongo.Collection
}

// NewStatisticsRepository creates a new statistics repository
func NewStatisticsRepository(db *mongo.Database) *StatisticsRepository {
	return &StatisticsRepository{
		db:                db,
		articleCollection: db.Collection("articles"),
		viewCollection:    db.Collection("article_views"),
	}
}

// GetArticleStatistics gets overall article statistics
func (r *StatisticsRepository) GetArticleStatistics(ctx context.Context, startDate, endDate time.Time, tenantID *primitive.ObjectID) (*model.ArticleStatistics, error) {
	stats := &model.ArticleStatistics{
		ArticlesByType:     make(map[string]int64),
		ArticlesByCategory: make(map[string]int64),
		ArticlesByAuthor:   make(map[string]int64),
		TopViewedArticles:  []*model.ArticleViewSummary{},
		Period: model.StatisticsPeriod{
			StartDate: startDate,
			EndDate:   endDate,
		},
		GeneratedAt: time.Now(),
	}

	// Build base filter
	baseFilter := bson.M{
		"createdAt": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	if tenantID != nil {
		baseFilter["tenantId"] = *tenantID
	}

	// Total articles
	total, err := r.articleCollection.CountDocuments(ctx, baseFilter)
	if err != nil {
		return nil, err
	}
	stats.TotalArticles = total

	// Count by status
	statusCounts := []bson.M{
		{"$match": baseFilter},
		{"$group": bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}},
	}
	cursor, err := r.articleCollection.Aggregate(ctx, statusCounts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		switch result.ID {
		case string(model.ArticleStatusPublished):
			stats.PublishedArticles = result.Count
		case string(model.ArticleStatusDraft):
			stats.DraftArticles = result.Count
		case string(model.ArticleStatusPendingReview):
			stats.PendingArticles = result.Count
		case string(model.ArticleStatusArchived):
			stats.ArchivedArticles = result.Count
		}
	}

	// Count by type
	typeCounts := []bson.M{
		{"$match": baseFilter},
		{"$group": bson.M{
			"_id":   "$articleType",
			"count": bson.M{"$sum": 1},
		}},
	}
	cursor, err = r.articleCollection.Aggregate(ctx, typeCounts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		stats.ArticlesByType[result.ID] = result.Count
	}

	// Count by category
	categoryCounts := []bson.M{
		{"$match": baseFilter},
		{"$group": bson.M{
			"_id":   "$categoryId",
			"count": bson.M{"$sum": 1},
		}},
	}
	cursor, err = r.articleCollection.Aggregate(ctx, categoryCounts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    primitive.ObjectID `bson:"_id"`
			Count int64              `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		stats.ArticlesByCategory[result.ID.Hex()] = result.Count
	}

	// Count by author
	authorCounts := []bson.M{
		{"$match": baseFilter},
		{"$group": bson.M{
			"_id":   "$createdBy",
			"count": bson.M{"$sum": 1},
		}},
	}
	cursor, err = r.articleCollection.Aggregate(ctx, authorCounts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		stats.ArticlesByAuthor[result.ID] = result.Count
	}

	// Total views
	viewPipeline := []bson.M{
		{"$match": bson.M{
			"date": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
		}},
		{"$group": bson.M{
			"_id":        nil,
			"totalViews": bson.M{"$sum": "$views"},
		}},
	}
	cursor, err = r.viewCollection.Aggregate(ctx, viewPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			TotalViews int64 `bson:"totalViews"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.TotalViews = result.TotalViews
		}
	}

	return stats, nil
}

// GetTopViewedArticles gets the most viewed articles
func (r *StatisticsRepository) GetTopViewedArticles(ctx context.Context, startDate, endDate time.Time, limit int, tenantID *primitive.ObjectID) ([]*model.ArticleViewSummary, error) {
	matchStage := bson.M{
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	pipeline := []bson.M{
		{"$match": matchStage},
		{"$group": bson.M{
			"_id":       "$articleId",
			"viewCount": bson.M{"$sum": "$views"},
		}},
		{"$sort": bson.M{"viewCount": -1}},
		{"$limit": limit},
		{"$lookup": bson.M{
			"from":         "articles",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "article",
		}},
		{"$unwind": "$article"},
	}

	if tenantID != nil {
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{"article.tenantId": *tenantID},
		})
	}

	cursor, err := r.viewCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*model.ArticleViewSummary
	for cursor.Next(ctx) {
		var result struct {
			ViewCount int64 `bson:"viewCount"`
			Article   struct {
				ID          primitive.ObjectID `bson:"_id"`
				Title       string             `bson:"title"`
				Slug        string             `bson:"slug"`
				ArticleType model.ArticleType  `bson:"articleType"`
				CategoryID  primitive.ObjectID `bson:"categoryId"`
			} `bson:"article"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}

		summary := &model.ArticleViewSummary{
			ArticleID:   result.Article.ID,
			Title:       result.Article.Title,
			Slug:        result.Article.Slug,
			ArticleType: result.Article.ArticleType,
			ViewCount:   result.ViewCount,
			CategoryID:  result.Article.CategoryID,
		}
		results = append(results, summary)
	}

	return results, nil
}

// GetCategoryStatistics gets statistics for categories
func (r *StatisticsRepository) GetCategoryStatistics(ctx context.Context, categoryID primitive.ObjectID, startDate, endDate time.Time) (*model.CategoryStatistics, error) {
	stats := &model.CategoryStatistics{
		CategoryID: categoryID,
	}

	// Count articles
	filter := bson.M{
		"categoryId": categoryID,
		"createdAt": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	total, err := r.articleCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.TotalArticles = total

	// Count published
	filter["status"] = model.ArticleStatusPublished
	published, err := r.articleCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.PublishedCount = published

	// Get category name
	var category struct {
		Name string `bson:"name"`
	}
	err = r.db.Collection("categories").FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
	if err == nil {
		stats.CategoryName = category.Name
	}

	// Get view statistics
	viewPipeline := []bson.M{
		{"$match": bson.M{
			"date": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
		}},
		{"$lookup": bson.M{
			"from":         "articles",
			"localField":   "articleId",
			"foreignField": "_id",
			"as":           "article",
		}},
		{"$unwind": "$article"},
		{"$match": bson.M{"article.categoryId": categoryID}},
		{"$group": bson.M{
			"_id":        nil,
			"totalViews": bson.M{"$sum": "$views"},
		}},
	}

	cursor, err := r.viewCollection.Aggregate(ctx, viewPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			TotalViews int64 `bson:"totalViews"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.TotalViews = result.TotalViews
			if stats.TotalArticles > 0 {
				stats.AverageViews = float64(stats.TotalViews) / float64(stats.TotalArticles)
			}
		}
	}

	stats.LastUpdated = time.Now()
	return stats, nil
}

// GetAuthorStatistics gets statistics for an author
func (r *StatisticsRepository) GetAuthorStatistics(ctx context.Context, authorID string, startDate, endDate time.Time) (*model.AuthorStatistics, error) {
	stats := &model.AuthorStatistics{
		AuthorID: authorID,
	}

	filter := bson.M{
		"createdBy": authorID,
		"createdAt": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	// Total articles
	total, err := r.articleCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.TotalArticles = total

	// Published count
	filter["status"] = model.ArticleStatusPublished
	published, err := r.articleCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.PublishedCount = published

	// Draft count
	filter["status"] = model.ArticleStatusDraft
	drafts, err := r.articleCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.DraftCount = drafts

	// Get author name
	var article struct {
		Author struct {
			Name string `bson:"name"`
		} `bson:"author"`
	}
	err = r.articleCollection.FindOne(ctx, bson.M{"createdBy": authorID}).Decode(&article)
	if err == nil {
		stats.AuthorName = article.Author.Name
	}

	// Total views
	pipeline := []bson.M{
		{"$match": bson.M{"createdBy": authorID}},
		{"$group": bson.M{
			"_id":        nil,
			"totalViews": bson.M{"$sum": "$viewCount"},
		}},
	}
	cursor, err := r.articleCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			TotalViews int64 `bson:"totalViews"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.TotalViews = result.TotalViews
			if stats.TotalArticles > 0 {
				stats.AverageViews = float64(stats.TotalViews) / float64(stats.TotalArticles)
			}
		}
	}

	return stats, nil
}

// GetViewTrend gets view trends over time
func (r *StatisticsRepository) GetViewTrend(ctx context.Context, startDate, endDate time.Time, tenantID *primitive.ObjectID) ([]*model.ViewTrendData, error) {
	matchStage := bson.M{
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	pipeline := []bson.M{
		{"$match": matchStage},
		{"$group": bson.M{
			"_id":       "$date",
			"viewCount": bson.M{"$sum": "$views"},
			"articles":  bson.M{"$addToSet": "$articleId"},
		}},
		{"$project": bson.M{
			"date":      "$_id",
			"viewCount": 1,
			"articles":  bson.M{"$size": "$articles"},
		}},
		{"$sort": bson.M{"date": 1}},
	}

	cursor, err := r.viewCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var trends []*model.ViewTrendData
	for cursor.Next(ctx) {
		var trend model.ViewTrendData
		if err := cursor.Decode(&trend); err != nil {
			continue
		}
		trends = append(trends, &trend)
	}

	return trends, nil
}
