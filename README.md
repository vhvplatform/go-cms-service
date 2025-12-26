# Go CMS Service

A comprehensive Content Management System built with microservices architecture using Go and MongoDB.

## Overview

This repository contains production-ready CMS microservices with support for multiple content types, multi-tenancy, advanced permissions, caching, and comprehensive user engagement features.

## Architecture

The project follows modern Go architectural standards with:

- **Shared Infrastructure (`pkg/`)**: Reusable packages for configuration, logging, database connections, HTTP server setup, error handling, and middleware
- **Clean Architecture**: Clear separation between handlers, services, and repositories
- **Microservices Pattern**: Independent, scalable services with focused responsibilities
- **Containerization**: Docker-optimized builds with multi-stage compilation and security best practices
- **CI/CD**: Automated testing, linting, and deployment pipelines
- **Graceful Shutdown**: Proper signal handling and connection cleanup
- **Structured Logging**: Centralized logging with log levels and service context

### Project Structure

```
.
├── pkg/                    # Shared infrastructure packages
│   ├── config/            # Configuration management
│   ├── database/          # Database utilities
│   ├── errors/            # Error handling
│   ├── httpserver/        # HTTP server setup
│   ├── logger/            # Structured logging
│   └── middleware/        # Common middleware
├── services/              # Microservices
│   ├── cms-admin-service/
│   ├── cms-stats-service/
│   ├── cms-frontend-service/
│   ├── cms-media-service/
│   └── cms-crawler-service/
├── Makefile              # Build automation
├── docker-compose.yml    # Service orchestration
└── .github/workflows/    # CI/CD pipelines
```

See [pkg/README.md](pkg/README.md) for detailed documentation on shared infrastructure components.

## Microservices

### 1. CMS Admin Service (Port 8080)
Main content management service handling articles, categories, permissions, workflows, version control, and AI features.

### 2. CMS Stats Service (Port 8081)
Dedicated service for comments, likes, favorites, and statistics - isolated for better scalability and performance.

### 3. CMS Frontend Service (Port 8082)
Public-facing API with Redis caching and service composition for optimal end-user performance.

### 4. CMS Media Service (Port 8083)
Advanced media processing including image compression, video HLS encoding, document thumbnails, and storage management.

### 5. CMS Crawler Service (Port 8084)
Automated content collection from external sources with anti-crawler bypass, duplicate detection, and similarity grouping.

## Features

### Core Content Management
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

### Advanced Features
- ✅ **Action Logs**: Complete audit trail of all article operations
- ✅ **Version Management**: Track and restore previous versions of articles
- ✅ **Rejection Notes**: Conversation threads for article feedback
- ✅ **Social Media Sharing**: One-click sharing to Facebook, Twitter, LinkedIn
- ✅ **Input Validation**: Comprehensive validation for all API endpoints
- ✅ **Comment System**: Nested comments (3 levels) with moderation and rate limiting
- ✅ **Polls & Surveys**: Configurable polls with vote tracking
- ✅ **Related Articles**: Manual and automatic article linking
- ✅ **RSS Feed**: RSS 2.0 generation with media enclosures

### AI & Content Intelligence
- ✅ **Sensitive Keyword Detection**: Pattern-based content moderation
- ✅ **AI Spell Check**: Multi-provider spell and grammar checking
- ✅ **AI Translation**: Translate to multiple languages
- ✅ **Content Improvement**: AI-powered content suggestions
- ✅ **Violation Detection**: Automatic policy violation detection

### Media Processing
- ✅ **Image Optimization**: Automatic compression and thumbnail generation
- ✅ **Video HLS Encoding**: Adaptive bitrate streaming (360p, 720p, 1080p)
- ✅ **Document Thumbnails**: PDF, DOCX, PPTX thumbnail extraction
- ✅ **Storage Management**: Per-tenant quota tracking and usage stats
- ✅ **Auto Image Download**: Download external images to local storage

### Content Automation
- ✅ **Web Crawler**: Automated article collection from external sources
- ✅ **Campaign Management**: Scheduled crawling with cron expressions
- ✅ **Dynamic Extraction**: CSS selector and XPath-based extraction
- ✅ **Anti-Crawler Bypass**: User agents, proxies, delays, custom headers
- ✅ **Duplicate Detection**: Content hash-based deduplication
- ✅ **Similar Grouping**: Auto-group similar articles from different sources
- ✅ **Auto-Approval**: Configure sources for automatic approval
- ✅ **Post-Review**: Review workflow for approved crawled content

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/vhvplatform/go-cms-service.git
cd go-cms-service

# Start all services
docker-compose up -d

# Check services health
curl http://localhost:8080/health  # CMS Service
curl http://localhost:8081/health  # CMS Stats Service
```

### Local Development

```bash
# Start CMS Service
cd services/cms-service
go mod download
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=cms
export REDIS_ADDR=localhost:6379
export SERVER_PORT=8080
go run cmd/main.go

# Start CMS Stats Service (in another terminal)
cd services/cms-stats-service
go mod download
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=cms_comments
export SERVER_PORT=8081
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

### Build System

The project uses a comprehensive Makefile for building and testing:

```bash
# Build all services
make build

# Build specific service
make build-admin
make build-stats
make build-frontend
make build-media
make build-crawler

# Run tests
make test              # All tests
make test-pkg          # Shared infrastructure tests
make test-services     # Service-specific tests
make test-coverage     # Generate coverage report

# Linting and formatting
make lint              # Run linters and format code
make lint-check        # Check formatting without modifying
make golangci-lint     # Run golangci-lint

# Docker operations
make docker-build      # Build all Docker images
make docker-up         # Start all services
make docker-down       # Stop all services
make docker-logs       # View logs

# Development tools
make deps              # Download dependencies
make tidy              # Tidy dependencies
make install-tools     # Install development tools
make clean             # Clean build artifacts

# Run all verification checks
make verify            # Lint + Test + golangci-lint
make ci                # Full CI pipeline locally
```

### Code Quality Standards

- **No Code Duplication**: DRY principle
- **Max Complexity**: 15 per function
- **Error Handling**: No panics, proper error wrapping with `pkg/errors`
- **Structured Logging**: Use `pkg/logger` for all logging
- **Testing**: Unit + Integration tests
- **Linting**: golangci-lint compliant
- **Security**: Docker images run as non-root users

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