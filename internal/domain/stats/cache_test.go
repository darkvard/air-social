package stats_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	cachemocks "air-social/internal/cache/mocks"
	"air-social/internal/domain/stats"
)

type statsCacheSuite struct {
	suite.Suite
}

func TestStatsCacheSuite(t *testing.T) {
	suite.Run(t, new(statsCacheSuite))
}

func (s *statsCacheSuite) TestGetStatsHash() {
	type testDeps struct {
		store *cachemocks.MockHashStore
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
			state: stats.StatePostLikes,
			setupMock: func(deps testDeps) {
				key := stats.StatsHashKey(stats.StatePostLikes)
				mockData := map[string]string{
					"1": "10",
					"2": "20",
					"3": "0", // should be ignored
				}
				deps.store.EXPECT().
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
			state: stats.StatePostLikes,
			setupMock: func(deps testDeps) {
				key := stats.StatsHashKey(stats.StatePostLikes)
				deps.store.EXPECT().
					HGetAll(mock.Anything, key).
					Return(map[string]string{}, nil).
					Once()
			},
			want:    map[int64]int64{},
			wantErr: nil,
		},
		{
			name:  "Error from cache",
			state: stats.StatePostLikes,
			setupMock: func(deps testDeps) {
				key := stats.StatsHashKey(stats.StatePostLikes)
				deps.store.EXPECT().
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
			mockStore := cachemocks.NewMockHashStore(s.T())
			deps := testDeps{store: mockStore}
			tt.setupMock(deps)

			c := stats.NewCache(mockStore)

			got, err := c.GetStatsHash(context.Background(), tt.state)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
			}
			s.Equal(tt.want, got)
		})
	}
}

func (s *statsCacheSuite) TestClearSyncedFields() {
	type testDeps struct {
		store *cachemocks.MockHashStore
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
			state: stats.StatePostLikes,
			syncData: map[int64]int64{
				1: 10,
				2: 5,
			},
			setupMock: func(deps testDeps) {
				key := stats.StatsHashKey(stats.StatePostLikes)
				deps.store.EXPECT().
					Eval(mock.Anything, mock.AnythingOfType("string"), []string{key}, "1", int64(-10)).
					Return(nil, nil).
					Maybe()
				deps.store.EXPECT().
					Eval(mock.Anything, mock.AnythingOfType("string"), []string{key}, "2", int64(-5)).
					Return(nil, nil).
					Maybe()
			},
			wantErr: nil,
		},
		{
			name:     "Success with no data to sync",
			state:    stats.StatePostLikes,
			syncData: map[int64]int64{},
			setupMock: func(deps testDeps) {
				// No calls to cache expected
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockStore := cachemocks.NewMockHashStore(s.T())
			deps := testDeps{store: mockStore}
			tt.setupMock(deps)

			c := stats.NewCache(mockStore)

			err := c.ClearSyncedFields(context.Background(), tt.state, tt.syncData)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *statsCacheSuite) TestUpdateStatsHash() {
	type testDeps struct {
		store *cachemocks.MockHashStore
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
			args: args{state: stats.StatePostLikes, id: 1, incr: 1},
			setupMock: func(deps testDeps, a args) {
				key := stats.StatsHashKey(a.state)
				field := strconv.FormatInt(a.id, 10)
				deps.store.EXPECT().
					HIncrBy(mock.Anything, key, field, a.incr).
					Return(int64(0), nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "Error from cache",
			args: args{state: stats.StatePostLikes, id: 1, incr: 1},
			setupMock: func(deps testDeps, a args) {
				key := stats.StatsHashKey(a.state)
				field := strconv.FormatInt(a.id, 10)
				deps.store.EXPECT().
					HIncrBy(mock.Anything, key, field, a.incr).
					Return(int64(0), assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockStore := cachemocks.NewMockHashStore(s.T())
			deps := testDeps{store: mockStore}
			tt.setupMock(deps, tt.args)

			c := stats.NewCache(mockStore)

			err := c.UpdateStatsHash(context.Background(), tt.args.state, tt.args.id, tt.args.incr)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *statsCacheSuite) TestGetStatsOffsets() {
	type testDeps struct {
		store *cachemocks.MockHashStore
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
			args: args{state: stats.StatePostLikes, ids: []int64{1, 2, 3, 4}},
			setupMock: func(deps testDeps, a args) {
				key := stats.StatsHashKey(a.state)
				retVals := []string{"10", "-5", "", "not-an-int"}
				deps.store.EXPECT().
					HMGet(mock.Anything, key, "1", "2", "3", "4").
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
			args: args{state: stats.StatePostLikes, ids: []int64{}},
			setupMock: func(deps testDeps, a args) {
				// No calls to cache
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "Error from cache",
			args: args{state: stats.StatePostLikes, ids: []int64{1}},
			setupMock: func(deps testDeps, a args) {
				key := stats.StatsHashKey(a.state)
				deps.store.EXPECT().
					HMGet(mock.Anything, key, "1").
					Return(nil, assert.AnError).
					Once()
			},
			want:    nil,
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockStore := cachemocks.NewMockHashStore(s.T())
			deps := testDeps{store: mockStore}
			tt.setupMock(deps, tt.args)

			c := stats.NewCache(mockStore)

			got, err := c.GetStatsOffsets(context.Background(), tt.args.state, tt.args.ids)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				if len(tt.args.ids) == 0 {
					s.Nil(got)
				} else {
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
