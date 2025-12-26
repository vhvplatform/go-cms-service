package main

import (
	"context"
	"net/http"
	"time"

	"github.com/vhvplatform/go-cms-service/pkg/config"
	"github.com/vhvplatform/go-cms-service/pkg/database"
	"github.com/vhvplatform/go-cms-service/pkg/httpserver"
	"github.com/vhvplatform/go-cms-service/pkg/logger"
	pkgMiddleware "github.com/vhvplatform/go-cms-service/pkg/middleware"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/handler"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/middleware"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/migrations"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/service"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/util"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/worker"
)

func main() {
	// Load configuration
	cfg := config.NewConfig("cms-admin-service")
	mongoCfg := config.NewMongoConfig()
	
	baseURL := config.GetEnv("BASE_URL", "http://localhost:"+cfg.ServerPort)
	uploadDir := config.GetEnv("UPLOAD_DIR", "./uploads")
	runMigrations := config.GetEnvBool("RUN_MIGRATIONS", true)

	// Initialize logger
	log := logger.New(cfg.ServiceName, cfg.LogLevel)

	log.Info("Starting CMS Admin Service...")
	log.Info("MongoDB URI: %s", mongoCfg.URI)
	log.Info("Database: %s", mongoCfg.Database)
	log.Info("Server Port: %s", cfg.ServerPort)
	log.Info("Upload Directory: %s", uploadDir)

	// Connect to MongoDB
	ctx := context.Background()
	mongoClient, err := database.ConnectMongo(ctx, mongoCfg.URI, mongoCfg.Database, mongoCfg.Timeout)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Close(ctx)

	log.Info("✓ Connected to MongoDB")

	db := mongoClient.Database

	// Run migrations
	if runMigrations {
		if err := migrations.RunMigrations(ctx, db); err != nil {
			log.Fatal("Failed to run migrations: %v", err)
		}
	}

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	_ = repository.NewEventStreamRepository(db) // Initialize for migrations
	permissionRepo := repository.NewPermissionRepository(db)
	viewStatsRepo := repository.NewViewStatsRepository(db)
	actionLogRepo := repository.NewActionLogRepository(db)
	versionRepo := repository.NewArticleVersionRepository(db)
	rejectionNoteRepo := repository.NewRejectionNoteRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	// Initialize utilities
	imageDownloader := util.NewImageDownloader(uploadDir, baseURL)

	// Initialize view queue
	viewQueue := worker.NewViewQueue(articleRepo, viewStatsRepo, 10000, 100, 5*time.Second)
	viewQueue.Start(ctx)
	defer viewQueue.Stop()

	// Initialize services
	articleService := service.NewArticleService(articleRepo, permissionRepo, viewStatsRepo, viewQueue, actionLogRepo, versionRepo, rejectionNoteRepo, imageDownloader)
	categoryService := service.NewCategoryService(categoryRepo)
	commentService := service.NewCommentService(commentRepo)
	rssService := service.NewRSSService(articleRepo, baseURL)

	// Initialize handlers
	articleHandler := handler.NewArticleHandler(articleService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	commentHandler := handler.NewCommentHandler(commentService)
	rssHandler := handler.NewRSSHandler(rssService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware()

	// Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Article routes
	mux.Handle("/api/v1/articles", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			articleHandler.CreateArticle(w, r)
		case http.MethodGet:
			articleHandler.ListArticles(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/articles/", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route based on path and method
		if r.URL.Path == "/api/v1/articles/reorder" && r.Method == http.MethodPost {
			articleHandler.ReorderArticles(w, r)
			return
		}

		// Extract article ID and handle article-specific routes
		switch r.Method {
		case http.MethodGet:
			articleHandler.GetArticle(w, r)
		case http.MethodPatch:
			articleHandler.UpdateArticle(w, r)
		case http.MethodDelete:
			articleHandler.DeleteArticle(w, r)
		case http.MethodPost:
			// Check for /publish or /view endpoint
			if containsSegment(r.URL.Path, "publish") {
				articleHandler.PublishArticle(w, r)
			} else if containsSegment(r.URL.Path, "view") {
				articleHandler.ViewArticle(w, r)
			} else {
				http.Error(w, "Not found", http.StatusNotFound)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Category routes
	mux.Handle("/api/v1/categories", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			categoryHandler.CreateCategory(w, r)
		case http.MethodGet:
			categoryHandler.GetCategoryTree(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/categories/tree", authMiddleware.Authenticate(http.HandlerFunc(categoryHandler.GetCategoryTree)))

	mux.Handle("/api/v1/categories/", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			categoryHandler.GetCategory(w, r)
		case http.MethodPatch:
			categoryHandler.UpdateCategory(w, r)
		case http.MethodDelete:
			categoryHandler.DeleteCategory(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Search route
	mux.Handle("/api/v1/search", authMiddleware.Authenticate(http.HandlerFunc(articleHandler.SearchArticles)))

	// Statistics route
	mux.Handle("/api/v1/statistics/articles/", authMiddleware.Authenticate(http.HandlerFunc(articleHandler.GetArticleStats)))

	// RSS route (public)
	mux.HandleFunc("/api/v1/rss", rssHandler.GetRSSFeed)

	// Comment routes
	mux.Handle("/api/v1/comments/pending", authMiddleware.Authenticate(http.HandlerFunc(commentHandler.GetPendingComments)))
	
	// User favorites route
	mux.Handle("/api/v1/users/favorites", authMiddleware.Authenticate(http.HandlerFunc(commentHandler.GetUserFavorites)))

	// Comment-specific routes with path-based routing
	mux.Handle("/api/v1/comments/", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})))

	// Article-specific comment and favorite routes
	mux.Handle("/api/v1/articles/", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route based on path and method
		if r.URL.Path == "/api/v1/articles/reorder" && r.Method == http.MethodPost {
			articleHandler.ReorderArticles(w, r)
			return
		}

		// Handle /comments endpoint
		if containsSegment(r.URL.Path, "comments") {
			if r.Method == http.MethodPost {
				commentHandler.CreateComment(w, r)
			} else if r.Method == http.MethodGet {
				commentHandler.GetArticleComments(w, r)
			}
			return
		}

		// Handle /favorite endpoint
		if containsSegment(r.URL.Path, "favorite") {
			if r.Method == http.MethodPost {
				commentHandler.AddFavorite(w, r)
			} else if r.Method == http.MethodDelete {
				commentHandler.RemoveFavorite(w, r)
			}
			return
		}

		// Extract article ID and handle article-specific routes
		switch r.Method {
		case http.MethodGet:
			articleHandler.GetArticle(w, r)
		case http.MethodPatch:
			articleHandler.UpdateArticle(w, r)
		case http.MethodDelete:
			articleHandler.DeleteArticle(w, r)
		case http.MethodPost:
			// Check for /publish or /view endpoint
			if containsSegment(r.URL.Path, "publish") {
				articleHandler.PublishArticle(w, r)
			} else if containsSegment(r.URL.Path, "view") {
				articleHandler.ViewArticle(w, r)
			} else if containsSegment(r.URL.Path, "reject") {
				articleHandler.RejectArticle(w, r)
			} else if containsSegment(r.URL.Path, "rejection-notes") {
				if containsSegment(r.URL.Path, "resolve") {
					articleHandler.ResolveRejectionNotes(w, r)
				} else {
					articleHandler.AddRejectionNote(w, r)
				}
			} else if containsSegment(r.URL.Path, "versions") {
				if containsSegment(r.URL.Path, "restore") {
					articleHandler.RestoreArticleVersion(w, r)
				} else {
					articleHandler.GetArticleVersions(w, r)
				}
			} else if containsSegment(r.URL.Path, "logs") {
				articleHandler.GetActionLogs(w, r)
			} else if containsSegment(r.URL.Path, "share") {
				articleHandler.GetSocialShareURLs(w, r)
			} else {
				http.Error(w, "Not found", http.StatusNotFound)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Start scheduler
	scheduler := worker.NewScheduler(articleService, 1*time.Minute)
	go scheduler.Start(ctx)
	defer scheduler.Stop()

	// Wrap mux with common middleware
	handler := pkgMiddleware.Chain(
		pkgMiddleware.LoggingMiddleware(log),
		pkgMiddleware.RecoveryMiddleware(log),
	)(mux)

	// Start HTTP server with graceful shutdown
	serverCfg := httpserver.DefaultConfig(cfg.ServerPort, handler)
	server := httpserver.NewServer(serverCfg, log)

	if err := server.Start(); err != nil {
		log.Fatal("Server error: %v", err)
	}
}

func containsSegment(path, segment string) bool {
	return len(path) > 0 && (path[len(path)-len(segment):] == segment || 
		len(path) > len(segment) && path[len(path)-len(segment)-1:] == "/"+segment)
}
