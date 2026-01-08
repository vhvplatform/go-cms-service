package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/service"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/util"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArticleHandler handles HTTP requests for articles
type ArticleHandler struct {
	service *service.ArticleService
}

// NewArticleHandler creates a new article handler
func NewArticleHandler(service *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{
		service: service,
	}
}

// CreateArticle handles POST /api/v1/articles
func (h *ArticleHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var article model.Article
	if err := json.NewDecoder(r.Body).Decode(&article); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	validator := validator.NewArticleValidator()
	if err := validator.ValidateCreate(&article); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserID(r)

	if err := h.service.Create(r.Context(), &article, userID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, article)
}

// GetArticle handles GET /api/v1/articles/{id}
func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	article, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, article)
}

// ListArticles handles GET /api/v1/articles
func (h *ArticleHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
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

	if status := query.Get("status"); status != "" {
		filter["status"] = model.ArticleStatus(status)
	}

	if q := query.Get("q"); q != "" {
		filter["q"] = q
	}

	if eventLineID := query.Get("eventLineId"); eventLineID != "" {
		if id, err := primitive.ObjectIDFromHex(eventLineID); err == nil {
			filter["eventLineId"] = id
		}
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

	articles, total, err := h.service.FindAll(r.Context(), filter, page, limit, sort)
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

// UpdateArticle handles PATCH /api/v1/articles/{id}
func (h *ArticleHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	article, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	if err := json.NewDecoder(r.Body).Decode(article); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID := getUserID(r)
	userRole := getUserRole(r)
	if err := h.service.Update(r.Context(), article, userID, userRole); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, article)
}

// DeleteArticle handles DELETE /api/v1/articles/{id}
func (h *ArticleHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	// Get user ID from context
	userID := getUserID(r)

	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Article deleted successfully"})
}

// PublishArticle handles POST /api/v1/articles/{id}/publish
func (h *ArticleHandler) PublishArticle(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	userID := getUserID(r)
	userRole := getUserRole(r)
	if err := h.service.Publish(r.Context(), id, userID, userRole); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Article published successfully"})
}

// ReorderArticles handles POST /api/v1/articles/reorder
func (h *ArticleHandler) ReorderArticles(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Articles []struct {
			ID       string `json:"id"`
			Ordering int    `json:"ordering"`
		} `json:"articles"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert to internal format
	articles := make([]struct {
		ID       primitive.ObjectID
		Ordering int
	}, len(request.Articles))

	for i, item := range request.Articles {
		id, err := primitive.ObjectIDFromHex(item.ID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid article ID")
			return
		}
		articles[i].ID = id
		articles[i].Ordering = item.Ordering
	}

	if err := h.service.Reorder(r.Context(), articles); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Articles reordered successfully"})
}

// GetArticleStats handles GET /api/v1/statistics/articles/{id}
func (h *ArticleHandler) GetArticleStats(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	query := r.URL.Query()

	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if start := query.Get("startDate"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startDate = t
		}
	}

	if end := query.Get("endDate"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			endDate = t
		}
	}

	stats, err := h.service.GetArticleStats(r.Context(), id, startDate, endDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// SearchArticles handles GET /api/v1/search
func (h *ArticleHandler) SearchArticles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	q := query.Get("q")

	if q == "" {
		respondError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	articles, total, err := h.service.Search(r.Context(), q, page, limit)
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

// ViewArticle handles POST /api/v1/articles/{id}/view
func (h *ArticleHandler) ViewArticle(w http.ResponseWriter, r *http.Request) {
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

// RejectArticle handles POST /api/v1/articles/{id}/reject
func (h *ArticleHandler) RejectArticle(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	var req struct {
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	v := validator.NewArticleValidator()
	if err := v.ValidateReject(id, req.Note); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := getUserID(r)
	userName := getUserName(r)
	userRole := getUserRole(r)

	if err := h.service.RejectArticle(r.Context(), id, userID, userName, userRole, req.Note); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Article rejected"})
}

// AddRejectionNote handles POST /api/v1/articles/{id}/rejection-notes
func (h *ArticleHandler) AddRejectionNote(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	var req struct {
		Note     string  `json:"note"`
		ParentID *string `json:"parentId,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	v := validator.NewArticleValidator()
	if err := v.ValidateReject(id, req.Note); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := getUserID(r)
	userName := getUserName(r)
	userRole := getUserRole(r)

	var parentID *primitive.ObjectID
	if req.ParentID != nil && *req.ParentID != "" {
		objID, err := primitive.ObjectIDFromHex(*req.ParentID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid parent ID")
			return
		}
		parentID = &objID
	}

	if err := h.service.AddRejectionNote(r.Context(), id, userID, userName, userRole, req.Note, parentID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"message": "Note added successfully"})
}

// GetRejectionNotes handles GET /api/v1/articles/{id}/rejection-notes
func (h *ArticleHandler) GetRejectionNotes(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	notes, err := h.service.GetRejectionNotes(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, notes)
}

// ResolveRejectionNotes handles POST /api/v1/articles/{id}/rejection-notes/resolve
func (h *ArticleHandler) ResolveRejectionNotes(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	userID := getUserID(r)
	userRole := getUserRole(r)

	if err := h.service.ResolveRejectionNotes(r.Context(), id, userID, userRole); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Rejection notes resolved"})
}

// GetArticleVersions handles GET /api/v1/articles/{id}/versions
func (h *ArticleHandler) GetArticleVersions(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	versions, err := h.service.GetArticleVersions(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, versions)
}

// GetArticleVersion handles GET /api/v1/articles/{id}/versions/{versionId}
func (h *ArticleHandler) GetArticleVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := getIDFromPath(r, "versionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid version ID")
		return
	}

	version, err := h.service.GetArticleVersion(r.Context(), versionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, version)
}

// RestoreArticleVersion handles POST /api/v1/articles/{id}/versions/{versionNum}/restore
func (h *ArticleHandler) RestoreArticleVersion(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	// Extract version number from path
	versionNumStr := r.URL.Query().Get("version")
	if versionNumStr == "" {
		respondError(w, http.StatusBadRequest, "Version number is required")
		return
	}

	versionNum, err := strconv.Atoi(versionNumStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid version number")
		return
	}

	userID := getUserID(r)
	userRole := getUserRole(r)

	if err := h.service.RestoreArticleVersion(r.Context(), id, versionNum, userID, userRole); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Version restored successfully"})
}

// GetActionLogs handles GET /api/v1/articles/{id}/logs
func (h *ArticleHandler) GetActionLogs(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	// Pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logs, total, err := h.service.GetActionLogs(r.Context(), id, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetSocialShareURLs handles GET /api/v1/articles/{id}/share
func (h *ArticleHandler) GetSocialShareURLs(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	article, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get base URL from environment or request
	baseURL := r.URL.Scheme + "://" + r.Host
	if baseURL == "://" {
		baseURL = "https://example.com" // Default for development
	}

	sharer := util.NewSocialMediaSharer(baseURL)
	shareURLs := sharer.GenerateShareURLs(article)

	respondJSON(w, http.StatusOK, shareURLs)
}
