package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-crawler-service/internal/crawler"
	"github.com/vhvplatform/go-cms-service/services/cms-crawler-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-crawler-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CrawlerService struct {
	articleRepo     *repository.CrawlerArticleRepository
	sourceRepo      *repository.CrawlerSourceRepository
	campaignRepo    *repository.CrawlerCampaignRepository
	extractor       *crawler.ContentExtractor
	similarityCalc  *crawler.SimilarityCalculator
}

func NewCrawlerService(
	articleRepo *repository.CrawlerArticleRepository,
	sourceRepo *repository.CrawlerSourceRepository,
	campaignRepo *repository.CrawlerCampaignRepository,
) *CrawlerService {
	return &CrawlerService{
		articleRepo:    articleRepo,
		sourceRepo:     sourceRepo,
		campaignRepo:   campaignRepo,
		extractor:      crawler.NewContentExtractor(),
		similarityCalc: crawler.NewSimilarityCalculator(),
	}
}

// RunCampaign executes a crawler campaign
func (s *CrawlerService) RunCampaign(ctx context.Context, campaignID primitive.ObjectID) error {
	campaign, err := s.campaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		return err
	}
	
	if !campaign.IsActive {
		return fmt.Errorf("campaign is not active")
	}
	
	// Get all sources for this campaign
	for _, sourceID := range campaign.SourceIDs {
		source, err := s.sourceRepo.GetByID(ctx, sourceID)
		if err != nil {
			log.Printf("Failed to get source %s: %v", sourceID.Hex(), err)
			continue
		}
		
		if !source.IsActive {
			continue
		}
		
		// Apply delay between requests
		if source.DelayMs > 0 {
			time.Sleep(time.Duration(source.DelayMs) * time.Millisecond)
		}
		
		// Crawl the source
		if err := s.CrawlSource(ctx, source, campaign); err != nil {
			log.Printf("Failed to crawl source %s: %v", source.Name, err)
			if updateErr := s.sourceRepo.UpdateLastCrawled(ctx, source.ID, false); updateErr != nil {
				log.Printf("Failed to update source status: %v", updateErr)
			}
		} else {
			if updateErr := s.sourceRepo.UpdateLastCrawled(ctx, source.ID, true); updateErr != nil {
				log.Printf("Failed to update source status: %v", updateErr)
			}
		}
	}
	
	// Update campaign last run
	s.campaignRepo.UpdateLastRun(ctx, campaignID)
	
	return nil
}

// CrawlSource crawls a single source
func (s *CrawlerService) CrawlSource(ctx context.Context, source *model.CrawlerSource, campaign *model.CrawlerCampaign) error {
	var articles []*model.CrawlerArticle
	var err error
	
	switch source.Type {
	case "html":
		article, err := s.extractor.Extract(ctx, source.URL, source.ExtractionConfig, source)
		if err != nil {
			return err
		}
		articles = []*model.CrawlerArticle{article}
		
	case "rss":
		articles, err = s.extractor.ExtractFromRSS(ctx, source.URL, source)
		if err != nil {
			return err
		}
		
	default:
		return fmt.Errorf("unsupported source type: %s", source.Type)
	}
	
	// Process each extracted article
	for _, article := range articles {
		article.CampaignID = campaign.ID
		article.TenantID = campaign.TenantID
		
		// Check for duplicates
		duplicates, err := s.articleRepo.FindDuplicates(ctx, article.ContentHash)
		if err != nil {
			log.Printf("Error checking duplicates for article: %v", err)
			// Continue with save - better to have potential duplicate than lose content
		} else if len(duplicates) > 0 {
			log.Printf("Duplicate article found: %s", article.Title)
			continue
		}
		
		// Set auto-approve status
		if source.AutoApprove || campaign.AutoApprove {
			article.Status = "approved"
			article.ApprovedAt = time.Now()
		}
		
		// Save article
		if err := s.articleRepo.Create(ctx, article); err != nil {
			log.Printf("Failed to save article: %v", err)
			continue
		}
		
		// If auto-approved, find similar articles for grouping
		if article.Status == "approved" {
			s.findAndGroupSimilar(ctx, article)
		}
	}
	
	return nil
}

// findAndGroupSimilar finds similar articles and groups them
func (s *CrawlerService) findAndGroupSimilar(ctx context.Context, article *model.CrawlerArticle) {
	// Get recent articles from the same tenant
	recentArticles, err := s.articleRepo.GetByTenant(ctx, article.TenantID, "approved", 100, 0)
	if err != nil {
		return
	}
	
	similarThreshold := 0.7 // 70% similarity
	var similarArticles []*model.CrawlerArticle
	
	for _, other := range recentArticles {
		if other.ID == article.ID {
			continue
		}
		
		similarity := s.similarityCalc.CalculateSimilarity(article.Content, other.Content)
		if similarity >= similarThreshold {
			similarArticles = append(similarArticles, other)
		}
	}
	
	// If we found similar articles, create or update group
	if len(similarArticles) > 0 {
		log.Printf("Found %d similar articles for: %s", len(similarArticles), article.Title)
		// Group creation logic would go here
	}
}

// ApproveArticle approves a crawled article
func (s *CrawlerService) ApproveArticle(ctx context.Context, articleID primitive.ObjectID, userID string) error {
	return s.articleRepo.UpdateStatus(ctx, articleID, "approved", userID)
}

// RejectArticle rejects a crawled article
func (s *CrawlerService) RejectArticle(ctx context.Context, articleID primitive.ObjectID, reason string) error {
	return s.articleRepo.UpdateStatus(ctx, articleID, "rejected", "")
}

// CleanupOldArticles removes old crawled articles based on retention policy
func (s *CrawlerService) CleanupOldArticles(ctx context.Context, tenantID string) error {
	campaigns, err := s.campaignRepo.GetActiveCampaigns(ctx, tenantID)
	if err != nil {
		return err
	}
	
	for _, campaign := range campaigns {
		if campaign.RetentionDays > 0 {
			beforeDate := time.Now().AddDate(0, 0, -campaign.RetentionDays)
			deleted, err := s.articleRepo.DeleteOldArticles(ctx, tenantID, beforeDate)
			if err != nil {
				log.Printf("Failed to cleanup old articles: %v", err)
			} else {
				log.Printf("Deleted %d old articles for tenant %s", deleted, tenantID)
			}
		}
	}
	
	return nil
}

// GetStats retrieves crawler statistics
func (s *CrawlerService) GetStats(ctx context.Context, tenantID string) (*model.CrawlerStats, error) {
	return s.articleRepo.GetStats(ctx, tenantID)
}

// ConvertToArticle converts a crawled article to a real article
func (s *CrawlerService) ConvertToArticle(ctx context.Context, crawlerArticleID primitive.ObjectID, userID string) (primitive.ObjectID, error) {
	article, err := s.articleRepo.GetByID(ctx, crawlerArticleID)
	if err != nil {
		return primitive.NilObjectID, err
	}
	
	if article.Status != "approved" {
		return primitive.NilObjectID, fmt.Errorf("article must be approved before conversion")
	}
	
	// Here we would call the main CMS service to create the article
	// For now, just update the status
	if err := s.articleRepo.UpdateStatus(ctx, crawlerArticleID, "converted", userID); err != nil {
		return primitive.NilObjectID, err
	}
	
	// Return a placeholder article ID
	// In production, this would be the actual created article ID
	return primitive.NewObjectID(), nil
}
