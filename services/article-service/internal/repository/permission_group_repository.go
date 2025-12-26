package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/vhvplatform/go-cms-service/services/article-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PermissionGroupRepository handles permission group data operations
type PermissionGroupRepository struct {
	collection *mongo.Collection
}

// NewPermissionGroupRepository creates a new permission group repository
func NewPermissionGroupRepository(db *mongo.Database) *PermissionGroupRepository {
	return &PermissionGroupRepository{
		collection: db.Collection("permission_groups"),
	}
}

// Create creates a new permission group
func (r *PermissionGroupRepository) Create(ctx context.Context, group *model.PermissionGroup) error {
	group.ID = primitive.NewObjectID()
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()
	
	_, err := r.collection.InsertOne(ctx, group)
	return err
}

// FindByID finds a permission group by ID
func (r *PermissionGroupRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.PermissionGroup, error) {
	var group model.PermissionGroup
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&group)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("permission group not found")
		}
		return nil, err
	}
	return &group, nil
}

// Update updates a permission group
func (r *PermissionGroupRepository) Update(ctx context.Context, group *model.PermissionGroup) error {
	group.UpdatedAt = time.Now()
	
	filter := bson.M{"_id": group.ID}
	update := bson.M{"$set": group}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete deletes a permission group
func (r *PermissionGroupRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// FindAll finds all permission groups
func (r *PermissionGroupRepository) FindAll(ctx context.Context, page, limit int) ([]*model.PermissionGroup, int64, error) {
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
		SetSort(bson.M{"name": 1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var groups []*model.PermissionGroup
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// FindByUserID finds permission groups that contain a specific user
func (r *PermissionGroupRepository) FindByUserID(ctx context.Context, userID string) ([]*model.PermissionGroup, error) {
	filter := bson.M{
		"userIds": userID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []*model.PermissionGroup
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

// FindByCategoryID finds permission groups that have access to a specific category
func (r *PermissionGroupRepository) FindByCategoryID(ctx context.Context, categoryID primitive.ObjectID) ([]*model.PermissionGroup, error) {
	filter := bson.M{
		"categoryIds": categoryID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []*model.PermissionGroup
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

// CheckUserCategoryPermission checks if a user has permission to access a category through groups
func (r *PermissionGroupRepository) CheckUserCategoryPermission(ctx context.Context, userID string, categoryID primitive.ObjectID, requiredRole model.Role) (bool, error) {
	// Find groups that contain this user and have access to this category
	filter := bson.M{
		"userIds":     userID,
		"categoryIds": categoryID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return false, err
	}
	defer cursor.Close(ctx)

	var groups []*model.PermissionGroup
	if err := cursor.All(ctx, &groups); err != nil {
		return false, err
	}

	// Check if any group has sufficient role
	roleHierarchy := map[model.Role]int{
		model.RoleWriter:    1,
		model.RoleEditor:    2,
		model.RoleModerator: 3,
	}

	requiredLevel := roleHierarchy[requiredRole]

	for _, group := range groups {
		groupLevel := roleHierarchy[group.Role]
		if groupLevel >= requiredLevel {
			return true, nil
		}
	}

	return false, nil
}

// GetUserCategoriesWithRole gets all categories a user has access to with their highest role
func (r *PermissionGroupRepository) GetUserCategoriesWithRole(ctx context.Context, userID string) (map[primitive.ObjectID]model.Role, error) {
	// Find all groups that contain this user
	filter := bson.M{
		"userIds": userID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []*model.PermissionGroup
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}

	// Build map of category -> highest role
	roleHierarchy := map[model.Role]int{
		model.RoleWriter:    1,
		model.RoleEditor:    2,
		model.RoleModerator: 3,
	}

	categoryRoles := make(map[primitive.ObjectID]model.Role)

	for _, group := range groups {
		for _, catID := range group.CategoryIDs {
			existingRole, exists := categoryRoles[catID]
			if !exists {
				categoryRoles[catID] = group.Role
			} else {
				// Keep the highest role
				if roleHierarchy[group.Role] > roleHierarchy[existingRole] {
					categoryRoles[catID] = group.Role
				}
			}
		}
	}

	return categoryRoles, nil
}

// AddUserToGroup adds a user to a permission group
func (r *PermissionGroupRepository) AddUserToGroup(ctx context.Context, groupID primitive.ObjectID, userID string) error {
	filter := bson.M{"_id": groupID}
	update := bson.M{
		"$addToSet": bson.M{"userIds": userID},
		"$set":      bson.M{"updatedAt": time.Now()},
	}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// RemoveUserFromGroup removes a user from a permission group
func (r *PermissionGroupRepository) RemoveUserFromGroup(ctx context.Context, groupID primitive.ObjectID, userID string) error {
	filter := bson.M{"_id": groupID}
	update := bson.M{
		"$pull": bson.M{"userIds": userID},
		"$set":  bson.M{"updatedAt": time.Now()},
	}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// AddCategoryToGroup adds a category to a permission group
func (r *PermissionGroupRepository) AddCategoryToGroup(ctx context.Context, groupID primitive.ObjectID, categoryID primitive.ObjectID) error {
	filter := bson.M{"_id": groupID}
	update := bson.M{
		"$addToSet": bson.M{"categoryIds": categoryID},
		"$set":      bson.M{"updatedAt": time.Now()},
	}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// RemoveCategoryFromGroup removes a category from a permission group
func (r *PermissionGroupRepository) RemoveCategoryFromGroup(ctx context.Context, groupID primitive.ObjectID, categoryID primitive.ObjectID) error {
	filter := bson.M{"_id": groupID}
	update := bson.M{
		"$pull": bson.M{"categoryIds": categoryID},
		"$set":  bson.M{"updatedAt": time.Now()},
	}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// CreateIndexes creates necessary indexes for the permission_groups collection
func (r *PermissionGroupRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "userIds", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "categoryIds", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "userIds", Value: 1},
				{Key: "categoryIds", Value: 1},
			},
		},
		{
			Keys: bson.D{{Key: "name", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
