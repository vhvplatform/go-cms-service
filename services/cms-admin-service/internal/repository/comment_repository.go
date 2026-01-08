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

// CommentRepository handles comment data operations
type CommentRepository struct {
	collection          *mongo.Collection
	likeCollection      *mongo.Collection
	reportCollection    *mongo.Collection
	rateLimitCollection *mongo.Collection
	favoriteCollection  *mongo.Collection
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *mongo.Database) *CommentRepository {
	return &CommentRepository{
		collection:          db.Collection("comments"),
		likeCollection:      db.Collection("comment_likes"),
		reportCollection:    db.Collection("comment_reports"),
		rateLimitCollection: db.Collection("user_rate_limits"),
		favoriteCollection:  db.Collection("favorite_articles"),
	}
}

// Create creates a new comment
func (r *CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	comment.ID = primitive.NewObjectID()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()
	comment.Status = model.CommentStatusPending
	comment.LikeCount = 0
	comment.ReplyCount = 0

	_, err := r.collection.InsertOne(ctx, comment)
	return err
}

// FindByID finds a comment by ID
func (r *CommentRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Comment, error) {
	var comment model.Comment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&comment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &comment, nil
}

// FindByArticleID finds all approved comments for an article with sorting
func (r *CommentRepository) FindByArticleID(ctx context.Context, articleID primitive.ObjectID, sortBy string, page, limit int) ([]*model.Comment, int64, error) {
	filter := bson.M{
		"articleId": articleID,
		"status":    model.CommentStatusApproved,
		"parentId":  nil, // Only root comments
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Determine sort order
	var sortField string
	var sortOrder int
	switch sortBy {
	case "likes":
		sortField = "likeCount"
		sortOrder = -1
	case "newest":
		sortField = "createdAt"
		sortOrder = -1
	case "oldest":
		sortField = "createdAt"
		sortOrder = 1
	default:
		sortField = "likeCount"
		sortOrder = -1
	}

	opts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var comments []*model.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// FindRepliesByParentID finds all replies to a comment
func (r *CommentRepository) FindRepliesByParentID(ctx context.Context, parentID primitive.ObjectID, sortBy string) ([]*model.Comment, error) {
	filter := bson.M{
		"parentId": parentID,
		"status":   model.CommentStatusApproved,
	}

	var sortField string
	var sortOrder int
	switch sortBy {
	case "likes":
		sortField = "likeCount"
		sortOrder = -1
	default:
		sortField = "createdAt"
		sortOrder = 1
	}

	opts := options.Find().SetSort(bson.D{{Key: sortField, Value: sortOrder}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []*model.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// Update updates a comment
func (r *CommentRepository) Update(ctx context.Context, comment *model.Comment) error {
	comment.UpdatedAt = time.Now()

	filter := bson.M{"_id": comment.ID}
	update := bson.M{"$set": comment}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// UpdateStatus updates comment status
func (r *CommentRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status model.CommentStatus, moderatorID string, note string) error {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":         status,
			"moderatedBy":    moderatorID,
			"moderatedAt":    now,
			"moderationNote": note,
			"updatedAt":      now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// IncrementReplyCount increments reply count for a comment
func (r *CommentRepository) IncrementReplyCount(ctx context.Context, id primitive.ObjectID) error {
	update := bson.M{
		"$inc": bson.M{"replyCount": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// LikeComment adds a like to a comment
func (r *CommentRepository) LikeComment(ctx context.Context, commentID primitive.ObjectID, userID string) error {
	// Check if already liked
	existing := r.likeCollection.FindOne(ctx, bson.M{
		"commentId": commentID,
		"userId":    userID,
	})
	if existing.Err() == nil {
		return nil // Already liked
	}

	// Create like
	like := &model.CommentLike{
		ID:        primitive.NewObjectID(),
		CommentID: commentID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if _, err := r.likeCollection.InsertOne(ctx, like); err != nil {
		return err
	}

	// Increment like count
	update := bson.M{
		"$inc": bson.M{"likeCount": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": commentID}, update)
	return err
}

// UnlikeComment removes a like from a comment
func (r *CommentRepository) UnlikeComment(ctx context.Context, commentID primitive.ObjectID, userID string) error {
	// Delete like
	result, err := r.likeCollection.DeleteOne(ctx, bson.M{
		"commentId": commentID,
		"userId":    userID,
	})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return nil // Not liked
	}

	// Decrement like count
	update := bson.M{
		"$inc": bson.M{"likeCount": -1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": commentID}, update)
	return err
}

// HasUserLiked checks if a user has liked a comment
func (r *CommentRepository) HasUserLiked(ctx context.Context, commentID primitive.ObjectID, userID string) (bool, error) {
	count, err := r.likeCollection.CountDocuments(ctx, bson.M{
		"commentId": commentID,
		"userId":    userID,
	})
	return count > 0, err
}

// ReportComment creates a comment report
func (r *CommentRepository) ReportComment(ctx context.Context, report *model.CommentReport) error {
	// Check if user already reported this comment
	existing := r.reportCollection.FindOne(ctx, bson.M{
		"commentId":  report.CommentID,
		"reporterId": report.ReporterID,
	})
	if existing.Err() == nil {
		return nil // Already reported
	}

	report.ID = primitive.NewObjectID()
	report.CreatedAt = time.Now()
	report.Status = "pending"

	if _, err := r.reportCollection.InsertOne(ctx, report); err != nil {
		return err
	}

	// Increment report count and mark as reported
	update := bson.M{
		"$inc": bson.M{"reportCount": 1},
		"$set": bson.M{
			"isReported": true,
			"updatedAt":  time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": report.CommentID}, update)
	return err
}

// CheckRateLimit checks and updates rate limit for a user
func (r *CommentRepository) CheckRateLimit(ctx context.Context, userID string, maxCommentsPerHour int) (bool, error) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)

	// Find or create rate limit record
	var rateLimit model.UserRateLimit
	err := r.rateLimitCollection.FindOne(ctx, bson.M{
		"userId":      userID,
		"windowStart": bson.M{"$gte": oneHourAgo},
	}).Decode(&rateLimit)

	if err == mongo.ErrNoDocuments {
		// Create new rate limit window
		rateLimit = model.UserRateLimit{
			ID:           primitive.NewObjectID(),
			UserID:       userID,
			CommentCount: 1,
			WindowStart:  now,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		_, err := r.rateLimitCollection.InsertOne(ctx, rateLimit)
		return true, err
	}

	if err != nil {
		return false, err
	}

	// Check if limit exceeded
	if rateLimit.CommentCount >= maxCommentsPerHour {
		return false, nil
	}

	// Increment count
	update := bson.M{
		"$inc": bson.M{"commentCount": 1},
		"$set": bson.M{"updatedAt": now},
	}

	_, err = r.rateLimitCollection.UpdateOne(ctx, bson.M{"_id": rateLimit.ID}, update)
	return true, err
}

// AddFavorite adds an article to user's favorites
func (r *CommentRepository) AddFavorite(ctx context.Context, userID string, articleID primitive.ObjectID) error {
	// Check if already favorited
	existing := r.favoriteCollection.FindOne(ctx, bson.M{
		"userId":    userID,
		"articleId": articleID,
	})
	if existing.Err() == nil {
		return nil // Already favorited
	}

	favorite := &model.FavoriteArticle{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		ArticleID: articleID,
		CreatedAt: time.Now(),
	}

	_, err := r.favoriteCollection.InsertOne(ctx, favorite)
	return err
}

// RemoveFavorite removes an article from user's favorites
func (r *CommentRepository) RemoveFavorite(ctx context.Context, userID string, articleID primitive.ObjectID) error {
	_, err := r.favoriteCollection.DeleteOne(ctx, bson.M{
		"userId":    userID,
		"articleId": articleID,
	})
	return err
}

// GetUserFavorites gets all favorited articles for a user
func (r *CommentRepository) GetUserFavorites(ctx context.Context, userID string, page, limit int) ([]primitive.ObjectID, int64, error) {
	filter := bson.M{"userId": userID}

	// Count total
	total, err := r.favoriteCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.favoriteCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var favorites []*model.FavoriteArticle
	if err := cursor.All(ctx, &favorites); err != nil {
		return nil, 0, err
	}

	articleIDs := make([]primitive.ObjectID, len(favorites))
	for i, fav := range favorites {
		articleIDs[i] = fav.ArticleID
	}

	return articleIDs, total, nil
}

// IsFavorited checks if a user has favorited an article
func (r *CommentRepository) IsFavorited(ctx context.Context, userID string, articleID primitive.ObjectID) (bool, error) {
	count, err := r.favoriteCollection.CountDocuments(ctx, bson.M{
		"userId":    userID,
		"articleId": articleID,
	})
	return count > 0, err
}

// FindPendingComments finds all comments pending moderation
func (r *CommentRepository) FindPendingComments(ctx context.Context, page, limit int) ([]*model.Comment, int64, error) {
	filter := bson.M{"status": model.CommentStatusPending}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: 1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var comments []*model.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}
