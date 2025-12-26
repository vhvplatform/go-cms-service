package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/vhvplatform/go-cms-service/services/cms-stats-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-stats-service/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentHandler handles HTTP requests for comments
type CommentHandler struct {
	service *service.CommentService
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(service *service.CommentService) *CommentHandler {
	return &CommentHandler{
		service: service,
	}
}

// CreateComment handles POST /api/v1/articles/{id}/comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	articleID, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	var req struct {
		Content  string              `json:"content"`
		ParentID *string             `json:"parentId,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID := getUserID(r)
	userName := getUserName(r)

	comment := &model.Comment{
		ArticleID: articleID,
		Content:   req.Content,
	}

	if req.ParentID != nil && *req.ParentID != "" {
		parentObjID, err := primitive.ObjectIDFromHex(*req.ParentID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid parent comment ID")
			return
		}
		comment.ParentID = &parentObjID
	}

	if err := h.service.CreateComment(r.Context(), comment, userID, userName); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, comment)
}

// GetArticleComments handles GET /api/v1/articles/{id}/comments
func (h *CommentHandler) GetArticleComments(w http.ResponseWriter, r *http.Request) {
	articleID, err := getIDFromPath(r, "id")
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

	// Sorting
	sortBy := r.URL.Query().Get("sortBy")
	if sortBy == "" {
		sortBy = "likes" // Default sort by likes
	}

	comments, total, err := h.service.GetArticleComments(r.Context(), articleID, sortBy, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"comments": comments,
		"total":    total,
		"page":     page,
		"limit":    limit,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetCommentReplies handles GET /api/v1/comments/{id}/replies
func (h *CommentHandler) GetCommentReplies(w http.ResponseWriter, r *http.Request) {
	commentID, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	sortBy := r.URL.Query().Get("sortBy")
	if sortBy == "" {
		sortBy = "newest"
	}

	replies, err := h.service.GetCommentReplies(r.Context(), commentID, sortBy)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, replies)
}

// ModerateComment handles POST /api/v1/comments/{id}/moderate
func (h *CommentHandler) ModerateComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	var req struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	status := model.CommentStatus(req.Status)
	if status != model.CommentStatusApproved && status != model.CommentStatusRejected {
		respondError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	userID := getUserID(r)
	userRole := getUserRole(r)

	if err := h.service.ModerateComment(r.Context(), commentID, status, userID, req.Note, userRole); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Comment moderated successfully"})
}

// LikeComment handles POST /api/v1/comments/{id}/like
func (h *CommentHandler) LikeComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	userID := getUserID(r)

	if err := h.service.LikeComment(r.Context(), commentID, userID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Comment liked"})
}

// UnlikeComment handles DELETE /api/v1/comments/{id}/like
func (h *CommentHandler) UnlikeComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	userID := getUserID(r)

	if err := h.service.UnlikeComment(r.Context(), commentID, userID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Comment unliked"})
}

// ReportComment handles POST /api/v1/comments/{id}/report
func (h *CommentHandler) ReportComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID := getUserID(r)

	if err := h.service.ReportComment(r.Context(), commentID, userID, req.Reason); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Comment reported"})
}

// AddFavorite handles POST /api/v1/articles/{id}/favorite
func (h *CommentHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	articleID, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	userID := getUserID(r)

	if err := h.service.AddFavorite(r.Context(), userID, articleID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Article added to favorites"})
}

// RemoveFavorite handles DELETE /api/v1/articles/{id}/favorite
func (h *CommentHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	articleID, err := getIDFromPath(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	userID := getUserID(r)

	if err := h.service.RemoveFavorite(r.Context(), userID, articleID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Article removed from favorites"})
}

// GetUserFavorites handles GET /api/v1/users/favorites
func (h *CommentHandler) GetUserFavorites(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	// Pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	articleIDs, total, err := h.service.GetUserFavorites(r.Context(), userID, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"articleIds": articleIDs,
		"total":      total,
		"page":       page,
		"limit":      limit,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetPendingComments handles GET /api/v1/comments/pending
func (h *CommentHandler) GetPendingComments(w http.ResponseWriter, r *http.Request) {
	// Pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	comments, total, err := h.service.GetPendingComments(r.Context(), page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"comments": comments,
		"total":    total,
		"page":     page,
		"limit":    limit,
	}

	respondJSON(w, http.StatusOK, response)
}
