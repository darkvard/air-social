package message

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	presenceTTL    = 30 * time.Second
	presenceKeyFmt = "presence:online:%d"
)

type PresenceStore struct {
	client *redis.Client
}

func NewPresenceStorage(client *redis.Client) *PresenceStore {
	return &PresenceStore{client: client}
}

func (r *PresenceStore) SetOnline(ctx context.Context, userID int64) error {
	return r.client.Set(ctx, r.getKey(userID), "1", presenceTTL).Err()
}

func (r *PresenceStore) IsOnlineBatch(ctx context.Context, userIDs []int64) (map[int64]bool, error) {
	if len(userIDs) == 0 {
		return map[int64]bool{}, nil
	}
	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = r.getKey(id)
	}
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	result := make(map[int64]bool, len(userIDs))
	for i, v := range vals {
		result[userIDs[i]] = v != nil
	}
	return result, nil
}

func (r *PresenceStore) getKey(userID int64) string {
	return fmt.Sprintf(presenceKeyFmt, userID)
}
