package middleware

import (
	"context"
	"net/http"
	"strings"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	// In production, this would integrate with the go-framework auth system
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// Authenticate validates the request and extracts user information
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// For development/testing, allow requests without auth
			ctx := context.WithValue(r.Context(), "userID", "guest")
			ctx = context.WithValue(ctx, "userRole", "writer")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// In production, validate token with JWT or framework auth
		// For now, simple mock validation
		userID, role := m.validateToken(token)
		if userID == "" {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "userRole", role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken validates a token and returns user information
// This is a mock implementation - in production, integrate with JWT/OAuth
func (m *AuthMiddleware) validateToken(token string) (string, string) {
	// Mock validation
	if token == "test-token" {
		return "user-123", "editor"
	}
	if token == "admin-token" {
		return "admin-456", "moderator"
	}
	return "", ""
}

// RequireAuth is a convenience middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID")
		if userID == nil || userID == "" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
