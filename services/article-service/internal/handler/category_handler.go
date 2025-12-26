package handler

import (
	"encoding/json"
	"net/http"

	"github.com/vhvplatform/go-cms-service/services/article-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/article-service/internal/service"
)

// CategoryHandler handles HTTP requests for categories
type CategoryHandler struct {
	service *service.CategoryService
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		service: service,
	}
}

// CreateCategory handles POST /api/v1/categories
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category model.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID := getUserID(r)
	if err := h.service.Create(r.Context(), &category, userID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, category)
}

// GetCategory handles GET /api/v1/categories/{id}
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	category, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, category)
}

// GetCategoryTree handles GET /api/v1/categories/tree
func (h *CategoryHandler) GetCategoryTree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.service.GetTree(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tree)
}

// UpdateCategory handles PATCH /api/v1/categories/{id}
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	category, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	if err := json.NewDecoder(r.Body).Decode(category); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.Update(r.Context(), category); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, category)
}

// DeleteCategory handles DELETE /api/v1/categories/{id}
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Category deleted successfully"})
}
