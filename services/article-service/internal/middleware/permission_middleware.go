package middleware

import (
	"context"
	"net/http"

	"github.com/vhvplatform/go-cms-service/services/article-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/article-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PermissionMiddleware provides permission checking middleware
type PermissionMiddleware struct {
	permissionRepo *repository.PermissionRepository
}

// NewPermissionMiddleware creates a new permission middleware
func NewPermissionMiddleware(permissionRepo *repository.PermissionRepository) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionRepo: permissionRepo,
	}
}

// RequireRole creates middleware that requires a specific role
func (m *PermissionMiddleware) RequireRole(requiredRole model.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value("userRole")
			if userRole == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			role, ok := userRole.(string)
			if !ok || !m.hasRole(model.Role(role), requiredRole) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CheckCategoryPermission checks if user has permission on a category
func (m *PermissionMiddleware) CheckCategoryPermission(categoryID primitive.ObjectID, requiredRole model.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value("userID")
			if userID == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			uid, ok := userID.(string)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check permission
			hasPermission, err := m.permissionRepo.CheckPermission(
				r.Context(),
				uid,
				model.ResourceTypeCategory,
				categoryID,
				requiredRole,
			)

			if err != nil || !hasPermission {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasRole checks if a user's role meets the required role
// Hierarchy: moderator > editor > writer
func (m *PermissionMiddleware) hasRole(userRole, requiredRole model.Role) bool {
	roleHierarchy := map[model.Role]int{
		model.RoleWriter:    1,
		model.RoleEditor:    2,
		model.RoleModerator: 3,
	}

	userLevel := roleHierarchy[userRole]
	requiredLevel := roleHierarchy[requiredRole]

	return userLevel >= requiredLevel
}

// WithCategoryID adds category ID to request context from request body or params
func WithCategoryID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This would extract category ID from request and add to context
		// Implementation depends on router
		next.ServeHTTP(w, r)
	})
}
