package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-stats-service/internal/handler"
	"github.com/vhvplatform/go-cms-service/services/cms-stats-service/internal/repository"
	"github.com/vhvplatform/go-cms-service/services/cms-stats-service/internal/service"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load configuration
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGODB_DATABASE", "cms_comments")
	serverPort := getEnv("SERVER_PORT", "8081")

	log.Println("Starting Comment & Statistics Service...")
	log.Printf("MongoDB URI: %s", mongoURI)
	log.Printf("Database: %s", dbName)
	log.Printf("Server Port: %s", serverPort)

	// Connect to MongoDB
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Ping MongoDB
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("✓ Connected to MongoDB")

	db := client.Database(dbName)

	// Initialize repositories
	commentRepo := repository.NewCommentRepository(db)

	// Initialize services
	commentService := service.NewCommentService(commentRepo)

	// Initialize handlers
	commentHandler := handler.NewCommentHandler(commentService)

	// Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Comment routes
	mux.HandleFunc("/api/v1/articles/", func(w http.ResponseWriter, r *http.Request) {
		if containsSegment(r.URL.Path, "comments") {
			if r.Method == http.MethodPost {
				commentHandler.CreateComment(w, r)
			} else if r.Method == http.MethodGet {
				commentHandler.GetArticleComments(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if containsSegment(r.URL.Path, "favorite") {
			if r.Method == http.MethodPost {
				commentHandler.AddFavorite(w, r)
			} else if r.Method == http.MethodDelete {
				commentHandler.RemoveFavorite(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		http.Error(w, "Not found", http.StatusNotFound)
	})

	// Comment-specific routes
	mux.HandleFunc("/api/v1/comments/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/comments/pending" && r.Method == http.MethodGet {
			commentHandler.GetPendingComments(w, r)
			return
		}

		if containsSegment(r.URL.Path, "replies") && r.Method == http.MethodGet {
			commentHandler.GetCommentReplies(w, r)
		} else if containsSegment(r.URL.Path, "moderate") && r.Method == http.MethodPost {
			commentHandler.ModerateComment(w, r)
		} else if containsSegment(r.URL.Path, "like") {
			if r.Method == http.MethodPost {
				commentHandler.LikeComment(w, r)
			} else if r.Method == http.MethodDelete {
				commentHandler.UnlikeComment(w, r)
			}
		} else if containsSegment(r.URL.Path, "report") && r.Method == http.MethodPost {
			commentHandler.ReportComment(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	// User favorites route
	mux.HandleFunc("/api/v1/users/favorites", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			commentHandler.GetUserFavorites(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
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
		log.Printf("Comment & Statistics Service starting on port %s", serverPort)
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

func containsSegment(path, segment string) bool {
	return len(path) > 0 && (path[len(path)-len(segment):] == segment ||
		len(path) > len(segment) && path[len(path)-len(segment)-1:] == "/"+segment)
}
