package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache interface for caching operations
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeletePattern(ctx context.Context, pattern string) error
}

// RedisCache implements Cache using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Get retrieves a value from cache
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("cache miss")
	}
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in cache
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes keys from cache
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// DeletePattern removes all keys matching a pattern
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var keys []string

	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// MemoryCache implements Cache using in-memory map (for testing/fallback)
type MemoryCache struct {
	data map[string]cacheEntry
}

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]cacheEntry),
	}
}

// Get retrieves a value from cache
func (c *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	entry, exists := c.data[key]
	if !exists || time.Now().After(entry.expiresAt) {
		delete(c.data, key)
		return fmt.Errorf("cache miss")
	}

	return json.Unmarshal(entry.value, dest)
}

// Set stores a value in cache
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	c.data[key] = cacheEntry{
		value:     data,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes keys from cache
func (c *MemoryCache) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(c.data, key)
	}
	return nil
}

// DeletePattern removes all keys matching a pattern (simple prefix match)
func (c *MemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	// Simple pattern matching: remove trailing * and match prefix
	prefix := pattern
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix = pattern[:len(pattern)-1]
	}

	keysToDelete := []string{}
	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(c.data, key)
	}

	return nil
}
