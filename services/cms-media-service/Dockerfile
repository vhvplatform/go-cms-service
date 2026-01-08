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
COPY services/cms-media-service ./services/cms-media-service

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/cms-media-service \
    ./services/cms-media-service/cmd

# Final stage - using alpine for media processing tools
FROM alpine:latest

# Install runtime dependencies for media processing
RUN apk add --no-cache \
    ffmpeg \
    imagemagick \
    libreoffice \
    poppler-utils \
    ca-certificates \
    tzdata

# Copy the binary
COPY --from=builder /app/cms-media-service /app/cms-media-service

# Create uploads directory
RUN mkdir -p /app/uploads && chown -R nobody:nobody /app

# Create non-root user
USER nobody:nobody

WORKDIR /app

# Expose port
EXPOSE 8083

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8083/health"]

# Run the application
ENTRYPOINT ["/app/cms-media-service"]
