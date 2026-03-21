package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"air-social/internal/cache"
)

type memCacheSuite struct {
	suite.Suite
}

func TestMemCacheSuite(t *testing.T) {
	suite.Run(t, new(memCacheSuite))
}

func (s *memCacheSuite) TestGet() {
	type args struct {
		key string
	}

	tests := []struct {
		name    string
		setup   func(c *cache.MemCache[string])
		args    args
		want    string
		wantErr error
	}{
		{
			name: "hit",
			setup: func(c *cache.MemCache[string]) {
				_ = c.Set(context.Background(), "k1", "v1", 0)
			},
			args:    args{key: "k1"},
			want:    "v1",
			wantErr: nil,
		},
		{
			name:    "miss_key_not_found",
			setup:   nil,
			args:    args{key: "missing"},
			want:    "",
			wantErr: cache.ErrCacheMiss,
		},
		{
			name: "miss_expired",
			setup: func(c *cache.MemCache[string]) {
				_ = c.Set(context.Background(), "k2", "v2", time.Millisecond)
				time.Sleep(5 * time.Millisecond)
			},
			args:    args{key: "k2"},
			want:    "",
			wantErr: cache.ErrCacheMiss,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			c := cache.NewMemCache[string](0, time.Minute)
			if tc.setup != nil {
				tc.setup(c)
			}

			got, err := c.Get(context.Background(), tc.args.key)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want, got)
			}
		})
	}
}

func (s *memCacheSuite) TestSet() {
	tests := []struct {
		name      string
		maxSize   int
		setupKeys []string // keys to pre-fill before the actual Set
		key       string
		val       string
		ttl       time.Duration
		// after Set, try Get; expectHit=false means the entry was dropped
		expectHit bool
	}{
		{
			name:      "normal_set",
			maxSize:   0,
			key:       "k",
			val:       "v",
			ttl:       time.Minute,
			expectHit: true,
		},
		{
			name:      "no_expiry_ttl_zero",
			maxSize:   0,
			key:       "k",
			val:       "v",
			ttl:       0,
			expectHit: true,
		},
		{
			name:      "maxsize_full_drops_entry",
			maxSize:   2,
			setupKeys: []string{"a", "b"},
			key:       "c",
			val:       "v",
			ttl:       time.Minute,
			expectHit: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			c := cache.NewMemCache[string](tc.maxSize, time.Minute)
			ctx := context.Background()

			for _, k := range tc.setupKeys {
				_ = c.Set(ctx, k, "prefill", time.Minute)
			}

			err := c.Set(ctx, tc.key, tc.val, tc.ttl)
			s.NoError(err)

			got, getErr := c.Get(ctx, tc.key)
			if tc.expectHit {
				s.NoError(getErr)
				s.Equal(tc.val, got)
			} else {
				s.ErrorIs(getErr, cache.ErrCacheMiss)
			}
		})
	}
}

func (s *memCacheSuite) TestDelete() {
	tests := []struct {
		name  string
		setup func(c *cache.MemCache[string])
		key   string
	}{
		{
			name: "delete_existing_key",
			setup: func(c *cache.MemCache[string]) {
				_ = c.Set(context.Background(), "k", "v", time.Minute)
			},
			key: "k",
		},
		{
			name:  "delete_nonexistent_key_no_error",
			setup: nil,
			key:   "ghost",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			c := cache.NewMemCache[string](0, time.Minute)
			if tc.setup != nil {
				tc.setup(c)
			}

			err := c.Delete(context.Background(), tc.key)
			s.NoError(err)

			_, getErr := c.Get(context.Background(), tc.key)
			s.ErrorIs(getErr, cache.ErrCacheMiss)
		})
	}
}
