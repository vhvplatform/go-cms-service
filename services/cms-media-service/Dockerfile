FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY services/cms-media-service ./services/cms-media-service

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /cms-media-service ./services/cms-media-service/cmd

# Final stage
FROM alpine:latest

# Install runtime dependencies for media processing
RUN apk add --no-cache \
    ffmpeg \
    imagemagick \
    libreoffice \
    poppler-utils \
    ca-certificates \
    tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /cms-media-service .

# Create uploads directory
RUN mkdir -p /root/uploads

# Expose port
EXPOSE 8083

# Run the application
CMD ["./cms-media-service"]
