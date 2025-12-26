package worker

import (
	"context"
	"log"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/service"
)

// Scheduler handles scheduled tasks for articles
type Scheduler struct {
	articleService *service.ArticleService
	interval       time.Duration
	stopChan       chan bool
}

// NewScheduler creates a new scheduler
func NewScheduler(articleService *service.ArticleService, interval time.Duration) *Scheduler {
	return &Scheduler{
		articleService: articleService,
		interval:       interval,
		stopChan:       make(chan bool),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Println("Scheduler started")

	for {
		select {
		case <-ticker.C:
			s.processPublishSchedule(ctx)
			s.processExpireSchedule(ctx)
		case <-s.stopChan:
			log.Println("Scheduler stopped")
			return
		case <-ctx.Done():
			log.Println("Scheduler stopped due to context cancellation")
			return
		}
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
}

// processPublishSchedule processes articles scheduled for publication
func (s *Scheduler) processPublishSchedule(ctx context.Context) {
	articles, err := s.articleService.GetPublishableArticles(ctx)
	if err != nil {
		log.Printf("Error getting publishable articles: %v", err)
		return
	}

	for _, article := range articles {
		log.Printf("Auto-publishing article: %s (ID: %s)", article.Title, article.ID.Hex())
		
		err := s.articleService.UpdateStatus(ctx, article.ID, model.ArticleStatusPublished, "scheduler", model.RoleModerator)
		if err != nil {
			log.Printf("Error publishing article %s: %v", article.ID.Hex(), err)
			continue
		}
		
		log.Printf("Successfully published article: %s", article.Title)
	}

	if len(articles) > 0 {
		log.Printf("Processed %d articles for publication", len(articles))
	}
}

// processExpireSchedule processes articles scheduled for expiration
func (s *Scheduler) processExpireSchedule(ctx context.Context) {
	articles, err := s.articleService.GetExpirableArticles(ctx)
	if err != nil {
		log.Printf("Error getting expirable articles: %v", err)
		return
	}

	for _, article := range articles {
		log.Printf("Auto-expiring article: %s (ID: %s)", article.Title, article.ID.Hex())
		
		err := s.articleService.UpdateStatus(ctx, article.ID, model.ArticleStatusArchived, "scheduler", model.RoleModerator)
		if err != nil {
			log.Printf("Error expiring article %s: %v", article.ID.Hex(), err)
			continue
		}
		
		log.Printf("Successfully expired article: %s", article.Title)
	}

	if len(articles) > 0 {
		log.Printf("Processed %d articles for expiration", len(articles))
	}
}
