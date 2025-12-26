FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY services/cms-stats-service ./services/cms-stats-service

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /cms-stats-service ./services/cms-stats-service/cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /cms-stats-service .

# Expose port
EXPOSE 8081

# Run the application
CMD ["./cms-stats-service"]
