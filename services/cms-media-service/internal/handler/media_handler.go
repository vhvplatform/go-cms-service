package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/vhvplatform/go-cms-service/services/cms-media-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-media-service/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaHandler handles HTTP requests for media operations
type MediaHandler struct {
	service *service.MediaService
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(service *service.MediaService) *MediaHandler {
	return &MediaHandler{
		service: service,
	}
}

// UploadFile handles POST /api/v1/media/upload
func (h *MediaHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100MB max
		respondError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Get tenant ID from context/header
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	tenantID, err := primitive.ObjectIDFromHex(tenantIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	// Get user info
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "system"
	}

	folder := r.FormValue("folder")
	if folder == "" {
		folder = "/"
	}

	// Upload file
	mediaFile, err := h.service.UploadFile(
		r.Context(),
		file,
		fileHeader,
		tenantID,
		userID,
		folder,
		r.RemoteAddr,
		r.UserAgent(),
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, mediaFile)
}

// GetFile handles GET /api/v1/media/{id}
func (h *MediaHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	idStr := getIDFromPath(r.URL.Path)
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid file ID")
		return
	}

	file, err := h.service.GetFile(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "File not found")
		return
	}

	respondJSON(w, http.StatusOK, file)
}

// ListFiles handles GET /api/v1/media/files
func (h *MediaHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.URL.Query().Get("tenantId")
	tenantID, err := primitive.ObjectIDFromHex(tenantIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	folder := r.URL.Query().Get("folder")
	if folder == "" {
		folder = "/"
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	files, total, err := h.service.ListFiles(r.Context(), tenantID, folder, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"files": files,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	respondJSON(w, http.StatusOK, response)
}

// DeleteFile handles DELETE /api/v1/media/{id}
func (h *MediaHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	idStr := getIDFromPath(r.URL.Path)
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid file ID")
		return
	}

	userID := r.Header.Get("X-User-ID")
	role := r.Header.Get("X-User-Role")

	if err := h.service.DeleteFile(r.Context(), id, userID, role); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "File deleted successfully"})
}

// GetStorageUsage handles GET /api/v1/media/storage/{tenantId}
func (h *MediaHandler) GetStorageUsage(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := getIDFromPath(r.URL.Path)
	tenantID, err := primitive.ObjectIDFromHex(tenantIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	usage, err := h.service.GetStorageUsage(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, usage)
}

// CreateFolder handles POST /api/v1/media/folders
func (h *MediaHandler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	var folder model.Folder
	if err := json.NewDecoder(r.Body).Decode(&folder); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.CreateFolder(r.Context(), &folder); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, folder)
}

// ListFolders handles GET /api/v1/media/folders
func (h *MediaHandler) ListFolders(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.URL.Query().Get("tenantId")
	tenantID, err := primitive.ObjectIDFromHex(tenantIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	folders, err := h.service.ListFolders(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, folders)
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func getIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
