# Go CMS Backend Services

This directory contains the backend Go microservices for the CMS platform.

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Start all services
docker-compose up -d

# Check services health
curl http://localhost:8080/health  # CMS Admin Service
curl http://localhost:8081/health  # CMS Stats Service
curl http://localhost:8082/health  # CMS Frontend Service
curl http://localhost:8083/health  # CMS Media Service
curl http://localhost:8084/health  # CMS Crawler Service
```

### Local Development

```bash
# Install dependencies
go mod download

# Run a specific service
cd services/cms-admin-service
go run cmd/main.go
```

## Project Structure

```
server/
├── pkg/                    # Shared infrastructure packages
│   ├── config/            # Configuration management
│   ├── database/          # Database utilities
│   ├── errors/            # Error handling
│   ├── httpserver/        # HTTP server setup
│   ├── logger/            # Structured logging
│   └── middleware/        # Common middleware
├── services/              # Microservices
│   ├── cms-admin-service/     # Main content management (Port 8080)
│   ├── cms-stats-service/     # Comments & statistics (Port 8081)
│   ├── cms-frontend-service/  # Public API with caching (Port 8082)
│   ├── cms-media-service/     # Media processing (Port 8083)
│   └── cms-crawler-service/   # Content crawler (Port 8084)
├── Makefile              # Build automation
├── docker-compose.yml    # Service orchestration
└── go.mod                # Go module definition
```

## Microservices

### 1. CMS Admin Service (Port 8080)
Main content management service handling:
- Articles management (14+ types)
- Categories and taxonomies
- Permissions and workflows
- Version control
- AI features (spell check, translation, etc.)

### 2. CMS Stats Service (Port 8081)
Dedicated service for user engagement:
- Comments system (3-level nesting)
- Likes and favorites
- Statistics and analytics

### 3. CMS Frontend Service (Port 8082)
Public-facing API with:
- Redis caching
- Service composition
- Optimized for performance

### 4. CMS Media Service (Port 8083)
Advanced media processing:
- Image compression and thumbnails
- Video HLS encoding
- Document thumbnails
- Storage management

### 5. CMS Crawler Service (Port 8084)
Automated content collection:
- Web scraping
- Campaign management
- Duplicate detection
- Anti-crawler bypass

## Development Commands

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
make golangci-lint     # Run golangci-lint

# Docker operations
make docker-build      # Build all Docker images
make docker-up         # Start all services
make docker-down       # Stop all services
make docker-logs       # View logs
```

## Configuration

Environment variables:

```bash
# Database
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=cms

# Cache
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Server
SERVER_PORT=8080
BASE_URL=http://localhost:8080

# Features
RUN_MIGRATIONS=true
CACHE_TTL=300
LOG_LEVEL=info
```

## Documentation

For detailed documentation, see:
- [Complete Documentation](../docs/README.md)
- [Architecture Guide](../docs/ARCHITECTURE.md)
- [Development Guide](services/cms-admin-service/docs/DEVELOPMENT.md)
- [API Documentation](services/cms-admin-service/docs/openapi.yaml)
