package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/cache"
	cachemocks "air-social/internal/cache/mocks"
)

type tieredCacheSuite struct {
	suite.Suite
}

func TestTieredCacheSuite(t *testing.T) {
	suite.Run(t, new(tieredCacheSuite))
}

const (
	l1TTL = 5 * time.Minute
	l2TTL = 30 * time.Minute
)

func (s *tieredCacheSuite) TestGetOrLoad() {
	const key = "user:1"
	const val = "alice"

	type testDeps struct {
		l1     *cachemocks.MockCache[string]
		l2     *cachemocks.MockCache[string]
		loader func(context.Context) (string, error)
	}

	tests := []struct {
		name      string
		l1Nil     bool
		l2Nil     bool
		setupMock func(deps testDeps)
		want      string
		wantErr   error
	}{
		{
			name: "l1_hit",
			setupMock: func(deps testDeps) {
				deps.l1.EXPECT().
					Get(mock.Anything, key).
					Return(val, nil).
					Once()
			},
			want: val,
		},
		{
			name: "l1_miss_l2_hit_backfills_l1",
			setupMock: func(deps testDeps) {
				deps.l1.EXPECT().
					Get(mock.Anything, key).
					Return("", cache.ErrCacheMiss).
					Once()
				deps.l2.EXPECT().
					Get(mock.Anything, key).
					Return(val, nil).
					Once()
				deps.l1.EXPECT().
					Set(mock.Anything, key, val, l1TTL).
					Return(nil).
					Once()
			},
			want: val,
		},
		{
			name: "both_miss_loader_success_backfills_l2_l1",
			setupMock: func(deps testDeps) {
				deps.l1.EXPECT().
					Get(mock.Anything, key).
					Return("", cache.ErrCacheMiss).
					Once()
				deps.l2.EXPECT().
					Get(mock.Anything, key).
					Return("", cache.ErrCacheMiss).
					Once()
				deps.l2.EXPECT().
					Set(mock.Anything, key, val, l2TTL).
					Return(nil).
					Once()
				deps.l1.EXPECT().
					Set(mock.Anything, key, val, l1TTL).
					Return(nil).
					Once()
			},
			want: val,
		},
		{
			name: "both_miss_loader_error_no_backfill",
			setupMock: func(deps testDeps) {
				deps.l1.EXPECT().
					Get(mock.Anything, key).
					Return("", cache.ErrCacheMiss).
					Once()
				deps.l2.EXPECT().
					Get(mock.Anything, key).
					Return("", cache.ErrCacheMiss).
					Once()
			},
			want:    "",
			wantErr: context.DeadlineExceeded,
		},
		{
			name:  "l1_nil_l2_hit",
			l1Nil: true,
			setupMock: func(deps testDeps) {
				deps.l2.EXPECT().
					Get(mock.Anything, key).
					Return(val, nil).
					Once()
			},
			want: val,
		},
		{
			name:  "l2_nil_l1_miss_calls_loader",
			l2Nil: true,
			setupMock: func(deps testDeps) {
				deps.l1.EXPECT().
					Get(mock.Anything, key).
					Return("", cache.ErrCacheMiss).
					Once()
				deps.l1.EXPECT().
					Set(mock.Anything, key, val, l1TTL).
					Return(nil).
					Once()
			},
			want: val,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var (
				l1 cache.Cache[string]
				l2 cache.Cache[string]
			)

			deps := testDeps{}

			if !tc.l1Nil {
				m := cachemocks.NewMockCache[string](s.T())
				deps.l1 = m
				l1 = m
			}
			if !tc.l2Nil {
				m := cachemocks.NewMockCache[string](s.T())
				deps.l2 = m
				l2 = m
			}

			loaderErr := tc.wantErr
			deps.loader = func(_ context.Context) (string, error) {
				if loaderErr != nil {
					return "", loaderErr
				}
				return val, nil
			}

			tc.setupMock(deps)

			c := cache.NewTieredCache(l1, l2, l1TTL, l2TTL)
			got, err := c.GetOrLoad(context.Background(), key, deps.loader)

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

func (s *tieredCacheSuite) TestInvalidate() {
	const key = "user:1"

	type testDeps struct {
		l1 *cachemocks.MockCache[string]
		l2 *cachemocks.MockCache[string]
	}

	tests := []struct {
		name      string
		l2Nil     bool
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "deletes_from_both_l1_and_l2",
			setupMock: func(deps testDeps) {
				deps.l1.EXPECT().
					Delete(mock.Anything, key).
					Return(nil).
					Once()
				deps.l2.EXPECT().
					Delete(mock.Anything, key).
					Return(nil).
					Once()
			},
		},
		{
			name:  "l2_nil_only_l1_deleted",
			l2Nil: true,
			setupMock: func(deps testDeps) {
				deps.l1.EXPECT().
					Delete(mock.Anything, key).
					Return(nil).
					Once()
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var l2 cache.Cache[string]

			mockL1 := cachemocks.NewMockCache[string](s.T())
			deps := testDeps{l1: mockL1}

			if !tc.l2Nil {
				m := cachemocks.NewMockCache[string](s.T())
				deps.l2 = m
				l2 = m
			}

			tc.setupMock(deps)

			c := cache.NewTieredCache(mockL1, l2, l1TTL, l2TTL)
			err := c.Invalidate(context.Background(), key)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}
