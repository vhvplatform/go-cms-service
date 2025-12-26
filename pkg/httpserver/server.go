package httpserver

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vhvplatform/go-cms-service/pkg/logger"
)

// Server wraps http.Server with graceful shutdown
type Server struct {
	server *http.Server
	logger *logger.Logger
}

// Config holds server configuration
type Config struct {
	Port         string
	Handler      http.Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// NewServer creates a new HTTP server
func NewServer(cfg *Config, log *logger.Logger) *Server {
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      cfg.Handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		server: srv,
		logger: log,
	}
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		s.logger.Info("Server starting on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	s.logger.Info("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown: %v", err)
		return err
	}

	s.logger.Info("Server stopped gracefully")
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// DefaultConfig returns a server configuration with sensible defaults
func DefaultConfig(port string, handler http.Handler) *Config {
	return &Config{
		Port:         port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
