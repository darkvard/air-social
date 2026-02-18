package shared

import (
	"context"
	"fmt"
	"time"
)

type CacheStorage interface {
	Get(ctx context.Context, key string, dst any) error
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	IsExist(ctx context.Context, key string) (bool, error)
}

// key is system:feature:state:id
func BuildCacheKey(system, feature, state string, id any) string {
	return fmt.Sprintf("%s:%s:%s:%v", system, feature, state, id)
}
