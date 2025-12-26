package config

import (
	"os"
	"strconv"
	"time"
)

// GetEnv retrieves an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt retrieves an environment variable as an integer or returns a default value
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetEnvBool retrieves an environment variable as a boolean or returns a default value
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// GetEnvDuration retrieves an environment variable as a duration or returns a default value
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Config holds common application configuration
type Config struct {
	ServerPort    string
	LogLevel      string
	Environment   string
	ServiceName   string
}

// NewConfig creates a new Config with defaults
func NewConfig(serviceName string) *Config {
	return &Config{
		ServerPort:  GetEnv("SERVER_PORT", "8080"),
		LogLevel:    GetEnv("LOG_LEVEL", "info"),
		Environment: GetEnv("ENVIRONMENT", "development"),
		ServiceName: serviceName,
	}
}

// MongoConfig holds MongoDB configuration
type MongoConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// NewMongoConfig creates a new MongoDB configuration
func NewMongoConfig() *MongoConfig {
	return &MongoConfig{
		URI:      GetEnv("MONGODB_URI", "mongodb://localhost:27017"),
		Database: GetEnv("MONGODB_DATABASE", "cms"),
		Timeout:  GetEnvDuration("MONGODB_TIMEOUT", 10*time.Second),
	}
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisConfig creates a new Redis configuration
func NewRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:     GetEnv("REDIS_ADDR", "localhost:6379"),
		Password: GetEnv("REDIS_PASSWORD", ""),
		DB:       GetEnvInt("REDIS_DB", 0),
	}
}
