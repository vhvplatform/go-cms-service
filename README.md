# Go CMS Service

A comprehensive Content Management System microservice built with Go and MongoDB.

## Overview

This repository contains a production-ready article management service with support for multiple content types, multi-tenancy, advanced permissions, caching, and comprehensive statistics.

## Features

- ✅ **Multi-Type Articles**: 14+ article types (News, Video, Gallery, Legal, etc.)
- ✅ **Multi-Tenancy**: Tenant-based article type configuration
- ✅ **Advanced Permissions**: Role-based access + permission groups by categories
- ✅ **Caching**: Redis-based caching with auto-invalidation
- ✅ **Queue-Based Analytics**: Async view counting with batch processing
- ✅ **Scheduled Publishing**: Auto-publish and expire articles
- ✅ **Full-Text Search**: MongoDB text index based search
- ✅ **Editorial Workflow**: Draft → Review → Published → Archived
- ✅ **Access Control**: Public, login-required, role-based, premium content
- ✅ **Statistics & Reporting**: Comprehensive analytics and metrics

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/vhvplatform/go-cms-service.git
cd go-cms-service

# Start all services
docker-compose up -d

# Check service health
curl http://localhost:8080/health
```

### Local Development

```bash
# Install dependencies
cd services/cms-service
go mod download

# Set environment variables
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=cms
export REDIS_ADDR=localhost:6379
export SERVER_PORT=8080

# Run the service
go run cmd/main.go
```

## Documentation

- [**Development Guide**](services/cms-service/docs/DEVELOPMENT.md) - Complete architecture, code standards, and extension guide
- [**API Documentation**](services/cms-service/docs/openapi.yaml) - OpenAPI/Swagger specification

## Project Structure

```
services/cms-service/
├── cmd/                    # Application entry points
├── internal/
│   ├── cache/             # Caching layer (Redis + Memory)
│   ├── handler/           # HTTP request handlers
│   ├── middleware/        # Authentication & authorization
│   ├── migrations/        # Database migrations
│   ├── model/             # Data models
│   ├── repository/        # Data access layer
│   ├── service/           # Business logic
│   └── worker/            # Background workers
├── configs/               # Configuration files
├── docs/                  # Documentation
└── tests/                 # Test files
```

## Key Technologies

- **Language**: Go 1.21+
- **Database**: MongoDB 5.0+
- **Cache**: Redis 6.0+
- **Architecture**: Clean Architecture (Handler → Service → Repository)

## API Endpoints

### Admin APIs
- `POST /api/v1/articles` - Create article
- `GET /api/v1/articles` - List articles
- `PATCH /api/v1/articles/{id}` - Update article
- `POST /api/v1/articles/{id}/publish` - Publish article
- `POST /api/v1/categories` - Manage categories
- `POST /api/v1/permission-groups` - Manage permissions

### Public APIs (Cached)
- `GET /api/v1/public/articles` - List published articles
- `GET /api/v1/public/articles/{id}` - Get article details
- `POST /api/v1/public/articles/{id}/view` - Record view

## Article Types Supported

1. **News** - Standard news articles
2. **Video** - Video content with duration/thumbnail
3. **Photo Gallery** - Image collections
4. **Legal Document** - Legal texts with PDFs
5. **Staff Profile** - Employee information
6. **Job** - Job postings
7. **Procedure** - Process documentation
8. **Download** - Downloadable files
9. **Podcast** - Audio content
10. **Event Info** - Event information
11. **Infographic** - Visual graphics
12. **Destination** - Travel/location info
13. **Partner** - Partner information
14. **PDF** - PDF documents

**Easily extensible** - Add new types in minutes! See [Development Guide](services/article-service/docs/DEVELOPMENT.md#extending-article-types).

## Configuration

Environment variables:

```bash
# Required
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=cms

# Optional
SERVER_PORT=8080                    # HTTP server port
REDIS_ADDR=localhost:6379          # Redis address
REDIS_PASSWORD=                     # Redis password
RUN_MIGRATIONS=true                 # Run migrations on startup
CACHE_TTL=300                       # Cache TTL in seconds
QUEUE_SIZE=10000                    # View queue size
SCHEDULER_INTERVAL=60s              # Scheduler interval
```

## Development

### Code Quality Standards

- **No Code Duplication**: DRY principle
- **Max Complexity**: 15 per function
- **Error Handling**: No panics, proper error wrapping
- **Testing**: Unit + Integration tests
- **SonarQube**: Compliant with quality gates

### Running Tests

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Integration tests only
go test -tags=integration ./...
```

### Adding New Article Types

See the complete guide in [DEVELOPMENT.md](services/article-service/docs/DEVELOPMENT.md#extending-article-types).

Quick example:
```go
// 1. Add to model/article.go
const ArticleTypeRecipe ArticleType = "Recipe"

// 2. Add type-specific fields
type Article struct {
    // ... common fields
    Ingredients []string `json:"ingredients,omitempty"`
    PrepTime    int      `json:"prepTime,omitempty"`
}

// 3. Done! Use immediately in APIs
```

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## License

Copyright © 2024 VHV Platform. All rights reserved.

## Support

- **Issues**: [GitHub Issues](https://github.com/vhvplatform/go-cms-service/issues)
- **Email**: dev@vhvplatform.com