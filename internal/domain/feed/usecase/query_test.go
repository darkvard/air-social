package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/common"
	feedcachemocks "air-social/internal/domain/feed/cache/mocks"
	"air-social/internal/domain/feed/usecase"
	feedmocks "air-social/internal/domain/feed/usecase/mocks"
	"air-social/internal/domain/post"
	"air-social/pkg"
)

type queryUseCaseSuite struct {
	suite.Suite
}

func TestQueryUseCaseSuite(t *testing.T) {
	suite.Run(t, new(queryUseCaseSuite))
}

func (s *queryUseCaseSuite) TestGetNewsfeed() {
	var (
		viewerID   = int64(1)
		cursor     = int64(0)
		limit      = 10
		fetchLimit = limit + 1 // GetFetchLimit()
	)

	postIDs := []int64{5, 3, 1}
	posts := []*post.Post{
		{ID: 1},
		{ID: 3},
		{ID: 5},
	}

	type testDeps struct {
		cache       *feedcachemocks.MockProvider
		postFetcher *feedmocks.MockPostFetcher
	}

	type args struct {
		ctx      context.Context
		viewerID int64
		params   common.CursorQueryParams[int64]
	}

	tests := []struct {
		name           string
		args           args
		setupMock      func(deps testDeps)
		wantLen        int
		wantOrder      []int64
		wantHasNext    bool
		wantNextCursor int64
		wantErr        error
	}{
		{
			name: "cache_error",
			args: args{
				ctx:      context.Background(),
				viewerID: viewerID,
				params:   common.CursorQueryParams[int64]{Cursor: cursor, Limit: limit},
			},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetFeedPostIDs(mock.Anything, viewerID, cursor, fetchLimit).
					Return(nil, assert.AnError).
					Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "empty_feed",
			args: args{
				ctx:      context.Background(),
				viewerID: viewerID,
				params:   common.CursorQueryParams[int64]{Cursor: cursor, Limit: limit},
			},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetFeedPostIDs(mock.Anything, viewerID, cursor, fetchLimit).
					Return([]int64{}, nil).
					Once()
			},
			wantLen: 0,
			wantErr: nil,
		},
		{
			name: "post_fetcher_error",
			args: args{
				ctx:      context.Background(),
				viewerID: viewerID,
				params:   common.CursorQueryParams[int64]{Cursor: cursor, Limit: limit},
			},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetFeedPostIDs(mock.Anything, viewerID, cursor, fetchLimit).
					Return(postIDs, nil).
					Once()
				deps.postFetcher.EXPECT().
					GetPostsByIDs(mock.Anything, postIDs, viewerID).
					Return(nil, assert.AnError).
					Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success_order_preserved",
			args: args{
				ctx:      context.Background(),
				viewerID: viewerID,
				params:   common.CursorQueryParams[int64]{Cursor: cursor, Limit: limit},
			},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetFeedPostIDs(mock.Anything, viewerID, cursor, fetchLimit).
					Return(postIDs, nil).
					Once()
				// DB returns posts in arbitrary order
				deps.postFetcher.EXPECT().
					GetPostsByIDs(mock.Anything, postIDs, viewerID).
					Return(posts, nil).
					Once()
			},
			wantLen:   3,
			wantOrder: postIDs, // must match cache order: 5, 3, 1
			wantErr:   nil,
		},
		{
			name: "success_some_posts_deleted",
			args: args{
				ctx:      context.Background(),
				viewerID: viewerID,
				params:   common.CursorQueryParams[int64]{Cursor: cursor, Limit: limit},
			},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetFeedPostIDs(mock.Anything, viewerID, cursor, fetchLimit).
					Return(postIDs, nil).
					Once()
				// DB only returns 2 of the 3 posts (one was deleted)
				deps.postFetcher.EXPECT().
					GetPostsByIDs(mock.Anything, postIDs, viewerID).
					Return([]*post.Post{{ID: 5}, {ID: 1}}, nil).
					Once()
			},
			wantLen:   2,
			wantOrder: []int64{5, 1},
			wantErr:   nil,
		},
		{
			// Redis returns limit+1 items → hasNextPage=true, result trimmed to limit.
			// next_cursor = CreatedAt.UnixMilli() of the last post in the trimmed page.
			name: "has_next_page",
			args: args{
				ctx:      context.Background(),
				viewerID: viewerID,
				params:   common.CursorQueryParams[int64]{Cursor: cursor, Limit: 2},
			},
			setupMock: func(deps testDeps) {
				// fetchLimit = 2+1 = 3; cache returns 3 IDs → signals next page exists
				deps.cache.EXPECT().
					GetFeedPostIDs(mock.Anything, viewerID, cursor, 3).
					Return([]int64{5, 3, 1}, nil).
					Once()
				// only the first 2 IDs are queried (trimmed before DB call)
				deps.postFetcher.EXPECT().
					GetPostsByIDs(mock.Anything, []int64{5, 3}, viewerID).
					Return([]*post.Post{
						{ID: 5, CreatedAt: time.UnixMilli(1700000000000)},
						{ID: 3, CreatedAt: time.UnixMilli(1699999990000)},
					}, nil).
					Once()
			},
			wantLen:        2,
			wantOrder:      []int64{5, 3},
			wantHasNext:    true,
			wantNextCursor: 1699999990000, // CreatedAt.UnixMilli() of post ID=3 (last in page)
			wantErr:        nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := feedcachemocks.NewMockProvider(s.T())
			mockPostFetcher := feedmocks.NewMockPostFetcher(s.T())

			deps := testDeps{
				cache:       mockCache,
				postFetcher: mockPostFetcher,
			}

			uc := usecase.NewQueryUseCase(usecase.QueryDeps{
				CacheProvider: mockCache,
				PostFetcher:   mockPostFetcher,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetNewsfeed(tc.args.ctx, tc.args.viewerID, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
				s.Len(got.Data, tc.wantLen)
				for i, id := range tc.wantOrder {
					s.Equal(id, got.Data[i].ID)
				}
				s.Equal(tc.wantHasNext, got.HasNextPage)
				if tc.wantHasNext {
					s.Equal(tc.wantNextCursor, got.NextCursor)
				}
			}
		})
	}
}
