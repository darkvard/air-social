package cache

import (
	"context"
	"time"
)

// <system>:<feature>:<state>:<id>
const (
	WorkerEmailProcessed = "worker:email:processed:"
	WorkerEmailVerify    = "worker:email:verify:"
	WorkerEmailReset     = "worker:email:reset:"
	WorkerEmailRetry     = "worker:email:retry:"
)

type CacheStorage interface {
	Get(ctx context.Context, key string, dst any) error
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	IsExist(ctx context.Context, key string) (bool, error)
}
