package bootstrap

import (
	"github.com/redis/go-redis/v9"

	"air-social/internal/config"
)

func NewRedisClient(rc config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     rc.Addr,
		Password: rc.Password,
		DB:       rc.DB,
	})
}
