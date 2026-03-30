package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	limitdomain "air-social/internal/domain/limit"
)

// PRE-COMPILED LUA SCRIPT (GLOBAL)
var slidingWindowScript = redis.NewScript(`
	-- ARGV mapping
	local limit       = tonumber(ARGV[1])
	local window      = tonumber(ARGV[2])
	local now         = tonumber(ARGV[3])

	-- MENTAL MODEL: "Snapping to Grid" (Bucketing)
	-- math.floor(now / window) finds the index of the current time bucket.
	-- Multiplying back by window gives the exact start timestamp of this bucket.
	local window_start = math.floor(now / window) * window
	
	-- elapsed: How much time has passed since the current window started.
	local elapsed = now - window_start
	
	-- MENTAL MODEL: "The Shadow of the Past" (Overlap Weight)
	-- Linear interpolation: We assume previous requests were distributed evenly.
	-- If we are 25% into the current window, we still carry 75% of the previous window's requests.
	local prev_weight = 1 - (elapsed / window)

	-- Retrieve actual counts from Redis (0 if key doesn't exist yet)
	local prev_count    = tonumber(redis.call('GET', KEYS[2]) or 0)
	local current_count = tonumber(redis.call('GET', KEYS[1]) or 0)
	
	-- MENTAL MODEL: The Estimation
	-- Total requests = Current requests + (Previous requests * Overlap Weight)
	local estimated = current_count + math.floor(prev_count * prev_weight)
	
	if estimated >= limit then
		-- Blocked. Return remaining = 0, and the timestamp when this window resets
		return {0, estimated, window_start + window}
	end

	-- Allowed. Increment the current bucket counter
	local new_count = redis.call('INCR', KEYS[1])
	if new_count == 1 then
		-- MENTAL MODEL: "Two Lifetimes"
		-- The key needs to survive for this current window, PLUS one more window 
		-- to act as the 'previous window' for future calculations.
		redis.call('EXPIRE', KEYS[1], window * 2)
	end

	-- Return success. Recalculate estimated to include this newly added request.
	return {1, new_count + math.floor(prev_count * prev_weight), window_start + window}
`)

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (limitdomain.LimitResult, error) {
	var empty limitdomain.LimitResult

	windowSecs := int64(window.Seconds())
	nowSecs := int64(time.Now().Unix())

	windowStart := (nowSecs / windowSecs) * windowSecs
	prevStart := windowStart - windowSecs
	currentKey := fmt.Sprintf("%s:%d", key, windowStart)
	prevKey := fmt.Sprintf("%s:%d", key, prevStart)

	result, err := slidingWindowScript.Run(
		ctx,
		r.client,
		[]string{currentKey, prevKey}, // KEYS[1], KEYS[2]
		limit, windowSecs, nowSecs,    // ARGV[1], ARGV[2], ARGV[3]
	).Slice()

	if err != nil {
		return empty, err
	}

	allowed := result[0].(int64)
	estimated := result[1].(int64)
	resetUnix := result[2].(int64)
	remaining := max(int64(limit)-estimated, 0)

	return limitdomain.LimitResult{
		Allowed:   allowed == int64(1),  
		Limit:     limit,
		Remaining: int(remaining),
		ResetAt:   time.Unix(resetUnix, 0),
	}, nil
}
