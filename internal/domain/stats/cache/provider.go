package cache

import (
	"context"
	"strconv"

	"air-social/internal/domain/common"
)

const (
	SystemName   = "social"
	FeatureStats = "stats"
)

const (
	StatePostLikes      = "post_likes"
	StatePostShares     = "post_shares"
	StatePostComments   = "post_comments"
	StateCommentLikes   = "comment_likes"
	StateCommentReplies = "comment_replies"
)

type Provider interface {
	GetStatsHash(ctx context.Context, state string) (map[int64]int64, error)
	UpdateStatsHash(ctx context.Context, state string, id int64, incr int64) error
	ClearSyncedFields(ctx context.Context, state string, syncData map[int64]int64) error

	GetStatsOffsets(ctx context.Context, state string, ids []int64) (map[int64]int64, error)
}

type provider struct {
	cache common.HashCache
}

func NewProvider(c common.HashCache) *provider {
	return &provider{cache: c}
}

func (p *provider) GetStatsHash(ctx context.Context, state string) (map[int64]int64, error) {
	dataStr, err := p.cache.HGetAll(ctx, getKey(state))
	if err != nil {
		return nil, err
	}

	result := make(map[int64]int64, len(dataStr))
	for k, v := range dataStr {
		id, err1 := strconv.ParseInt(k, 10, 64)
		val, err2 := strconv.ParseInt(v, 10, 64)
		if err1 == nil && err2 == nil && val != 0 {
			result[id] = val
		}
	}

	return result, nil
}

func (p *provider) ClearSyncedFields(ctx context.Context, state string, syncData map[int64]int64) error {
	if len(syncData) == 0 {
		return nil
	}

	// We do not directly clear the field (HDEL) because another goroutine/process
	// may update the same field concurrently.
	//
	// Instead, we atomically decrement the value using a Lua script.
	// If the resulting value becomes 0, we remove the field.
	//
	// This avoids race conditions where:
	// 1. One process clears the cache
	// 2. Another process increments the field at the same time
	// which could lead to lost updates.
	//
	// The Lua script ensures the decrement + conditional delete happens atomically in Redis.
	script := `
		local current = redis.call('HINCRBY', KEYS[1], ARGV[1], ARGV[2])
		if current == 0 then
			redis.call('HDEL', KEYS[1], ARGV[1])
		end
		return current
	`

	for id, val := range syncData {
		field := strconv.FormatInt(id, 10)
		decrement := -val
		_, _ = p.cache.Eval(ctx, script, []string{getKey(state)}, field, decrement)
	}
	return nil
}

func (p *provider) UpdateStatsHash(ctx context.Context, state string, id int64, incr int64) error {
	field := strconv.FormatInt(id, 10)
	_, err := p.cache.HIncrBy(ctx, getKey(state), field, incr)
	return err
}

func (p *provider) GetStatsOffsets(ctx context.Context, state string, ids []int64) (map[int64]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	fields := make([]string, len(ids))
	for i, id := range ids {
		fields[i] = strconv.FormatInt(id, 10)
	}

	vals, err := p.cache.HMGet(ctx, getKey(state), fields...)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]int64, len(ids))
	for i, v := range vals {
		if intVal, err := strconv.ParseInt(v, 10, 64); err == nil {
			result[ids[i]] = intVal
		}
	}

	return result, nil
}

func getKey(state string) string {
	return common.BuildCacheKey(SystemName, FeatureStats, state, "")
}
