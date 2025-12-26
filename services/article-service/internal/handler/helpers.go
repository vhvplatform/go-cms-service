package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Response helpers

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Request helpers

func getIDFromPath(r *http.Request, param string) (primitive.ObjectID, error) {
	// Extract ID from URL path
	// This is a simplified version - in production use a proper router like chi or gorilla/mux
	parts := strings.Split(r.URL.Path, "/")
	var idStr string
	for i, part := range parts {
		if part == param && i > 0 {
			idStr = parts[i-1]
			break
		}
	}
	
	// If not found by param name, try last segment
	if idStr == "" && len(parts) > 0 {
		idStr = parts[len(parts)-1]
	}

	return primitive.ObjectIDFromHex(idStr)
}

func getUserID(r *http.Request) string {
	// Get user ID from context (set by auth middleware)
	if userID := r.Context().Value("userID"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return "system" // Default for testing
}
