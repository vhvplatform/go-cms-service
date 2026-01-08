# Article Management Service - Go CMS

A comprehensive article management microservice built with Go and MongoDB, designed for managing various types of content including news, videos, photo galleries, legal documents, and more.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Development Guide](#development-guide)
- [API Documentation](#api-documentation)
- [Extending Article Types](#extending-article-types)
- [Testing](#testing)
- [Deployment](#deployment)
- [Contributing](#contributing)

## Overview

The Article Management Service provides a robust platform for managing different types of content with features like:
- Multi-type article support (News, Video, Photo Gallery, Legal Documents, etc.)
- Category management with tree structure
- Permission groups for editorial access control
- Event stream tracking
- Full-text search
- View statistics with queue-based processing
- Redis caching for public APIs
- Scheduled publishing and expiration

## Features

### Core Features
- ✅ **Multi-type Articles**: Support for 14+ article types with type-specific fields
- ✅ **CRUD Operations**: Complete Create, Read, Update, Delete operations
- ✅ **Category Management**: Hierarchical category tree structure
- ✅ **Permission System**: Role-based access control (Writer, Editor, Moderator)
- ✅ **Permission Groups**: Manage editorial access by category groups
- ✅ **Workflow Management**: Draft → Pending Review → Published → Archived
- ✅ **Scheduled Publishing**: Auto-publish and expire articles based on dates
- ✅ **Full-text Search**: MongoDB text index based search
- ✅ **View Statistics**: Queue-based view counting with daily aggregation
- ✅ **Caching**: Redis-based caching for public APIs with auto-invalidation
- ✅ **Event Streams**: Organize articles by event timelines

### Article Types Supported
1. **News** - Standard news articles
2. **Video** - Video content with duration and thumbnail
3. **Photo Gallery** - Image collections with captions
4. **Legal Document** - Legal texts with issue dates and PDF attachments
5. **Staff Profile** - Employee/staff information
6. **Job** - Job postings and opportunities
7. **Procedure** - Process documentation
8. **Download** - Downloadable files
9. **Podcast** - Audio content with episodes
10. **Event Info** - Event information with dates and venue
11. **Infographic** - Visual information graphics
12. **Destination** - Travel/location information
13. **Partner** - Partner/sponsor information
14. **PDF** - PDF document articles

## Architecture

### Project Structure

```
services/article-service/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── cache/                  # Caching layer
│   │   └── cache.go           # Redis and memory cache implementations
│   ├── handler/               # HTTP request handlers
│   │   ├── article_handler.go        # Admin article endpoints
│   │   ├── public_article_handler.go # Public article endpoints
│   │   ├── category_handler.go       # Category management
│   │   ├── permission_group_handler.go # Permission group management
│   │   └── helpers.go                # Handler utilities
│   ├── middleware/            # HTTP middleware
│   │   ├── auth_middleware.go        # Authentication
│   │   └── permission_middleware.go  # Authorization
│   ├── migrations/            # Database migrations
│   │   └── initial.go         # Initial schema and seed data
│   ├── model/                 # Data models
│   │   ├── article.go         # Article model with all types
│   │   ├── category.go        # Category model
│   │   ├── event_stream.go    # Event stream model
│   │   ├── permission.go      # Permission model
│   │   └── permission_group.go # Permission group model
│   ├── repository/            # Data access layer
│   │   ├── article_repository.go
│   │   ├── category_repository.go
│   │   ├── event_stream_repository.go
│   │   ├── permission_repository.go
│   │   ├── permission_group_repository.go
│   │   └── view_stats_repository.go
│   ├── service/               # Business logic layer
│   │   ├── article_service.go
│   │   ├── public_article_service.go  # Public APIs with caching
│   │   ├── category_service.go
│   │   └── permission_group_service.go
│   └── worker/                # Background workers
│       ├── scheduler.go       # Publish/expire scheduler
│       └── view_queue.go      # View counting queue
├── configs/
│   └── config.example.env     # Configuration template
├── docs/
│   ├── openapi.yaml          # OpenAPI/Swagger specification
│   └── DEVELOPMENT.md        # This file
├── tests/                     # Test files
├── Dockerfile                 # Docker image definition
└── docker-compose.example.yml # Docker Compose example

```

### Technology Stack
- **Language**: Go 1.21+
- **Database**: MongoDB 5.0+
- **Cache**: Redis 6.0+ (with in-memory fallback)
- **Framework**: Standard Go net/http (ready for @vhvplatform/go-framework integration)

### Design Patterns
- **Repository Pattern**: Separates data access from business logic
- **Service Layer**: Encapsulates business logic
- **Handler/Controller**: HTTP request handling
- **Middleware Pattern**: Cross-cutting concerns (auth, logging)
- **Queue Pattern**: Asynchronous view counting
- **Cache-Aside Pattern**: Caching strategy for reads

## Getting Started

### Prerequisites
- Go 1.21 or higher
- MongoDB 5.0 or higher
- Redis 6.0 or higher (optional, will use in-memory cache if unavailable)
- Docker and Docker Compose (for containerized setup)

### Quick Start with Docker Compose

1. **Clone the repository**
```bash
git clone https://github.com/vhvplatform/go-cms-service.git
cd go-cms-service
```

2. **Copy and configure environment**
```bash
cp services/article-service/configs/config.example.env .env
# Edit .env with your configuration
```

3. **Start services**
```bash
docker-compose up -d
```

4. **Verify installation**
```bash
curl http://localhost:8080/health
```

### Local Development Setup

1. **Install dependencies**
```bash
cd services/article-service
go mod download
```

2. **Start MongoDB**
```bash
docker run -d -p 27017:27017 --name mongodb mongo:5.0
```

3. **Start Redis (optional)**
```bash
docker run -d -p 6379:6379 --name redis redis:6
```

4. **Set environment variables**
```bash
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=cms
export SERVER_PORT=8080
export REDIS_ADDR=localhost:6379
export RUN_MIGRATIONS=true
```

5. **Run the service**
```bash
go run cmd/main.go
```

## Development Guide

### Code Organization

#### 1. Models (internal/model/)
Models define the data structure and are shared across layers. Follow these guidelines:

```go
// Good: Clear field tags, documentation
type Article struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Title       string             `json:"title" bson:"title"`       // Required
    Slug        string             `json:"slug" bson:"slug"`         // Auto-generated
    ArticleType ArticleType        `json:"articleType" bson:"articleType"`
    Status      ArticleStatus      `json:"status" bson:"status"`
    // ... more fields
}

// Use enums for type safety
type ArticleType string

const (
    ArticleTypeNews   ArticleType = "News"
    ArticleTypeVideo  ArticleType = "Video"
    // ...
)
```

#### 2. Repository Layer (internal/repository/)
Repositories handle all database operations. Follow these guidelines:

```go
// Good: Clear separation of concerns
type ArticleRepository struct {
    collection *mongo.Collection
}

// Create creates a new article
func (r *ArticleRepository) Create(ctx context.Context, article *model.Article) error {
    article.ID = primitive.NewObjectID()
    article.CreatedAt = time.Now()
    article.UpdatedAt = time.Now()
    
    _, err := r.collection.InsertOne(ctx, article)
    return err
}

// Always use context for cancellation support
// Return errors, don't panic
// Use meaningful function names
```

#### 3. Service Layer (internal/service/)
Services contain business logic. Follow these guidelines:

```go
// Good: Business logic validation
func (s *ArticleService) Update(ctx context.Context, article *model.Article, userID string, userRole model.Role) error {
    // 1. Validate permissions
    existing, err := s.repo.FindByID(ctx, article.ID)
    if err != nil {
        return err
    }

    // 2. Check business rules
    if existing.Status == model.ArticleStatusPublished && userRole == model.RoleWriter {
        return fmt.Errorf("cannot edit published article: requires editor permission")
    }

    // 3. Perform operation
    return s.repo.Update(ctx, article)
}
```

#### 4. Handler Layer (internal/handler/)
Handlers manage HTTP requests/responses:

```go
// Good: Clean error handling and response structure
func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
    id, err := getIDFromPath(r, "id")
    if err != nil {
        respondError(w, http.StatusBadRequest, "Invalid article ID")
        return
    }

    article, err := h.service.FindByID(r.Context(), id)
    if err != nil {
        respondError(w, http.StatusNotFound, err.Error())
        return
    }

    respondJSON(w, http.StatusOK, article)
}
```

### SonarQube Compliance

#### Code Quality Standards

1. **No Code Duplication**
   - DRY principle: Extract common logic into functions
   - Maximum 5% duplication allowed

2. **Cyclomatic Complexity**
   - Keep functions simple: Max complexity of 15
   - Break down complex functions into smaller ones

3. **Function Length**
   - Max 50-75 lines per function
   - Single Responsibility Principle

4. **Error Handling**
   ```go
   // Good
   if err != nil {
       log.Printf("Failed to create article: %v", err)
       return fmt.Errorf("create article failed: %w", err)
   }

   // Bad
   if err != nil {
       panic(err) // Never panic in production code
   }
   ```

5. **No Magic Numbers**
   ```go
   // Good
   const (
       DefaultPageSize = 20
       MaxPageSize     = 100
   )

   // Bad
   limit := 20 // What does 20 mean?
   ```

6. **Comment Critical Code**
   ```go
   // Good: Explain WHY, not WHAT
   // Check if article has been reviewed to enforce editorial workflow.
   // Once reviewed, only editors can make changes to maintain content integrity.
   if existing.Status == model.ArticleStatusPublished {
       // ...
   }
   ```

7. **Use Context**
   ```go
   // Good: Always pass context as first parameter
   func (s *Service) DoWork(ctx context.Context, param string) error {
       // Check for cancellation
       select {
       case <-ctx.Done():
           return ctx.Err()
       default:
           // Continue work
       }
   }
   ```

### Testing Guidelines

#### Unit Tests
```go
func TestArticleService_Create(t *testing.T) {
    // Arrange
    repo := &mockArticleRepository{}
    service := NewArticleService(repo, nil, nil, nil)
    
    article := &model.Article{
        Title: "Test Article",
        ArticleType: model.ArticleTypeNews,
    }
    
    // Act
    err := service.Create(context.Background(), article, "user123")
    
    // Assert
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if article.Slug == "" {
        t.Error("Expected slug to be generated")
    }
}
```

#### Integration Tests
```go
func TestArticleRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Setup test database
    client, _ := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
    db := client.Database("cms_test")
    defer db.Drop(context.Background())
    
    repo := NewArticleRepository(db)
    
    // Test operations
    // ...
}
```

### Logging Best Practices

```go
import "log"

// Good: Structured logging
log.Printf("Creating article: title=%s, type=%s, user=%s", article.Title, article.ArticleType, userID)

// For errors
log.Printf("Error creating article: %v", err)

// For important events
log.Println("✓ Article published successfully:", article.ID.Hex())
```

## Extending Article Types

Adding a new article type is straightforward. Follow these steps:

### Step 1: Define the Article Type

Edit `internal/model/article.go`:

```go
// Add to ArticleType constants
const (
    // ... existing types
    ArticleTypeNewType ArticleType = "NewType"  // Add your new type
)

// Add type-specific fields to Article struct
type Article struct {
    // ... common fields
    
    // NewType specific fields
    NewTypeField1 string     `json:"newTypeField1,omitempty" bson:"newTypeField1,omitempty"`
    NewTypeField2 int        `json:"newTypeField2,omitempty" bson:"newTypeField2,omitempty"`
    NewTypeData   *CustomObj `json:"newTypeData,omitempty" bson:"newTypeData,omitempty"`
}
```

### Step 2: Update Validation (if needed)

Edit `internal/service/article_service.go`:

```go
func (s *ArticleService) Create(ctx context.Context, article *model.Article, userID string) error {
    // Add validation for new type
    if article.ArticleType == model.ArticleTypeNewType {
        if article.NewTypeField1 == "" {
            return fmt.Errorf("newTypeField1 is required for NewType articles")
        }
    }
    
    // ... rest of creation logic
}
```

### Step 3: Update Seed Data (optional)

Edit `internal/migrations/initial.go`:

```go
categories := []*model.Category{
    // ... existing categories
    {
        Name:         "New Type Content",
        Slug:         "new-type",
        CategoryType: model.CategoryTypeArticle,
        ArticleType:  model.ArticleTypeNewType,
        Ordering:     6,
        CreatedBy:    "system",
    },
}
```

### Step 4: Update API Documentation

Edit `docs/openapi.yaml` to document the new fields.

### Step 5: Test

```go
func TestNewArticleType(t *testing.T) {
    article := &model.Article{
        Title:         "Test New Type",
        ArticleType:   model.ArticleTypeNewType,
        NewTypeField1: "required value",
    }
    
    // Test creation, update, etc.
}
```

### Example: Adding a "Recipe" Article Type

```go
// 1. In model/article.go
const ArticleTypeRecipe ArticleType = "Recipe"

type Article struct {
    // ... existing fields
    
    // Recipe-specific fields
    Ingredients    []string `json:"ingredients,omitempty" bson:"ingredients,omitempty"`
    Instructions   []string `json:"instructions,omitempty" bson:"instructions,omitempty"`
    PrepTime       int      `json:"prepTime,omitempty" bson:"prepTime,omitempty"` // minutes
    CookTime       int      `json:"cookTime,omitempty" bson:"cookTime,omitempty"` // minutes
    Servings       int      `json:"servings,omitempty" bson:"servings,omitempty"`
    Difficulty     string   `json:"difficulty,omitempty" bson:"difficulty,omitempty"` // Easy, Medium, Hard
    NutritionalInfo map[string]interface{} `json:"nutritionalInfo,omitempty" bson:"nutritionalInfo,omitempty"`
}

// 2. In service/article_service.go - Add validation
func (s *ArticleService) validateRecipe(article *model.Article) error {
    if len(article.Ingredients) == 0 {
        return fmt.Errorf("recipe must have at least one ingredient")
    }
    if len(article.Instructions) == 0 {
        return fmt.Errorf("recipe must have cooking instructions")
    }
    return nil
}
```

## API Documentation

See `docs/openapi.yaml` for complete API specification.

### Key Endpoints

#### Admin APIs (require authentication)
- `POST /api/v1/articles` - Create article
- `GET /api/v1/articles` - List articles (with filters)
- `GET /api/v1/articles/{id}` - Get article details
- `PATCH /api/v1/articles/{id}` - Update article
- `DELETE /api/v1/articles/{id}` - Delete article
- `POST /api/v1/articles/{id}/publish` - Publish article
- `POST /api/v1/articles/reorder` - Reorder articles
- `GET /api/v1/search` - Full-text search

#### Public APIs (cached, no auth required)
- `GET /api/v1/public/articles` - List published articles
- `GET /api/v1/public/articles/{id}` - Get published article
- `POST /api/v1/public/articles/{id}/view` - Record view

#### Category APIs
- `POST /api/v1/categories` - Create category
- `GET /api/v1/categories/tree` - Get category tree
- `GET /api/v1/categories/{id}` - Get category
- `PATCH /api/v1/categories/{id}` - Update category

#### Permission Group APIs
- `POST /api/v1/permission-groups` - Create permission group
- `GET /api/v1/permission-groups` - List permission groups
- `POST /api/v1/permission-groups/{id}/users` - Add user to group
- `POST /api/v1/permission-groups/{id}/categories` - Add category to group

## Testing

### Run All Tests
```bash
go test ./...
```

### Run with Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Integration Tests Only
```bash
go test -v -tags=integration ./...
```

### Run with Race Detector
```bash
go test -race ./...
```

## Deployment

### Docker Build
```bash
docker build -t article-service:latest -f services/article-service/Dockerfile .
```

### Environment Variables

Required:
- `MONGODB_URI` - MongoDB connection string
- `MONGODB_DATABASE` - Database name

Optional:
- `SERVER_PORT` - HTTP server port (default: 8080)
- `REDIS_ADDR` - Redis address (default: uses memory cache)
- `REDIS_PASSWORD` - Redis password
- `REDIS_DB` - Redis database number (default: 0)
- `RUN_MIGRATIONS` - Run migrations on startup (default: true)
- `CACHE_TTL` - Cache TTL in seconds (default: 300)
- `QUEUE_SIZE` - View queue size (default: 10000)
- `QUEUE_BATCH_SIZE` - View queue batch size (default: 100)
- `SCHEDULER_INTERVAL` - Scheduler interval (default: 60s)

## Contributing

### Code Review Checklist
- [ ] Code follows Go conventions and style guide
- [ ] All tests pass
- [ ] Code coverage is maintained or improved
- [ ] SonarQube quality gate passes
- [ ] Documentation is updated
- [ ] No security vulnerabilities
- [ ] Performance impact considered

### Git Workflow
1. Create feature branch from `main`
2. Make changes with clear commit messages
3. Run tests and linting
4. Create pull request
5. Address review comments
6. Squash and merge

## License

Copyright © 2024 VHV Platform. All rights reserved.

## Support

For issues and questions:
- Create an issue on GitHub
- Contact: dev@vhvplatform.com
