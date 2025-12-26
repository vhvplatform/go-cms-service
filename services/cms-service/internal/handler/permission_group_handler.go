package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/vhvplatform/go-cms-service/services/cms-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-service/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PermissionGroupHandler handles HTTP requests for permission groups
type PermissionGroupHandler struct {
	service *service.PermissionGroupService
}

// NewPermissionGroupHandler creates a new permission group handler
func NewPermissionGroupHandler(service *service.PermissionGroupService) *PermissionGroupHandler {
	return &PermissionGroupHandler{
		service: service,
	}
}

// CreatePermissionGroup handles POST /api/v1/permission-groups
func (h *PermissionGroupHandler) CreatePermissionGroup(w http.ResponseWriter, r *http.Request) {
	var group model.PermissionGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID := getUserID(r)
	if err := h.service.Create(r.Context(), &group, userID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, group)
}

// GetPermissionGroup handles GET /api/v1/permission-groups/{id}
func (h *PermissionGroupHandler) GetPermissionGroup(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid permission group ID")
		return
	}

	group, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, group)
}

// ListPermissionGroups handles GET /api/v1/permission-groups
func (h *PermissionGroupHandler) ListPermissionGroups(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	groups, total, err := h.service.FindAll(r.Context(), page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"data":  groups,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	respondJSON(w, http.StatusOK, response)
}

// UpdatePermissionGroup handles PATCH /api/v1/permission-groups/{id}
func (h *PermissionGroupHandler) UpdatePermissionGroup(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid permission group ID")
		return
	}

	group, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	if err := json.NewDecoder(r.Body).Decode(group); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.Update(r.Context(), group); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, group)
}

// DeletePermissionGroup handles DELETE /api/v1/permission-groups/{id}
func (h *PermissionGroupHandler) DeletePermissionGroup(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid permission group ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Permission group deleted successfully"})
}

// AddUserToGroup handles POST /api/v1/permission-groups/{id}/users
func (h *PermissionGroupHandler) AddUserToGroup(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid permission group ID")
		return
	}

	var request struct {
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.AddUserToGroup(r.Context(), id, request.UserID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "User added to group successfully"})
}

// RemoveUserFromGroup handles DELETE /api/v1/permission-groups/{id}/users/{userId}
func (h *PermissionGroupHandler) RemoveUserFromGroup(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid permission group ID")
		return
	}

	// Extract userID from path - this is simplified, use proper router in production
	query := r.URL.Query()
	userID := query.Get("userId")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	if err := h.service.RemoveUserFromGroup(r.Context(), id, userID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "User removed from group successfully"})
}

// AddCategoryToGroup handles POST /api/v1/permission-groups/{id}/categories
func (h *PermissionGroupHandler) AddCategoryToGroup(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid permission group ID")
		return
	}

	var request struct {
		CategoryID string `json:"categoryId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	categoryID, err := primitive.ObjectIDFromHex(request.CategoryID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	if err := h.service.AddCategoryToGroup(r.Context(), id, categoryID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Category added to group successfully"})
}
