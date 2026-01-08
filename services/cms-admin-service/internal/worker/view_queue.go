package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ViewEvent represents a view event to be processed
type ViewEvent struct {
	ArticleID primitive.ObjectID
	Timestamp time.Time
}

// ViewQueue handles asynchronous view counting
type ViewQueue struct {
	queue         chan ViewEvent
	articleRepo   *repository.ArticleRepository
	viewStatsRepo *repository.ViewStatsRepository
	batchSize     int
	flushInterval time.Duration
	stopChan      chan bool
	wg            sync.WaitGroup
}

// NewViewQueue creates a new view queue
func NewViewQueue(
	articleRepo *repository.ArticleRepository,
	viewStatsRepo *repository.ViewStatsRepository,
	queueSize int,
	batchSize int,
	flushInterval time.Duration,
) *ViewQueue {
	return &ViewQueue{
		queue:         make(chan ViewEvent, queueSize),
		articleRepo:   articleRepo,
		viewStatsRepo: viewStatsRepo,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		stopChan:      make(chan bool),
	}
}

// Start starts the queue processor
func (q *ViewQueue) Start(ctx context.Context) {
	q.wg.Add(1)
	go q.processQueue(ctx)
	log.Println("View queue processor started")
}

// Stop stops the queue processor
func (q *ViewQueue) Stop() {
	close(q.stopChan)
	q.wg.Wait()
	log.Println("View queue processor stopped")
}

// Enqueue adds a view event to the queue
func (q *ViewQueue) Enqueue(articleID primitive.ObjectID) error {
	select {
	case q.queue <- ViewEvent{
		ArticleID: articleID,
		Timestamp: time.Now(),
	}:
		return nil
	default:
		log.Printf("Warning: View queue is full, dropping view event for article %s", articleID.Hex())
		return nil // Don't block, just log and continue
	}
}

// processQueue processes view events from the queue
func (q *ViewQueue) processQueue(ctx context.Context) {
	defer q.wg.Done()

	batch := make([]ViewEvent, 0, q.batchSize)
	ticker := time.NewTicker(q.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case event := <-q.queue:
			batch = append(batch, event)

			// Process batch if it reaches the batch size
			if len(batch) >= q.batchSize {
				q.processBatch(ctx, batch)
				batch = batch[:0] // Clear batch
			}

		case <-ticker.C:
			// Periodically flush the batch even if not full
			if len(batch) > 0 {
				q.processBatch(ctx, batch)
				batch = batch[:0] // Clear batch
			}

		case <-q.stopChan:
			// Process remaining events before stopping
			if len(batch) > 0 {
				q.processBatch(ctx, batch)
			}
			// Drain remaining queue
			for len(q.queue) > 0 {
				event := <-q.queue
				batch = append(batch, event)
				if len(batch) >= q.batchSize {
					q.processBatch(ctx, batch)
					batch = batch[:0]
				}
			}
			if len(batch) > 0 {
				q.processBatch(ctx, batch)
			}
			return

		case <-ctx.Done():
			log.Println("View queue processor stopped due to context cancellation")
			return
		}
	}
}

// processBatch processes a batch of view events
func (q *ViewQueue) processBatch(ctx context.Context, batch []ViewEvent) {
	if len(batch) == 0 {
		return
	}

	log.Printf("Processing batch of %d view events", len(batch))

	// Aggregate views by article and date
	viewCounts := make(map[string]map[string]int) // articleID -> date -> count

	for _, event := range batch {
		articleID := event.ArticleID.Hex()
		date := time.Date(event.Timestamp.Year(), event.Timestamp.Month(), event.Timestamp.Day(), 0, 0, 0, 0, event.Timestamp.Location())
		dateKey := date.Format("2006-01-02")

		if _, exists := viewCounts[articleID]; !exists {
			viewCounts[articleID] = make(map[string]int)
		}
		viewCounts[articleID][dateKey]++
	}

	// Update database
	for articleIDHex, dates := range viewCounts {
		articleID, err := primitive.ObjectIDFromHex(articleIDHex)
		if err != nil {
			log.Printf("Error parsing article ID %s: %v", articleIDHex, err)
			continue
		}

		// Increment article view count
		totalViews := 0
		for _, count := range dates {
			totalViews += count
		}

		// Update article's total view count (non-blocking)
		if err := q.articleRepo.IncrementViewCount(ctx, articleID); err != nil {
			log.Printf("Error incrementing view count for article %s: %v", articleIDHex, err)
		}

		// Update daily stats
		for dateStr, count := range dates {
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				log.Printf("Error parsing date %s: %v", dateStr, err)
				continue
			}

			// Record views for the specific date
			for i := 0; i < count; i++ {
				if err := q.viewStatsRepo.RecordView(ctx, articleID, date); err != nil {
					log.Printf("Error recording view stats for article %s on %s: %v", articleIDHex, dateStr, err)
					break // Avoid logging same error multiple times
				}
			}
		}
	}

	log.Printf("Successfully processed %d view events", len(batch))
}

// GetQueueSize returns the current queue size
func (q *ViewQueue) GetQueueSize() int {
	return len(q.queue)
}

// GetQueueCapacity returns the queue capacity
func (q *ViewQueue) GetQueueCapacity() int {
	return cap(q.queue)
}
