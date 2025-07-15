package cache

import (
	"context"
	"time"
	"vybes/internal/config"

	"github.com/redis/go-redis/v9"
)

// Client defines the interface for a cache client that provides
// basic caching operations like get, set, and delete.
type Client interface {
	// Get retrieves a value from cache by key
	Get(ctx context.Context, key string) (string, error)
	// Set stores a value in cache with optional expiration
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	// Del removes one or more keys from cache
	Del(ctx context.Context, keys ...string) error
}

// redisClient implements the Client interface using Redis as the backend
type redisClient struct {
	client *redis.Client
}

// NewClient creates and initializes a new Redis client with the provided configuration.
// It establishes a connection to Redis and verifies connectivity before returning.
//
// Parameters:
//   - cfg: Configuration containing Redis connection details (address, password, database)
//
// Returns:
//   - Client: A configured cache client ready for use
//   - error: Any error that occurred during client initialization
func NewClient(cfg *config.Config) (Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Verify Redis connection by sending a ping command
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &redisClient{client: rdb}, nil
}

// Get retrieves a value from Redis cache by its key.
// Returns an empty string and error if the key doesn't exist or on connection issues.
func (c *redisClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set stores a value in Redis cache with an optional expiration time.
// The value is automatically serialized to JSON if it's not a string.
func (c *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Del removes one or more keys from Redis cache.
// Returns the number of keys that were actually deleted.
func (c *redisClient) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}
