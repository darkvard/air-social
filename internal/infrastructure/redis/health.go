package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Health struct {
	client *redis.Client
}

func NewHealth(client *redis.Client) *Health {
	return &Health{
		client: client,
	}
}


func (h *Health) Ping(ctx context.Context) error {
	return h.client.Ping(ctx).Err()
}

