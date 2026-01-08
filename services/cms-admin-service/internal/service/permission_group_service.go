package service

import (
	"context"
	"fmt"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PermissionGroupService handles permission group business logic
type PermissionGroupService struct {
	repo *repository.PermissionGroupRepository
}

// NewPermissionGroupService creates a new permission group service
func NewPermissionGroupService(repo *repository.PermissionGroupRepository) *PermissionGroupService {
	return &PermissionGroupService{
		repo: repo,
	}
}

// Create creates a new permission group
func (s *PermissionGroupService) Create(ctx context.Context, group *model.PermissionGroup, creatorID string) error {
	// Validate
	if group.Name == "" {
		return fmt.Errorf("group name is required")
	}

	if len(group.CategoryIDs) == 0 {
		return fmt.Errorf("at least one category is required")
	}

	group.CreatedBy = creatorID
	return s.repo.Create(ctx, group)
}

// FindByID finds a permission group by ID
func (s *PermissionGroupService) FindByID(ctx context.Context, id primitive.ObjectID) (*model.PermissionGroup, error) {
	return s.repo.FindByID(ctx, id)
}

// Update updates a permission group
func (s *PermissionGroupService) Update(ctx context.Context, group *model.PermissionGroup) error {
	// Validate
	if group.Name == "" {
		return fmt.Errorf("group name is required")
	}

	if len(group.CategoryIDs) == 0 {
		return fmt.Errorf("at least one category is required")
	}

	return s.repo.Update(ctx, group)
}

// Delete deletes a permission group
func (s *PermissionGroupService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

// FindAll finds all permission groups
func (s *PermissionGroupService) FindAll(ctx context.Context, page, limit int) ([]*model.PermissionGroup, int64, error) {
	return s.repo.FindAll(ctx, page, limit)
}

// FindByUserID finds permission groups for a user
func (s *PermissionGroupService) FindByUserID(ctx context.Context, userID string) ([]*model.PermissionGroup, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// FindByCategoryID finds permission groups for a category
func (s *PermissionGroupService) FindByCategoryID(ctx context.Context, categoryID primitive.ObjectID) ([]*model.PermissionGroup, error) {
	return s.repo.FindByCategoryID(ctx, categoryID)
}

// CheckUserCategoryPermission checks if a user has permission on a category
func (s *PermissionGroupService) CheckUserCategoryPermission(ctx context.Context, userID string, categoryID primitive.ObjectID, requiredRole model.Role) (bool, error) {
	return s.repo.CheckUserCategoryPermission(ctx, userID, categoryID, requiredRole)
}

// GetUserCategoriesWithRole gets all categories a user has access to
func (s *PermissionGroupService) GetUserCategoriesWithRole(ctx context.Context, userID string) (map[primitive.ObjectID]model.Role, error) {
	return s.repo.GetUserCategoriesWithRole(ctx, userID)
}

// AddUserToGroup adds a user to a permission group
func (s *PermissionGroupService) AddUserToGroup(ctx context.Context, groupID primitive.ObjectID, userID string) error {
	return s.repo.AddUserToGroup(ctx, groupID, userID)
}

// RemoveUserFromGroup removes a user from a permission group
func (s *PermissionGroupService) RemoveUserFromGroup(ctx context.Context, groupID primitive.ObjectID, userID string) error {
	return s.repo.RemoveUserFromGroup(ctx, groupID, userID)
}

// AddCategoryToGroup adds a category to a permission group
func (s *PermissionGroupService) AddCategoryToGroup(ctx context.Context, groupID primitive.ObjectID, categoryID primitive.ObjectID) error {
	return s.repo.AddCategoryToGroup(ctx, groupID, categoryID)
}

// RemoveCategoryFromGroup removes a category from a permission group
func (s *PermissionGroupService) RemoveCategoryFromGroup(ctx context.Context, groupID primitive.ObjectID, categoryID primitive.ObjectID) error {
	return s.repo.RemoveCategoryFromGroup(ctx, groupID, categoryID)
}

// GetUserHighestRole gets the user's highest role across all their groups
func (s *PermissionGroupService) GetUserHighestRole(ctx context.Context, userID string) (model.Role, error) {
	groups, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	if len(groups) == 0 {
		return model.RoleWriter, nil // Default role
	}

	roleHierarchy := map[model.Role]int{
		model.RoleWriter:    1,
		model.RoleEditor:    2,
		model.RoleModerator: 3,
	}

	highestRole := model.RoleWriter
	highestLevel := 0

	for _, group := range groups {
		level := roleHierarchy[group.Role]
		if level > highestLevel {
			highestLevel = level
			highestRole = group.Role
		}
	}

	return highestRole, nil
}
