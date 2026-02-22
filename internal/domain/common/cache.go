package common

import (
	"context"
	"fmt"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string, dst any) error
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	SetNX(ctx context.Context, key string, val any, ttl time.Duration) (bool, error)
	Delete(ctx context.Context, key string) error
	IsExist(ctx context.Context, key string) (bool, error)
}

// <system>:<feature>:<state>:<id>
func BuildCacheKey(system, feature, state string, id any) string {
	return fmt.Sprintf("%s:%s:%s:%v", system, feature, state, id)
}
