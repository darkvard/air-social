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

func (r *PresenceStore) SetOffline(ctx context.Context, userID int64) error {
	return r.client.Del(ctx, r.getKey(userID)).Err()
}

func (r *PresenceStore) RefreshPresence(ctx context.Context, userID int64) error {
	_, err := r.client.Expire(ctx, r.getKey(userID), presenceTTL).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *PresenceStore) IsOnline(ctx context.Context, userID int64) (bool, error) {
	n, err := r.client.Exists(ctx, r.getKey(userID)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *PresenceStore) getKey(userID int64) string {
	return fmt.Sprintf(presenceKeyFmt, userID)
}
