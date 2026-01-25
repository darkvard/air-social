package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"air-social/internal/config"
	"air-social/pkg"
)

func NewConnection(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:       cfg.Addr,
		Password:   cfg.Password,
		DB:         cfg.DB,
		MaxRetries: 1,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := pkg.Retry(ctx, 10, 2*time.Second, func() error {
		return client.Ping(ctx).Err()
	}); err != nil {
		return nil, fmt.Errorf("redis: failed to connect: %w", err)
	}

	return client, nil
}

func NewRedisCache(client *redis.Client) (*redisCache, error) {
	if client == nil {
		return nil, errors.New("redis client cannot nil")
	}
	return newRedisCache(client), nil
}