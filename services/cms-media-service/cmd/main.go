package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-media-service/internal/handler"
	"github.com/vhvplatform/go-cms-service/services/cms-media-service/internal/repository"
	"github.com/vhvplatform/go-cms-service/services/cms-media-service/internal/service"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load configuration
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGODB_DATABASE", "cms_media")
	serverPort := getEnv("SERVER_PORT", "8083")
	uploadDir := getEnv("UPLOAD_DIR", "./uploads")
	baseURL := getEnv("BASE_URL", "http://localhost:"+serverPort)

	log.Println("Starting CMS Media Service...")
	log.Printf("MongoDB URI: %s", mongoURI)
	log.Printf("Database: %s", dbName)
	log.Printf("Server Port: %s", serverPort)
	log.Printf("Upload Directory: %s", uploadDir)

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

	// Create upload directory if not exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Initialize repositories
	mediaRepo := repository.NewMediaRepository(db)

	// Initialize services
	mediaService := service.NewMediaService(mediaRepo, uploadDir, baseURL)

	// Initialize handlers
	mediaHandler := handler.NewMediaHandler(mediaService)

	// Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Media routes
	mux.HandleFunc("/api/v1/media/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			mediaHandler.UploadFile(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/media/files", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			mediaHandler.ListFiles(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/media/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			mediaHandler.GetFile(w, r)
		case http.MethodDelete:
			mediaHandler.DeleteFile(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/media/storage/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			mediaHandler.GetStorageUsage(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/media/folders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			mediaHandler.ListFolders(w, r)
		case http.MethodPost:
			mediaHandler.CreateFolder(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Serve uploaded files
	fs := http.FileServer(http.Dir(uploadDir))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

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
		log.Printf("CMS Media Service starting on port %s", serverPort)
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
