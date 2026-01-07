# Go CMS Service

A comprehensive Content Management System built with microservices architecture using Go and MongoDB, with React.js frontend and Flutter mobile app.

## Overview

This repository contains a complete CMS platform with:
- **Backend microservices** built with Go
- **Frontend web application** built with React.js
- **Mobile application** built with Flutter
- **Comprehensive documentation** for all components

## Repository Structure

```
.
├── server/                 # Backend Go microservices
│   ├── pkg/               # Shared infrastructure packages
│   ├── services/          # Microservices
│   ├── Makefile           # Build automation
│   └── docker-compose.yml # Service orchestration
├── client/                # Frontend React.js application
├── flutter/               # Mobile Flutter application
├── docs/                  # Project documentation
│   ├── README.md          # Detailed documentation
│   ├── README_VI.md       # Vietnamese documentation
│   └── ARCHITECTURE.md    # Architecture details
└── .github/workflows/     # CI/CD pipelines
```

## Components

### 🖥️ Server (Backend)
Go-based microservices with clean architecture:
- **CMS Admin Service** - Content management, permissions, workflows
- **CMS Stats Service** - Comments, likes, statistics
- **CMS Frontend Service** - Public-facing API with caching
- **CMS Media Service** - Media processing and storage
- **CMS Crawler Service** - Automated content collection

📖 See [server/README.md](docs/README.md) for detailed backend documentation.

### 🌐 Client (Frontend)
React.js web application for content management interface.

📖 See [client/README.md](client/README.md) for frontend documentation.

### 📱 Flutter (Mobile)
Cross-platform mobile application for iOS and Android.

📖 See [flutter/README.md](flutter/README.md) for mobile app documentation.

## Quick Start

### Backend Services
```bash
cd server
docker-compose up -d
```

### Frontend Application
```bash
cd client
# Instructions coming soon
```

### Mobile Application
```bash
cd flutter
# Instructions coming soon
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
- **[Development Guide](server/services/cms-admin-service/docs/DEVELOPMENT.md)** - Development and extension guide

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