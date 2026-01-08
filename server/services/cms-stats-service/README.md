# CMS Stats Service

A dedicated microservice for handling comments and statistics, separated from the main CMS service for better scalability and performance isolation.

## Features

- **Comment Management**: Full comment system with moderation
- **Nested Comments**: Support for threaded conversations (up to 3 levels)
- **Comment Likes**: Users can like/unlike comments
- **Comment Reporting**: Report inappropriate comments
- **Rate Limiting**: Prevent spam with configurable rate limits
- **Favorites**: Users can save favorite articles
- **Statistics**: View statistics and metrics

## Architecture

This service is designed to be independent from the main CMS service to:
- Prevent comment traffic from impacting core CMS operations
- Allow independent scaling based on comment load
- Enable separate deployment and updates
- Isolate database operations for better performance

## Configuration

Environment variables:

```bash
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=cms_comments
SERVER_PORT=8081
```

## API Endpoints

### Comments
- `POST /api/v1/articles/{id}/comments` - Create comment
- `GET /api/v1/articles/{id}/comments` - Get article comments
- `GET /api/v1/comments/{id}/replies` - Get comment replies
- `POST /api/v1/comments/{id}/moderate` - Moderate comment
- `POST /api/v1/comments/{id}/like` - Like comment
- `DELETE /api/v1/comments/{id}/like` - Unlike comment
- `POST /api/v1/comments/{id}/report` - Report comment
- `GET /api/v1/comments/pending` - Get pending comments (moderation)

### Favorites
- `POST /api/v1/articles/{id}/favorite` - Add to favorites
- `DELETE /api/v1/articles/{id}/favorite` - Remove from favorites
- `GET /api/v1/users/favorites` - Get user favorites

## Running the Service

### Local Development
```bash
cd services/cms-stats-service
go run cmd/main.go
```

### Docker
```bash
docker-compose up cms-stats-service
```

## Database

Uses MongoDB with separate database `cms_comments` to isolate data from main CMS.

Collections:
- `comments` - Comment data
- `comment_likes` - Like tracking
- `comment_reports` - Violation reports
- `user_rate_limits` - Rate limiting data
- `favorite_articles` - User favorites
