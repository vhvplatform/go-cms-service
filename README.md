# Go CMS Service

A comprehensive Content Management System built with microservices architecture using Go and MongoDB.

## Overview

This repository contains backend microservices built with Go, featuring clean architecture and modern development practices.

## Repository Structure

```
.
├── pkg/                   # Shared infrastructure packages
├── services/              # Microservices
│   ├── cms-admin-service/     # Main content management
│   ├── cms-stats-service/     # Comments & statistics
│   ├── cms-frontend-service/  # Public API with caching
│   ├── cms-media-service/     # Media processing
│   └── cms-crawler-service/   # Content crawler
├── docs/                  # Project documentation
│   ├── README.md          # Detailed documentation
│   ├── README_VI.md       # Vietnamese documentation
│   └── ARCHITECTURE.md    # Architecture details
├── Makefile               # Build automation
├── docker-compose.yml     # Service orchestration
├── go.mod                 # Go module definition
└── .github/workflows/     # CI/CD pipelines
```

## Microservices

### 🖥️ Backend Services
Go-based microservices with clean architecture:
- **CMS Admin Service** (Port 8080) - Content management, permissions, workflows
- **CMS Stats Service** (Port 8081) - Comments, likes, statistics
- **CMS Frontend Service** (Port 8082) - Public-facing API with caching
- **CMS Media Service** (Port 8083) - Media processing and storage
- **CMS Crawler Service** (Port 8084) - Automated content collection

📖 See [docs/README.md](docs/README.md) for detailed documentation.

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

## Features

### Core Content Management
- ✅ **Multi-Type Articles**: 14+ article types (News, Video, Gallery, Legal, etc.)
- ✅ **Multi-Tenancy**: Tenant-based article type configuration
- ✅ **Advanced Permissions**: Role-based access + permission groups by categories
- ✅ **Caching**: Redis-based caching with auto-invalidation
- ✅ **Scheduled Publishing**: Auto-publish and expire articles
- ✅ **Full-Text Search**: MongoDB text index based search
- ✅ **Editorial Workflow**: Draft → Review → Published → Archived

For more features and detailed information, see [docs/README.md](docs/README.md).

## Documentation

- **[Backend Documentation](docs/README.md)** - Complete backend microservices documentation
- **[Architecture Guide](docs/ARCHITECTURE.md)** - System architecture and design patterns
- **[Vietnamese Documentation](docs/README_VI.md)** - Tài liệu tiếng Việt
- **[Development Guide](services/cms-admin-service/docs/DEVELOPMENT.md)** - Development and extension guide

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