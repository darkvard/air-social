package cache_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/common"
	commonmocks "air-social/internal/domain/common/mocks"
	"air-social/internal/domain/stats/cache"
)

type CacheProviderSuite struct {
	suite.Suite
}

func TestCacheProviderSuite(t *testing.T) {
	suite.Run(t, new(CacheProviderSuite))
}

func (s *CacheProviderSuite) TestGetStatsHash() {
	type testDeps struct {
		cache *commonmocks.MockCache
	}

	tests := []struct {
		name      string
		state     string
		setupMock func(deps testDeps)
		want      map[int64]int64
		wantErr   error
	}{
		{
			name:  "Success",
			state: cache.StatePostLikes,
			setupMock: func(deps testDeps) {
				key := common.BuildCacheKey(cache.SystemName, cache.FeatureStats, cache.StatePostLikes, "")
				mockData := map[string]string{
					"1": "10",
					"2": "20",
					"3": "0", // should be ignored
				}
				deps.cache.EXPECT().
					HGetAll(mock.Anything, key).
					Return(mockData, nil).
					Once()
			},
			want: map[int64]int64{
				1: 10,
				2: 20,
			},
			wantErr: nil,
		},
		{
			name:  "Success with empty result from cache",
			state: cache.StatePostLikes,
			setupMock: func(deps testDeps) {
				key := common.BuildCacheKey(cache.SystemName, cache.FeatureStats, cache.StatePostLikes, "")
				deps.cache.EXPECT().
					HGetAll(mock.Anything, key).
					Return(map[string]string{}, nil).
					Once()
			},
			want:    map[int64]int64{},
			wantErr: nil,
		},
		{
			name:  "Error from cache",
			state: cache.StatePostLikes,
			setupMock: func(deps testDeps) {
				key := common.BuildCacheKey(cache.SystemName, cache.FeatureStats, cache.StatePostLikes, "")
				deps.cache.EXPECT().
					HGetAll(mock.Anything, key).
					Return(nil, assert.AnError).
					Once()
			},
			want:    nil,
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockCache := commonmocks.NewMockCache(s.T())
			deps := testDeps{cache: mockCache}
			tt.setupMock(deps)

			provider := cache.NewProvider(mockCache)

			got, err := provider.GetStatsHash(context.Background(), tt.state)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
			}
			s.Equal(tt.want, got)
		})
	}
}

func (s *CacheProviderSuite) TestClearSyncedFields() {
	type testDeps struct {
		cache *commonmocks.MockCache
	}

	tests := []struct {
		name      string
		state     string
		syncData  map[int64]int64
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name:  "Success",
			state: cache.StatePostLikes,
			syncData: map[int64]int64{
				1: 10,
				2: 5,
			},
			setupMock: func(deps testDeps) {
				key := common.BuildCacheKey(cache.SystemName, cache.FeatureStats, cache.StatePostLikes, "")
				deps.cache.EXPECT().
					Eval(mock.Anything, mock.AnythingOfType("string"), []string{key}, []interface{}{"1", int64(-10)}).
					Return(nil, nil).
					Once()
				deps.cache.EXPECT().
					Eval(mock.Anything, mock.AnythingOfType("string"), []string{key}, []interface{}{"2", int64(-5)}).
					Return(nil, nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name:     "Success with no data to sync",
			state:    cache.StatePostLikes,
			syncData: map[int64]int64{},
			setupMock: func(deps testDeps) {
				// No calls to cache expected
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockCache := commonmocks.NewMockCache(s.T())
			deps := testDeps{cache: mockCache}
			tt.setupMock(deps)

			provider := cache.NewProvider(mockCache)

			err := provider.ClearSyncedFields(context.Background(), tt.state, tt.syncData)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *CacheProviderSuite) TestUpdateStatsHash() {
	type testDeps struct {
		cache *commonmocks.MockCache
	}

	type args struct {
		state string
		id    int64
		incr  int64
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps, a args)
		wantErr   error
	}{
		{
			name: "Success",
			args: args{state: cache.StatePostLikes, id: 1, incr: 1},
			setupMock: func(deps testDeps, a args) {
				key := common.BuildCacheKey(cache.SystemName, cache.FeatureStats, a.state, "")
				field := strconv.FormatInt(a.id, 10)
				deps.cache.EXPECT().
					HIncrBy(mock.Anything, key, field, a.incr).
					Return(int64(0), nil). // return value doesn't matter for this test
					Once()
			},
			wantErr: nil,
		},
		{
			name: "Error from cache",
			args: args{state: cache.StatePostLikes, id: 1, incr: 1},
			setupMock: func(deps testDeps, a args) {
				key := common.BuildCacheKey(cache.SystemName, cache.FeatureStats, a.state, "")
				field := strconv.FormatInt(a.id, 10)
				deps.cache.EXPECT().
					HIncrBy(mock.Anything, key, field, a.incr).
					Return(int64(0), assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockCache := commonmocks.NewMockCache(s.T())
			deps := testDeps{cache: mockCache}
			tt.setupMock(deps, tt.args)

			provider := cache.NewProvider(mockCache)

			err := provider.UpdateStatsHash(context.Background(), tt.args.state, tt.args.id, tt.args.incr)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *CacheProviderSuite) TestGetStatsOffsets() {
	type testDeps struct {
		cache *commonmocks.MockCache
	}

	type args struct {
		state string
		ids   []int64
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps, a args)
		want      map[int64]int64
		wantErr   error
	}{
		{
			name: "Success",
			args: args{state: cache.StatePostLikes, ids: []int64{1, 2, 3, 4}},
			setupMock: func(deps testDeps, a args) {
				key := common.BuildCacheKey(cache.SystemName, cache.FeatureStats, a.state, "")
				retVals := []string{"10", "-5", "", "not-an-int"}
				deps.cache.EXPECT().
					HMGet(mock.Anything, key, []string{"1", "2", "3", "4"}).
					Return(retVals, nil).
					Once()
			},
			want: map[int64]int64{
				1: 10,
				2: -5,
			},
			wantErr: nil,
		},
		{
			name: "Success with no ids",
			args: args{state: cache.StatePostLikes, ids: []int64{}},
			setupMock: func(deps testDeps, a args) {
				// No calls to cache
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "Error from cache",
			args: args{state: cache.StatePostLikes, ids: []int64{1}},
			setupMock: func(deps testDeps, a args) {
				key := common.BuildCacheKey(cache.SystemName, cache.FeatureStats, a.state, "")
				deps.cache.EXPECT().
					HMGet(mock.Anything, key, []string{"1"}).
					Return(nil, assert.AnError).
					Once()
			},
			want:    nil,
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockCache := commonmocks.NewMockCache(s.T())
			deps := testDeps{cache: mockCache}
			tt.setupMock(deps, tt.args)

			provider := cache.NewProvider(mockCache)

			got, err := provider.GetStatsOffsets(context.Background(), tt.args.state, tt.args.ids)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				if len(tt.args.ids) == 0 {
					s.Nil(got)
				} else {
					// Manually check map for zero values not present in want map
					for _, id := range tt.args.ids {
						if _, ok := tt.want[id]; !ok {
							s.Equal(int64(0), got[id], fmt.Sprintf("expected 0 for id %d", id))
						}
					}
					s.Equal(len(tt.want), len(got))
				}
			}
		})
	}
}
