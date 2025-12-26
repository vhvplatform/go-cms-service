package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-frontend-service/internal/client"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cmsServiceURL := getEnv("CMS_SERVICE_URL", "http://localhost:8080")
	statsServiceURL := getEnv("STATS_SERVICE_URL", "http://localhost:8081")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	serverPort := getEnv("SERVER_PORT", "8082")
	cacheTTL := getEnvInt("CACHE_TTL", 300)

	log.Println("Starting CMS Frontend Service...")
	log.Printf("CMS Service URL: %s", cmsServiceURL)
	log.Printf("Stats Service URL: %s", statsServiceURL)
	log.Printf("Server Port: %s", serverPort)

	// Initialize Redis for caching
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		rdb = nil
	} else {
		log.Println("✓ Connected to Redis")
	}

	// Initialize clients
	cmsClient := client.NewCMSClient(cmsServiceURL)
	statsClient := client.NewStatsClient(statsServiceURL)

	// Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Public article routes
	mux.HandleFunc("/api/v1/articles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 20
		}

		filters := make(map[string]string)
		if category := r.URL.Query().Get("category"); category != "" {
			filters["category"] = category
		}
		if tag := r.URL.Query().Get("tag"); tag != "" {
			filters["tag"] = tag
		}

		articles, total, err := cmsClient.ListArticles(r.Context(), page, limit, filters)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"articles": articles,
			"total":    total,
			"page":     page,
			"limit":    limit,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/api/v1/articles/", func(w http.ResponseWriter, r *http.Request) {
		// Extract article ID from path
		path := r.URL.Path
		articleID := path[len("/api/v1/articles/"):]
		if articleID == "" {
			http.Error(w, "Article ID required", http.StatusBadRequest)
			return
		}

		// Check for specific endpoints
		if len(articleID) > 24 && articleID[24:] == "/comments" {
			// Get comments
			actualID := articleID[:24]
			if r.Method != http.MethodGet && r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			if r.Method == http.MethodGet {
				page, _ := strconv.Atoi(r.URL.Query().Get("page"))
				if page < 1 {
					page = 1
				}
				limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
				if limit < 1 || limit > 100 {
					limit = 20
				}
				sortBy := r.URL.Query().Get("sortBy")
				if sortBy == "" {
					sortBy = "likes"
				}

				comments, total, err := statsClient.GetComments(r.Context(), actualID, page, limit, sortBy)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				response := map[string]interface{}{
					"comments": comments,
					"total":    total,
					"page":     page,
					"limit":    limit,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			} else {
				// POST comment
				var req struct {
					Content  string  `json:"content"`
					ParentID *string `json:"parentId"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				// Get auth token from header (simplified)
				authToken := r.Header.Get("Authorization")
				if authToken == "" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Extract user info from token (simplified - should validate JWT)
				userID := "user123"    // TODO: Extract from JWT
				userName := "User 123" // TODO: Extract from JWT

				if err := statsClient.CreateComment(r.Context(), actualID, userID, userName, req.Content, req.ParentID, authToken); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"message":"Comment created"}`))
			}
			return
		}

		// Get single article
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Try cache first if Redis is available
		var article map[string]interface{}
		var err error
		cacheKey := "article:" + articleID

		if rdb != nil {
			cached, err := rdb.Get(r.Context(), cacheKey).Result()
			if err == nil {
				json.Unmarshal([]byte(cached), &article)
			}
		}

		// If not in cache, fetch from CMS service
		if article == nil {
			article, err = cmsClient.GetArticle(r.Context(), articleID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Cache the result
			if rdb != nil {
				data, _ := json.Marshal(article)
				rdb.Set(r.Context(), cacheKey, data, time.Duration(cacheTTL)*time.Second)
			}

			// Record view asynchronously
			go cmsClient.RecordView(context.Background(), articleID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(article)
	})

	// RSS feed
	mux.HandleFunc("/api/v1/rss", func(w http.ResponseWriter, r *http.Request) {
		// Proxy to CMS service
		resp, err := http.Get(cmsServiceURL + "/api/v1/rss?" + r.URL.RawQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	// Start HTTP server
	server := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("CMS Frontend Service starting on port %s", serverPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
