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
}

type provider struct {
	cache common.Cache
}

func NewProvider(c common.Cache) *provider {
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
		if err1 == nil && err2 == nil {
			result[id] = val
		}
	}

	return result, nil
}

func (p *provider) ClearSyncedFields(ctx context.Context, state string, syncData map[int64]int64) error {
	if len(syncData) == 0 {
		return nil
	}

	fields := make([]string, 0, len(syncData))
	for id := range syncData {
		fields = append(fields, strconv.FormatInt(id, 10))
	}

	return p.cache.HDel(ctx, getKey(state), fields...)
}

func (p *provider) UpdateStatsHash(ctx context.Context, state string, id int64, incr int64) error {
	field := strconv.FormatInt(id, 10)
	_, err := p.cache.HIncrBy(ctx, getKey(state), field, incr)
	return err
}

func getKey(state string) string {
	return common.BuildCacheKey(SystemName, FeatureStats, state, "")
}
