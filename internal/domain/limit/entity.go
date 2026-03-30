package limit

import (
	"context"
	"time"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (LimitResult, error)
}

type LimitResult struct {
	Allowed   bool
	Limit     int
	Remaining int       // max(0, limit - estimated)
	ResetAt   time.Time // unix timestamp khi window reset
}
