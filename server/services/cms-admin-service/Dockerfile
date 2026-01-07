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
COPY services/cms-admin-service ./services/cms-admin-service

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/cms-admin-service \
    ./services/cms-admin-service/cmd

# Final stage
FROM scratch

# Copy timezone data and CA certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/cms-admin-service /app/cms-admin-service

# Create non-root user
USER 65534:65534

WORKDIR /app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD ["/app/cms-admin-service", "--health"]

# Run the application
ENTRYPOINT ["/app/cms-admin-service"]
