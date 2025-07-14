package cache

import (
	"context"
	"time"
	"vybes/internal/config"

	"github.com/redis/go-redis/v9"
)

// Client defines the interface for a cache client.
type Client interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

type redisClient struct {
	client *redis.Client
}

// NewClient creates a new Redis client.
func NewClient(cfg *config.Config) (Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Ping the server to check the connection.
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &redisClient{client: rdb}, nil
}

func (c *redisClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *redisClient) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}