package migrations

import (
	"context"
	"log"

	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/model"
	"github.com/vhvplatform/go-cms-service/services/cms-admin-service/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

// Migration interface
type Migration interface {
	Up(ctx context.Context, db *mongo.Database) error
	Down(ctx context.Context, db *mongo.Database) error
}

// InitialMigration creates initial collections and indexes
type InitialMigration struct{}

// Up applies the migration
func (m *InitialMigration) Up(ctx context.Context, db *mongo.Database) error {
	log.Println("Running initial migration...")

	// Create indexes for articles
	articleRepo := repository.NewArticleRepository(db)
	if err := articleRepo.CreateIndexes(ctx); err != nil {
		return err
	}
	log.Println("✓ Created article indexes")

	// Create indexes for categories
	categoryRepo := repository.NewCategoryRepository(db)
	if err := categoryRepo.CreateIndexes(ctx); err != nil {
		return err
	}
	log.Println("✓ Created category indexes")

	// Create indexes for event streams
	eventStreamRepo := repository.NewEventStreamRepository(db)
	if err := eventStreamRepo.CreateIndexes(ctx); err != nil {
		return err
	}
	log.Println("✓ Created event stream indexes")

	// Create indexes for permissions
	permissionRepo := repository.NewPermissionRepository(db)
	if err := permissionRepo.CreateIndexes(ctx); err != nil {
		return err
	}
	log.Println("✓ Created permission indexes")

	// Create indexes for view stats
	viewStatsRepo := repository.NewViewStatsRepository(db)
	if err := viewStatsRepo.CreateIndexes(ctx); err != nil {
		return err
	}
	log.Println("✓ Created view stats indexes")

	// Seed sample categories
	if err := m.seedCategories(ctx, categoryRepo); err != nil {
		return err
	}
	log.Println("✓ Seeded sample categories")

	log.Println("Initial migration completed successfully")
	return nil
}

// Down reverts the migration
func (m *InitialMigration) Down(ctx context.Context, db *mongo.Database) error {
	log.Println("Reverting initial migration...")

	// Drop collections
	collections := []string{"articles", "categories", "event_lines", "permissions", "article_views"}
	for _, coll := range collections {
		if err := db.Collection(coll).Drop(ctx); err != nil {
			log.Printf("Warning: Failed to drop collection %s: %v", coll, err)
		}
	}

	log.Println("Initial migration reverted")
	return nil
}

// seedCategories creates sample categories
func (m *InitialMigration) seedCategories(ctx context.Context, repo *repository.CategoryRepository) error {
	categories := []*model.Category{
		{
			Name:         "News",
			Slug:         "news",
			Description:  "Latest news and updates",
			CategoryType: model.CategoryTypeArticle,
			ArticleType:  model.ArticleTypeNews,
			Ordering:     1,
			CreatedBy:    "system",
		},
		{
			Name:         "Videos",
			Slug:         "videos",
			Description:  "Video content",
			CategoryType: model.CategoryTypeArticle,
			ArticleType:  model.ArticleTypeVideo,
			Ordering:     2,
			CreatedBy:    "system",
		},
		{
			Name:         "Photo Galleries",
			Slug:         "galleries",
			Description:  "Photo galleries and albums",
			CategoryType: model.CategoryTypeArticle,
			ArticleType:  model.ArticleTypePhotoGallery,
			Ordering:     3,
			CreatedBy:    "system",
		},
		{
			Name:         "Legal Documents",
			Slug:         "legal",
			Description:  "Legal documents and regulations",
			CategoryType: model.CategoryTypeArticle,
			ArticleType:  model.ArticleTypeLegalDocument,
			Ordering:     4,
			CreatedBy:    "system",
		},
		{
			Name:         "External Link",
			Slug:         "external",
			Description:  "Link to external site",
			CategoryType: model.CategoryTypeLink,
			CategoryLink: "https://example.com",
			Ordering:     5,
			CreatedBy:    "system",
		},
	}

	for _, category := range categories {
		// Check if category already exists
		existing, _ := repo.FindBySlug(ctx, category.Slug)
		if existing != nil {
			log.Printf("Category '%s' already exists, skipping", category.Name)
			continue
		}

		if err := repo.Create(ctx, category); err != nil {
			return err
		}
		log.Printf("Created category: %s", category.Name)
	}

	return nil
}

// RunMigrations runs all migrations
func RunMigrations(ctx context.Context, db *mongo.Database) error {
	migrations := []Migration{
		&InitialMigration{},
	}

	for i, migration := range migrations {
		log.Printf("Running migration %d...", i+1)
		if err := migration.Up(ctx, db); err != nil {
			return err
		}
	}

	return nil
}
