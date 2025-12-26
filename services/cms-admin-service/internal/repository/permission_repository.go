package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PermissionRepository handles permission data operations
type PermissionRepository struct {
	collection *mongo.Collection
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *mongo.Database) *PermissionRepository {
	return &PermissionRepository{
		collection: db.Collection("permissions"),
	}
}

// Create creates a new permission
func (r *PermissionRepository) Create(ctx context.Context, permission *model.Permission) error {
	permission.ID = primitive.NewObjectID()
	permission.CreatedAt = time.Now()
	permission.UpdatedAt = time.Now()
	
	_, err := r.collection.InsertOne(ctx, permission)
	return err
}

// FindByID finds a permission by ID
func (r *PermissionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Permission, error) {
	var permission model.Permission
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&permission)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, err
	}
	return &permission, nil
}

// Update updates a permission
func (r *PermissionRepository) Update(ctx context.Context, permission *model.Permission) error {
	permission.UpdatedAt = time.Now()
	
	filter := bson.M{"_id": permission.ID}
	update := bson.M{"$set": permission}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete deletes a permission
func (r *PermissionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// FindByResource finds permissions for a specific resource
func (r *PermissionRepository) FindByResource(ctx context.Context, resourceType model.ResourceType, resourceID primitive.ObjectID) ([]*model.Permission, error) {
	filter := bson.M{
		"resourceType": resourceType,
		"resourceId":   resourceID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []*model.Permission
	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

// FindByUserAndResource finds permissions for a user on a specific resource
func (r *PermissionRepository) FindByUserAndResource(ctx context.Context, userID string, resourceType model.ResourceType, resourceID primitive.ObjectID) ([]*model.Permission, error) {
	filter := bson.M{
		"resourceType": resourceType,
		"resourceId":   resourceID,
		"userId":       userID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []*model.Permission
	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

// CheckPermission checks if a user has a specific role on a resource
func (r *PermissionRepository) CheckPermission(ctx context.Context, userID string, resourceType model.ResourceType, resourceID primitive.ObjectID, requiredRole model.Role) (bool, error) {
	filter := bson.M{
		"resourceType": resourceType,
		"resourceId":   resourceID,
		"userId":       userID,
		"role":         requiredRole,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CreateIndexes creates necessary indexes for the permissions collection
func (r *PermissionRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "resourceType", Value: 1},
				{Key: "resourceId", Value: 1},
				{Key: "userId", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "resourceType", Value: 1},
				{Key: "resourceId", Value: 1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
