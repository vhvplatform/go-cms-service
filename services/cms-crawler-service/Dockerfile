# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files for dependency caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy pkg directory (shared infrastructure)
COPY pkg ./pkg

# Copy service source code
COPY services/cms-crawler-service ./services/cms-crawler-service

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/cms-crawler-service \
    ./services/cms-crawler-service/cmd

# Final stage
FROM alpine:latest

# Install runtime dependencies for web crawling
RUN apk add --no-cache ca-certificates wget tzdata

# Copy the binary
COPY --from=builder /app/cms-crawler-service /app/cms-crawler-service

# Create non-root user
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
USER appuser:appuser

WORKDIR /app

# Expose port
EXPOSE 8084

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8084/health"]

# Run the application
ENTRYPOINT ["/app/cms-crawler-service"]
