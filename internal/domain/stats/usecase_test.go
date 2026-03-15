package stats_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/stats"
	cachemocks "air-social/internal/domain/stats/cache/mocks"
	statmocks "air-social/internal/domain/stats/mocks"
	"air-social/pkg"
)

type StatsUseCaseSuite struct {
	suite.Suite
}

func TestStatsUseCaseSuite(t *testing.T) {
	suite.Run(t, new(StatsUseCaseSuite))
}

func (s *StatsUseCaseSuite) TestSyncPostStats() {
	type testDeps struct {
		repo  *statmocks.MockRepository
		cache *cachemocks.MockProvider
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "Success",
			setupMock: func(deps testDeps) {
				likes := map[int64]int64{1: 10, 2: 5}
				comments := map[int64]int64{1: 5, 3: 1}
				shares := map[int64]int64{2: 1}

				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, "post_likes").
					Return(likes, nil).Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, "post_comments").
					Return(comments, nil).Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, "post_shares").
					Return(shares, nil).Once()

				deps.repo.EXPECT().
					BulkUpsertPostStats(mock.Anything, mock.MatchedBy(func(p stats.PostParams) bool {
						if len(p.IDs) != 3 {
							return false
						}
						foundIDs := make(map[int64]bool)
						// Verify data integrity regardless of order
						for i, id := range p.IDs {
							foundIDs[id] = true
							if id == 1 && (p.Likes[i] != 10 || p.Comments[i] != 5 || p.Shares[i] != 0) {
								return false
							}
							if id == 2 && (p.Likes[i] != 5 || p.Comments[i] != 0 || p.Shares[i] != 1) {
								return false
							}
							if id == 3 && (p.Likes[i] != 0 || p.Comments[i] != 1 || p.Shares[i] != 0) {
								return false
							}
						}
						return len(foundIDs) == 3
					})).
					Return(nil).Once()

				deps.cache.EXPECT().
					ClearSyncedFields(mock.Anything, "post_likes", likes).
					Return(nil).Once()
				deps.cache.EXPECT().
					ClearSyncedFields(mock.Anything, "post_comments", comments).
					Return(nil).Once()
				deps.cache.EXPECT().
					ClearSyncedFields(mock.Anything, "post_shares", shares).
					Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "Success with no stats to sync",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, mock.Anything).
					Return(map[int64]int64{}, nil).Times(3)
			},
			wantErr: nil,
		},
		{
			name: "Error fetching from cache",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, "post_likes").
					Return(nil, assert.AnError).Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, "post_comments").
					Return(map[int64]int64{}, nil).Maybe()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, "post_shares").
					Return(map[int64]int64{}, nil).Maybe()
			},
			wantErr: assert.AnError,
		},
		{
			name: "Error on bulk upsert",
			setupMock: func(deps testDeps) {
				likes := map[int64]int64{1: 10}
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, mock.Anything).
					Return(likes, nil).Times(3)

				deps.repo.EXPECT().
					BulkUpsertPostStats(mock.Anything, mock.AnythingOfType("stats.PostParams")).
					Return(assert.AnError).Once()
				deps.repo.EXPECT().
					ReconcilePostStats(mock.Anything, mock.Anything).Return(nil).Maybe()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := statmocks.NewMockRepository(s.T())
			mockCache := cachemocks.NewMockProvider(s.T())
			deps := testDeps{repo: mockRepo, cache: mockCache}
			tt.setupMock(deps)

			uc := stats.NewUseCase(stats.Deps{
				Repo:  mockRepo,
				Cache: mockCache,
			})

			err := uc.SyncPostStats(context.Background())

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
				if tt.wantErr == assert.AnError {
					s.Error(err) // Check strict error existence for generic errors
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *StatsUseCaseSuite) TestSyncCommentStats() {
	type testDeps struct {
		repo  *statmocks.MockRepository
		cache *cachemocks.MockProvider
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "Success",
			setupMock: func(deps testDeps) {
				likes := map[int64]int64{1: 10, 2: 5}
				replies := map[int64]int64{1: 5, 3: 1}

				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, "comment_likes").
					Return(likes, nil).Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, "comment_replies").
					Return(replies, nil).Once()

				deps.repo.EXPECT().
					BulkUpsertCommentStats(mock.Anything, mock.MatchedBy(func(p stats.CommentParams) bool {
						if len(p.IDs) != 3 {
							return false
						}
						for i, id := range p.IDs {
							if id == 1 && (p.Likes[i] != 10 || p.Replies[i] != 5) {
								return false
							}
							if id == 2 && (p.Likes[i] != 5 || p.Replies[i] != 0) {
								return false
							}
							if id == 3 && (p.Likes[i] != 0 || p.Replies[i] != 1) {
								return false
							}
						}
						return true
					})).
					Return(nil).Once()

				deps.cache.EXPECT().
					ClearSyncedFields(mock.Anything, "comment_likes", likes).
					Return(nil).Once()
				deps.cache.EXPECT().
					ClearSyncedFields(mock.Anything, "comment_replies", replies).
					Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "Error fetching from cache",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, "comment_likes").
					Return(nil, assert.AnError).Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, "comment_replies").
					Return(map[int64]int64{}, nil).Maybe()
			},
			wantErr: assert.AnError,
		},
		{
			name: "Error on bulk upsert",
			setupMock: func(deps testDeps) {
				likes := map[int64]int64{1: 10}
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, mock.Anything).
					Return(likes, nil).Times(2)

				deps.repo.EXPECT().
					BulkUpsertCommentStats(mock.Anything, mock.AnythingOfType("stats.CommentParams")).
					Return(assert.AnError).Once()
				deps.repo.EXPECT().
					ReconcileCommentStats(mock.Anything, mock.Anything).Return(nil).Maybe()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := statmocks.NewMockRepository(s.T())
			mockCache := cachemocks.NewMockProvider(s.T())
			deps := testDeps{repo: mockRepo, cache: mockCache}
			tt.setupMock(deps)

			uc := stats.NewUseCase(stats.Deps{
				Repo:  mockRepo,
				Cache: mockCache,
			})

			err := uc.SyncCommentStats(context.Background())

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
				if tt.wantErr == assert.AnError {
					s.Error(err)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *StatsUseCaseSuite) TestReconcilePostStats() {
	type testDeps struct {
		repo  *statmocks.MockRepository
		cache *cachemocks.MockProvider
	}

	tests := []struct {
		name      string
		args      []int64
		setupMock func(deps testDeps, ids []int64)
		wantErr   error
	}{
		{
			name: "Success",
			args: []int64{1, 2},
			setupMock: func(deps testDeps, ids []int64) {
				deps.repo.EXPECT().ReconcilePostStats(mock.Anything, ids).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "Error",
			args: []int64{1, 2},
			setupMock: func(deps testDeps, ids []int64) {
				deps.repo.EXPECT().ReconcilePostStats(mock.Anything, ids).Return(assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "Empty IDs",
			args: []int64{},
			setupMock: func(deps testDeps, ids []int64) {
				// Should return early, no repo call
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := statmocks.NewMockRepository(s.T())
			mockCache := cachemocks.NewMockProvider(s.T())
			deps := testDeps{repo: mockRepo, cache: mockCache}
			tt.setupMock(deps, tt.args)

			uc := stats.NewUseCase(stats.Deps{
				Repo:  mockRepo,
				Cache: mockCache,
			})

			err := uc.ReconcilePostStats(context.Background(), tt.args)

			if tt.wantErr != nil {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *StatsUseCaseSuite) TestReconcileCommentStats() {
	type testDeps struct {
		repo  *statmocks.MockRepository
		cache *cachemocks.MockProvider
	}

	tests := []struct {
		name      string
		args      []int64
		setupMock func(deps testDeps, ids []int64)
		wantErr   error
	}{
		{
			name: "Success",
			args: []int64{1, 2},
			setupMock: func(deps testDeps, ids []int64) {
				deps.repo.EXPECT().ReconcileCommentStats(mock.Anything, ids).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "Error",
			args: []int64{1, 2},
			setupMock: func(deps testDeps, ids []int64) {
				deps.repo.EXPECT().ReconcileCommentStats(mock.Anything, ids).Return(assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "Empty IDs",
			args: []int64{},
			setupMock: func(deps testDeps, ids []int64) {
				// Should return early, no repo call
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := statmocks.NewMockRepository(s.T())
			mockCache := cachemocks.NewMockProvider(s.T())
			deps := testDeps{repo: mockRepo, cache: mockCache}
			tt.setupMock(deps, tt.args)

			uc := stats.NewUseCase(stats.Deps{
				Repo:  mockRepo,
				Cache: mockCache,
			})

			err := uc.ReconcileCommentStats(context.Background(), tt.args)

			if tt.wantErr != nil {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *StatsUseCaseSuite) TestGetPostsStats() {
	type testDeps struct {
		repo  *statmocks.MockRepository
		cache *cachemocks.MockProvider
	}

	postIDs := []int64{1, 2, 3}

	tests := []struct {
		name      string
		args      []int64
		setupMock func(deps testDeps)
		want      map[int64]stats.PostStats
		wantErr   error
	}{
		{
			name: "Success merging DB and cache",
			args: postIDs,
			setupMock: func(deps testDeps) {
				dbStats := []stats.PostStats{
					{PostID: 1, LikesCount: 100, CommentsCount: 50, SharesCount: 10},
					{PostID: 2, LikesCount: 200, CommentsCount: 0, SharesCount: 0},
				}
				deps.repo.EXPECT().
					GetPostsStats(mock.Anything, postIDs).
					Return(dbStats, nil).Once()

				likesOffset := map[int64]int64{1: 5, 2: -10}  // Post 2 has a decrement
				commentsOffset := map[int64]int64{1: 2, 3: 5} // Post 3 is new, only in cache
				sharesOffset := map[int64]int64{1: 1}
				deps.cache.EXPECT().
					GetStatsOffsets(mock.Anything, "post_likes", postIDs).
					Return(likesOffset, nil).Once()
				deps.cache.EXPECT().
					GetStatsOffsets(mock.Anything, "post_comments", postIDs).
					Return(commentsOffset, nil).Once()
				deps.cache.EXPECT().
					GetStatsOffsets(mock.Anything, "post_shares", postIDs).
					Return(sharesOffset, nil).Once()
			},
			want: map[int64]stats.PostStats{
				1: {PostID: 1, LikesCount: 105, CommentsCount: 52, SharesCount: 11},
				2: {PostID: 2, LikesCount: 190, CommentsCount: 0, SharesCount: 0},
				3: {PostID: 3, LikesCount: 0, CommentsCount: 5, SharesCount: 0},
			},
			wantErr: nil,
		},
		{
			name: "Success with negative offset resulting in zero",
			args: []int64{1},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetPostsStats(mock.Anything, []int64{1}).
					Return([]stats.PostStats{{PostID: 1, LikesCount: 5}}, nil).Once()

				deps.cache.EXPECT().
					GetStatsOffsets(mock.Anything, mock.Anything, []int64{1}).
					Return(map[int64]int64{1: -10}, nil).Times(3)
			},
			want: map[int64]stats.PostStats{
				1: {PostID: 1, LikesCount: 0, CommentsCount: 0, SharesCount: 0},
			},
			wantErr: nil,
		},
		{
			name:      "Success with empty IDs",
			args:      []int64{},
			setupMock: func(deps testDeps) {},
			want:      nil,
			wantErr:   nil,
		},
		{
			name: "Error from DB",
			args: postIDs,
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetPostsStats(mock.Anything, postIDs).
					Return(nil, assert.AnError).Once()
				deps.cache.EXPECT().
					GetStatsOffsets(mock.Anything, mock.Anything, postIDs).
					Return(map[int64]int64{}, nil).Maybe()
			},
			want:    nil,
			wantErr: assert.AnError,
		},
		{
			name: "Error from Cache",
			args: postIDs,
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetPostsStats(mock.Anything, postIDs).
					Return([]stats.PostStats{}, nil).Once()
				deps.cache.EXPECT().
					GetStatsOffsets(mock.Anything, "post_likes", postIDs).
					Return(nil, assert.AnError).Once()
				deps.cache.EXPECT().
					GetStatsOffsets(mock.Anything, mock.Anything, postIDs).
					Return(map[int64]int64{}, nil).Maybe()
			},
			want:    nil,
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := statmocks.NewMockRepository(s.T())
			mockCache := cachemocks.NewMockProvider(s.T())
			deps := testDeps{repo: mockRepo, cache: mockCache}
			tt.setupMock(deps)

			uc := stats.NewUseCase(stats.Deps{
				Repo:  mockRepo,
				Cache: mockCache,
			})

			got, err := uc.GetPostsStats(context.Background(), tt.args)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
				if tt.wantErr == assert.AnError {
					s.Error(err)
				}
			} else {
				s.NoError(err)
				s.Equal(tt.want, got)
			}
		})
	}
}

func (s *StatsUseCaseSuite) TestGetCommentsStats() {
	type testDeps struct {
		repo  *statmocks.MockRepository
		cache *cachemocks.MockProvider
	}

	commentIDs := []int64{1, 2, 3}

	tests := []struct {
		name      string
		args      []int64
		setupMock func(deps testDeps)
		want      map[int64]stats.CommentStats
		wantErr   error
	}{
		{
			name: "Success merging DB and cache",
			args: commentIDs,
			setupMock: func(deps testDeps) {
				dbStats := []stats.CommentStats{
					{CommentID: 1, LikesCount: 100, RepliesCount: 50},
					{CommentID: 2, LikesCount: 200, RepliesCount: 0},
				}
				deps.repo.EXPECT().
					GetCommentsStats(mock.Anything, commentIDs).
					Return(dbStats, nil).Once()

				likesOffset := map[int64]int64{1: 5, 2: -10}
				repliesOffset := map[int64]int64{1: 2, 3: 5}
				deps.cache.EXPECT().
					GetStatsOffsets(mock.Anything, "comment_likes", commentIDs).
					Return(likesOffset, nil).Once()
				deps.cache.EXPECT().
					GetStatsOffsets(mock.Anything, "comment_replies", commentIDs).
					Return(repliesOffset, nil).Once()
			},
			want: map[int64]stats.CommentStats{
				1: {CommentID: 1, LikesCount: 105, RepliesCount: 52},
				2: {CommentID: 2, LikesCount: 190, RepliesCount: 0},
				3: {CommentID: 3, LikesCount: 0, RepliesCount: 5},
			},
			wantErr: nil,
		},
		{
			name:      "Success with empty IDs",
			args:      []int64{},
			setupMock: func(deps testDeps) {},
			want:      nil,
			wantErr:   nil,
		},
		{
			name: "Error from DB",
			args: commentIDs,
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetCommentsStats(mock.Anything, commentIDs).
					Return(nil, assert.AnError).Once()
				deps.cache.EXPECT().
					GetStatsOffsets(mock.Anything, mock.Anything, commentIDs).
					Return(map[int64]int64{}, nil).Maybe()
			},
			want:    nil,
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := statmocks.NewMockRepository(s.T())
			mockCache := cachemocks.NewMockProvider(s.T())
			deps := testDeps{repo: mockRepo, cache: mockCache}
			tt.setupMock(deps)

			uc := stats.NewUseCase(stats.Deps{
				Repo:  mockRepo,
				Cache: mockCache,
			})

			got, err := uc.GetCommentsStats(context.Background(), tt.args)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
				if tt.wantErr == assert.AnError {
					s.Error(err)
				}
			} else {
				s.NoError(err)
				s.Equal(tt.want, got)
			}
		})
	}
}
