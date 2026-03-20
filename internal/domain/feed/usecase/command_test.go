package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	feedcachemocks "air-social/internal/domain/feed/cache/mocks"
	"air-social/internal/domain/feed/usecase"
	feedmocks "air-social/internal/domain/feed/usecase/mocks"
)

type commandUseCaseSuite struct {
	suite.Suite
}

func TestCommandUseCaseSuite(t *testing.T) {
	suite.Run(t, new(commandUseCaseSuite))
}

func (s *commandUseCaseSuite) TestDistributePost() {
	var (
		postID    = int64(10)
		authorID  = int64(1)
		timestamp = int64(1700000000)
	)

	followerIDs := []int64{2, 3, 4}
	expectedFeedIDs := append(followerIDs, authorID)

	type testDeps struct {
		cache         *feedcachemocks.MockProvider
		followFetcher *feedmocks.MockFollowFetcher
	}

	type args struct {
		ctx       context.Context
		postID    int64
		authorID  int64
		timestamp int64
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "follow_fetcher_error",
			args: args{ctx: context.Background(), postID: postID, authorID: authorID, timestamp: timestamp},
			setupMock: func(deps testDeps) {
				deps.followFetcher.EXPECT().
					GetFollowerIDs(mock.Anything, authorID).
					Return(nil, assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "cache_push_error",
			args: args{ctx: context.Background(), postID: postID, authorID: authorID, timestamp: timestamp},
			setupMock: func(deps testDeps) {
				deps.followFetcher.EXPECT().
					GetFollowerIDs(mock.Anything, authorID).
					Return(followerIDs, nil).
					Once()
				deps.cache.EXPECT().
					PushPostToFeeds(mock.Anything, expectedFeedIDs, postID, float64(timestamp)).
					Return(assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: args{ctx: context.Background(), postID: postID, authorID: authorID, timestamp: timestamp},
			setupMock: func(deps testDeps) {
				deps.followFetcher.EXPECT().
					GetFollowerIDs(mock.Anything, authorID).
					Return(followerIDs, nil).
					Once()
				deps.cache.EXPECT().
					PushPostToFeeds(mock.Anything, expectedFeedIDs, postID, float64(timestamp)).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "success_no_followers_author_only",
			args: args{ctx: context.Background(), postID: postID, authorID: authorID, timestamp: timestamp},
			setupMock: func(deps testDeps) {
				deps.followFetcher.EXPECT().
					GetFollowerIDs(mock.Anything, authorID).
					Return([]int64{}, nil).
					Once()
				deps.cache.EXPECT().
					PushPostToFeeds(mock.Anything, []int64{authorID}, postID, float64(timestamp)).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := feedcachemocks.NewMockProvider(s.T())
			mockFollow := feedmocks.NewMockFollowFetcher(s.T())

			deps := testDeps{
				cache:         mockCache,
				followFetcher: mockFollow,
			}

			uc := usecase.NewCommandUseCase(usecase.CommandDeps{
				CacheProvider: mockCache,
				FollowFetcher: mockFollow,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.DistributePost(tc.args.ctx, tc.args.postID, tc.args.authorID, tc.args.timestamp)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *commandUseCaseSuite) TestRevokePost() {
	var (
		postID   = int64(10)
		authorID = int64(1)
	)

	followerIDs := []int64{2, 3, 4}
	expectedFeedIDs := append(followerIDs, authorID)

	type testDeps struct {
		cache         *feedcachemocks.MockProvider
		followFetcher *feedmocks.MockFollowFetcher
	}

	type args struct {
		ctx      context.Context
		postID   int64
		authorID int64
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "follow_fetcher_error",
			args: args{ctx: context.Background(), postID: postID, authorID: authorID},
			setupMock: func(deps testDeps) {
				deps.followFetcher.EXPECT().
					GetFollowerIDs(mock.Anything, authorID).
					Return(nil, assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "cache_remove_error",
			args: args{ctx: context.Background(), postID: postID, authorID: authorID},
			setupMock: func(deps testDeps) {
				deps.followFetcher.EXPECT().
					GetFollowerIDs(mock.Anything, authorID).
					Return(followerIDs, nil).
					Once()
				deps.cache.EXPECT().
					RemovePostFromFeeds(mock.Anything, expectedFeedIDs, postID).
					Return(assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: args{ctx: context.Background(), postID: postID, authorID: authorID},
			setupMock: func(deps testDeps) {
				deps.followFetcher.EXPECT().
					GetFollowerIDs(mock.Anything, authorID).
					Return(followerIDs, nil).
					Once()
				deps.cache.EXPECT().
					RemovePostFromFeeds(mock.Anything, expectedFeedIDs, postID).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := feedcachemocks.NewMockProvider(s.T())
			mockFollow := feedmocks.NewMockFollowFetcher(s.T())

			deps := testDeps{
				cache:         mockCache,
				followFetcher: mockFollow,
			}

			uc := usecase.NewCommandUseCase(usecase.CommandDeps{
				CacheProvider: mockCache,
				FollowFetcher: mockFollow,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.RevokePost(tc.args.ctx, tc.args.postID, tc.args.authorID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}
