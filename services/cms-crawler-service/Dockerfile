FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY services/cms-crawler-service/go.mod services/cms-crawler-service/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY services/cms-crawler-service/ ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/crawler-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /app

COPY --from=builder /app/crawler-service .

EXPOSE 8084

CMD ["./crawler-service"]
