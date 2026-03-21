package cache

import (
	"context"
	"fmt"
	"strconv"

	appcache "air-social/internal/cache"
	"air-social/internal/domain/common"
)

const (
	MaxFeedSizePerUser = 500
	FeedKeyPrefix      = "feed"
)

type Provider interface {
	PushPostToFeeds(ctx context.Context, followerIDs []int64, postID int64, score float64) error
	RemovePostFromFeeds(ctx context.Context, followerIDs []int64, postID int64) error
	GetFeedPostIDs(ctx context.Context, userID int64, cursor int64, limit int) ([]int64, error)
}

type provider struct {
	cache appcache.SortedSetStore
}

func NewProvider(c appcache.SortedSetStore) *provider {
	return &provider{cache: c}
}

// PushPostToFeeds distributes a newly created post to the feeds of all followers.
// It uses a Redis Pipeline to execute ZADD and ZREMRANGEBYRANK atomically.
func (p *provider) PushPostToFeeds(ctx context.Context, followerIDs []int64, postID int64, score float64) error {
	if len(followerIDs) == 0 {
		return nil
	}
	batch := p.cache.Pipeline()
	for _, id := range followerIDs {
		key := getFeedKey(id)

		// Add the new post to the sorted set. Score is the timestamp.
		// Higher score means newer post (it will bubble up to the top).
		batch.ZAdd(key, score, postID)

		// Trim the feed to maintain a maximum size of MaxFeedSizePerUser (e.g., 500).
		// In Redis, rank 0 is the oldest element, -1 is the newest, -2 is the second newest, etc.
		// ZRemRangeByRank(key, 0, -501) means:
		// "Delete everything from the oldest element (0) up to the 501st newest element (-501)".
		// This effectively leaves exactly the 500 newest elements intact (-1 to -500).
		batch.ZRemRangeByRank(key, 0, -(MaxFeedSizePerUser + 1))
	}
	return batch.Exec(ctx)
}

func (p *provider) RemovePostFromFeeds(ctx context.Context, followerIDs []int64, postID int64) error {
	if len(followerIDs) == 0 {
		return nil
	}
	batch := p.cache.Pipeline()
	for _, id := range followerIDs {
		batch.ZRem(getFeedKey(id), postID)
	}
	return batch.Exec(ctx)
}

// GetFeedPostIDs retrieves a paginated list of post IDs from a user's feed.
// The feed is stored as a Redis Sorted Set where:
// - member = postID
// - score = timestamp (higher = newer)
//
// Pagination is implemented using a cursor (timestamp).
// - If cursor == 0: fetch latest posts
// - If cursor > 0: fetch posts with score strictly less than cursor
//
// Redis ZRevRangeByScore is used to:
// - return results in descending order (newest → oldest)
// - filter by score range
func (p *provider) GetFeedPostIDs(ctx context.Context, userID int64, cursor int64, limit int) ([]int64, error) {
	key := getFeedKey(userID)

	// Default: fetch from newest posts (no upper bound)
	maxScore := "+inf"

	// If cursor is provided, fetch posts strictly older than the cursor
	// "(" makes the upper bound exclusive
	if cursor > 0 {
		maxScore = fmt.Sprintf("(%d", cursor)
	}

	// Fetch posts with score in range [-inf, maxScore]
	// Ordered from highest score → lowest score (newest → oldest)
	strs, err := p.cache.ZRevRangeByScore(ctx, key, "-inf", maxScore, 0, int64(limit))
	if err != nil {
		return nil, err
	}

	// Convert string IDs to int64
	ids := make([]int64, 0, len(strs))
	for _, s := range strs {
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

func getFeedKey(userID int64) string {
	return common.BuildCacheKey("feed", "newsfeed", userID)
}
