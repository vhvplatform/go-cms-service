package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/vhvplatform/go-cms-service/services/cms-crawler-service/internal/repository"
	"github.com/vhvplatform/go-cms-service/services/cms-crawler-service/internal/service"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Get configuration from environment
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGODB_DATABASE", "cms_crawler")
	serverPort := getEnv("SERVER_PORT", "8084")
	
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(context.Background())
	
	// Ping database
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}
	log.Println("Connected to MongoDB")
	
	db := client.Database(dbName)
	
	// Initialize repositories
	articleRepo := repository.NewCrawlerArticleRepository(db)
	sourceRepo := repository.NewCrawlerSourceRepository(db)
	campaignRepo := repository.NewCrawlerCampaignRepository(db)
	
	// Initialize service
	crawlerService := service.NewCrawlerService(articleRepo, sourceRepo, campaignRepo)
	
	// Setup HTTP server
	router := gin.Default()
	
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	
	// API routes
	api := router.Group("/api/v1")
	{
		// Campaign management
		api.POST("/campaigns", func(c *gin.Context) {
			// Create campaign handler
			c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
		})
		
		api.POST("/campaigns/:id/run", func(c *gin.Context) {
			// Run campaign handler
			c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
		})
		
		// Source management
		api.POST("/sources", func(c *gin.Context) {
			// Create source handler
			c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
		})
		
		// Article management
		api.GET("/articles", func(c *gin.Context) {
			// List articles handler
			c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
		})
		
		api.POST("/articles/:id/approve", func(c *gin.Context) {
			// Approve article handler
			c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
		})
		
		api.POST("/articles/:id/reject", func(c *gin.Context) {
			// Reject article handler
			c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
		})
		
		api.POST("/articles/:id/convert", func(c *gin.Context) {
			// Convert to article handler
			c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
		})
		
		// Statistics
		api.GET("/stats/:tenantId", func(c *gin.Context) {
			tenantID := c.Param("tenantId")
			stats, err := crawlerService.GetStats(c.Request.Context(), tenantID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, stats)
		})
	}
	
	// Setup cron scheduler for campaigns
	cronScheduler := cron.New()
	
	// Add a job to run cleanup daily
	cronScheduler.AddFunc("@daily", func() {
		log.Println("Running daily cleanup job")
		// Get all tenants and run cleanup
		// This is simplified - would need tenant service integration
	})
	
	cronScheduler.Start()
	defer cronScheduler.Stop()
	
	// Start HTTP server
	srv := &http.Server{
		Addr:    ":" + serverPort,
		Handler: router,
	}
	
	go func() {
		log.Printf("Starting CMS Crawler Service on port %s", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	
	log.Println("Server exited")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
