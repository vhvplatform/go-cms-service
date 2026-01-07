package handler

import (
	"net/http"
	"strconv"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/service"
)

// RSSHandler handles RSS feed requests
type RSSHandler struct {
	service *service.RSSService
}

// NewRSSHandler creates a new RSS handler
func NewRSSHandler(service *service.RSSService) *RSSHandler {
	return &RSSHandler{
		service: service,
	}
}

// GetRSSFeed handles GET /api/v1/rss
func (h *RSSHandler) GetRSSFeed(w http.ResponseWriter, r *http.Request) {
	// Get limit from query param
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Get optional category filter
	categoryID := r.URL.Query().Get("categoryId")
	var categoryPtr *string
	if categoryID != "" {
		categoryPtr = &categoryID
	}

	// Generate RSS feed
	rss, err := h.service.GenerateFeed(r.Context(), limit, categoryPtr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set appropriate headers for RSS
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rss))
}
