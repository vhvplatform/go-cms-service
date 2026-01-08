# Shared Infrastructure Package (pkg)

This directory contains shared infrastructure components that are used across all microservices in the go-cms-service project. These packages provide standardized implementations for common functionality, ensuring consistency and reducing code duplication.

## Structure

```
pkg/
├── config/         # Configuration management
├── database/       # Database connection utilities
├── errors/         # Standardized error handling
├── httpserver/     # HTTP server setup and graceful shutdown
├── logger/         # Structured logging
└── middleware/     # Common HTTP middleware
```

## Packages

### config

Provides standardized configuration management with environment variable support.

**Features:**
- Type-safe environment variable parsing (string, int, bool, duration)
- Configuration structs for common services (MongoDB, Redis)
- Default value support

**Usage:**
```go
import "github.com/vhvplatform/go-cms-service/pkg/config"

cfg := config.NewConfig("my-service")
mongoCfg := config.NewMongoConfig()
redisCfg := config.NewRedisConfig()
```

### database

Provides database connection utilities with proper error handling.

**Features:**
- MongoDB connection with timeout support
- Connection health checking
- Graceful disconnection

**Usage:**
```go
import "github.com/vhvplatform/go-cms-service/pkg/database"

mongoClient, err := database.ConnectMongo(ctx, uri, dbName, timeout)
if err != nil {
    log.Fatal(err)
}
defer mongoClient.Close(ctx)
```

### errors

Provides standardized error handling with context and error codes.

**Features:**
- Error wrapping with context
- Common error types (NotFound, InvalidInput, etc.)
- Error code constants

**Usage:**
```go
import "github.com/vhvplatform/go-cms-service/pkg/errors"

if item == nil {
    return errors.NotFound("item not found")
}

if err := db.Query(); err != nil {
    return errors.Database("query failed", err)
}
```

### httpserver

Provides HTTP server setup with graceful shutdown.

**Features:**
- Configurable timeouts
- Graceful shutdown on SIGINT/SIGTERM
- Sensible defaults

**Usage:**
```go
import "github.com/vhvplatform/go-cms-service/pkg/httpserver"

cfg := httpserver.DefaultConfig(port, handler)
server := httpserver.NewServer(cfg, log)
if err := server.Start(); err != nil {
    log.Fatal(err)
}
```

### logger

Provides structured logging with log levels.

**Features:**
- Multiple log levels (DEBUG, INFO, WARN, ERROR, FATAL)
- Service name prefix
- Context support

**Usage:**
```go
import "github.com/vhvplatform/go-cms-service/pkg/logger"

log := logger.New("my-service", "info")
log.Info("Server starting on port %s", port)
log.Error("Failed to connect: %v", err)
```

### middleware

Provides common HTTP middleware.

**Features:**
- Logging middleware (logs all requests with duration)
- Recovery middleware (recovers from panics)
- CORS middleware (configurable origin support)
- Middleware chaining

**Usage:**
```go
import "github.com/vhvplatform/go-cms-service/pkg/middleware"

handler := middleware.Chain(
    middleware.LoggingMiddleware(log),
    middleware.RecoveryMiddleware(log),
    middleware.CORSMiddleware([]string{"*"}),
)(mux)
```

## Design Principles

1. **Simplicity**: Keep interfaces simple and focused
2. **Consistency**: Provide consistent APIs across all packages
3. **Configurability**: Allow configuration while providing sensible defaults
4. **Error Handling**: Always return errors, never panic
5. **Context Awareness**: Support context for cancellation and timeouts
6. **Zero Dependencies**: Minimize external dependencies where possible

## Adding New Packages

When adding new shared functionality:

1. Create a new directory under `pkg/`
2. Keep the package focused on a single responsibility
3. Provide clear documentation and examples
4. Add tests for the new functionality
5. Update this README with package information

## Testing

Each package should have comprehensive unit tests:

```bash
# Test all packages
make test-pkg

# Test specific package
go test ./pkg/config/...
```

## Migration Guide

When migrating existing services to use these packages:

1. Replace custom configuration loading with `pkg/config`
2. Replace standard `log` with `pkg/logger`
3. Replace custom MongoDB connection with `pkg/database`
4. Replace custom server setup with `pkg/httpserver`
5. Add common middleware using `pkg/middleware`
6. Standardize error handling with `pkg/errors`

See `services/cms-admin-service/cmd/main.go` for a reference implementation.
