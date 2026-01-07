package handler

import (
	"net/http"
	"strconv"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PublicArticleHandler handles HTTP requests for public article APIs (user-facing)
type PublicArticleHandler struct {
	service *service.PublicArticleService
}

// NewPublicArticleHandler creates a new public article handler
func NewPublicArticleHandler(service *service.PublicArticleService) *PublicArticleHandler {
	return &PublicArticleHandler{
		service: service,
	}
}

// GetArticle handles GET /api/v1/public/articles/{id}
func (h *PublicArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	article, err := h.service.GetArticleByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, article)
}

// GetArticleBySlug handles GET /api/v1/public/articles/slug/{slug}
func (h *PublicArticleHandler) GetArticleBySlug(w http.ResponseWriter, r *http.Request) {
	// Extract slug from path - simplified, use proper router in production
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		respondError(w, http.StatusBadRequest, "Slug is required")
		return
	}

	article, err := h.service.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, article)
}

// ListArticles handles GET /api/v1/public/articles
func (h *PublicArticleHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Parse pagination parameters
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Build filter
	filter := make(map[string]interface{})

	if categoryID := query.Get("categoryId"); categoryID != "" {
		if id, err := primitive.ObjectIDFromHex(categoryID); err == nil {
			filter["categoryId"] = id
		}
	}

	if articleType := query.Get("articleType"); articleType != "" {
		filter["articleType"] = model.ArticleType(articleType)
	}

	if eventStreamID := query.Get("eventStreamId"); eventStreamID != "" {
		if id, err := primitive.ObjectIDFromHex(eventStreamID); err == nil {
			filter["eventStreamId"] = id
		}
	}

	if featured := query.Get("featured"); featured == "true" {
		filter["featured"] = true
	}

	if hot := query.Get("hot"); hot == "true" {
		filter["hot"] = true
	}

	// Tags filter
	if tags := query.Get("tags"); tags != "" {
		filter["tags"] = tags
	}

	// Parse sort
	sort := make(map[string]int)
	if sortBy := query.Get("sort"); sortBy != "" {
		if sortBy[0] == '-' {
			sort[sortBy[1:]] = -1
		} else {
			sort[sortBy] = 1
		}
	}

	articles, total, err := h.service.ListPublicArticles(r.Context(), filter, page, limit, sort)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"data":  articles,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	respondJSON(w, http.StatusOK, response)
}

// ViewArticle handles POST /api/v1/public/articles/{id}/view
func (h *PublicArticleHandler) ViewArticle(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	if err := h.service.IncrementViewCount(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "View recorded"})
}
