package stats_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/stats"
	"air-social/internal/domain/stats/cache"
	cachemocks "air-social/internal/domain/stats/cache/mocks"
	statsmocks "air-social/internal/domain/stats/mocks"
	"air-social/pkg"
)

type StatsUseCaseSuite struct {
	suite.Suite
}

func TestStatsUseCaseSuite(t *testing.T) {
	suite.Run(t, new(StatsUseCaseSuite))
}

func (s *StatsUseCaseSuite) TestSyncPostStats() {
	var (
		postID1 int64 = 10
		postID2 int64 = 20
	)

	likesMap := map[int64]int64{postID1: 5, postID2: 0}
	commentsMap := map[int64]int64{postID1: 2, postID2: 1}
	sharesMap := map[int64]int64{postID1: 0, postID2: 3}

	type testDeps struct {
		repo  *statsmocks.MockRepository
		cache *cachemocks.MockProvider
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostLikes).
					Return(likesMap, nil).
					Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostComments).
					Return(commentsMap, nil).
					Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostShares).
					Return(sharesMap, nil).
					Once()

				deps.repo.EXPECT().
					BulkUpsertPostStats(mock.Anything, mock.MatchedBy(func(p stats.PostParams) bool {
						if len(p.IDs) != 2 {
							return false
						}
						for i, id := range p.IDs {
							switch id {
							case postID1:
								if p.Likes[i] != 5 || p.Comments[i] != 2 || p.Shares[i] != 0 {
									return false
								}
							case postID2:
								if p.Likes[i] != 0 || p.Comments[i] != 1 || p.Shares[i] != 3 {
									return false
								}
							default:
								return false
							}
						}
						return true
					})).
					Return(nil).
					Once()

				deps.cache.EXPECT().
					ClearSyncedFields(mock.Anything, cache.StatePostLikes, likesMap).
					Return(nil).
					Once()
				deps.cache.EXPECT().
					ClearSyncedFields(mock.Anything, cache.StatePostComments, commentsMap).
					Return(nil).
					Once()
				deps.cache.EXPECT().
					ClearSyncedFields(mock.Anything, cache.StatePostShares, sharesMap).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "cache_fetch_error",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostLikes).
					Return(nil, assert.AnError).
					Once()

				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostComments).
					Return(map[int64]int64{}, nil).
					Maybe()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostShares).
					Return(map[int64]int64{}, nil).
					Maybe()
			},
			wantErr: assert.AnError,
		},
		{
			name: "empty_stats",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostLikes).
					Return(map[int64]int64{}, nil).
					Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostComments).
					Return(map[int64]int64{}, nil).
					Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostShares).
					Return(map[int64]int64{}, nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "repo_error",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostLikes).
					Return(likesMap, nil).
					Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostComments).
					Return(commentsMap, nil).
					Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StatePostShares).
					Return(sharesMap, nil).
					Once()

				deps.repo.EXPECT().
					BulkUpsertPostStats(mock.Anything, mock.Anything).
					Return(assert.AnError).
					Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := statsmocks.NewMockRepository(s.T())
			mockCache := cachemocks.NewMockProvider(s.T())

			deps := testDeps{
				repo:  mockRepo,
				cache: mockCache,
			}

			uc := stats.NewUseCase(stats.Deps{
				Repo:  mockRepo,
				Cache: mockCache,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.SyncPostStats(context.Background())

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *StatsUseCaseSuite) TestSyncCommentStats() {
	var (
		commentID1 int64 = 100
		commentID2 int64 = 200
	)

	likesMap := map[int64]int64{commentID1: 10, commentID2: 0}
	repliesMap := map[int64]int64{commentID1: 5, commentID2: 2}

	type testDeps struct {
		repo  *statsmocks.MockRepository
		cache *cachemocks.MockProvider
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StateCommentLikes).
					Return(likesMap, nil).
					Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StateCommentReplies).
					Return(repliesMap, nil).
					Once()

				deps.repo.EXPECT().
					BulkUpsertCommentStats(mock.Anything, mock.MatchedBy(func(p stats.CommentParams) bool {
						if len(p.IDs) != 2 {
							return false
						}
						for i, id := range p.IDs {
							switch id {
							case commentID1:
								if p.Likes[i] != 10 || p.Replies[i] != 5 {
									return false
								}
							case commentID2:
								if p.Likes[i] != 0 || p.Replies[i] != 2 {
									return false
								}
							default:
								return false
							}
						}
						return true
					})).
					Return(nil).
					Once()

				deps.cache.EXPECT().
					ClearSyncedFields(mock.Anything, cache.StateCommentLikes, likesMap).
					Return(nil).
					Once()
				deps.cache.EXPECT().
					ClearSyncedFields(mock.Anything, cache.StateCommentReplies, repliesMap).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "cache_fetch_error",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StateCommentLikes).
					Return(nil, assert.AnError).
					Once()

				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StateCommentReplies).
					Return(map[int64]int64{}, nil).
					Maybe()
			},
			wantErr: assert.AnError,
		},
		{
			name: "empty_stats",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StateCommentLikes).
					Return(map[int64]int64{}, nil).
					Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StateCommentReplies).
					Return(map[int64]int64{}, nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "repo_error",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StateCommentLikes).
					Return(likesMap, nil).
					Once()
				deps.cache.EXPECT().
					GetStatsHash(mock.Anything, cache.StateCommentReplies).
					Return(repliesMap, nil).
					Once()

				deps.repo.EXPECT().
					BulkUpsertCommentStats(mock.Anything, mock.Anything).
					Return(assert.AnError).
					Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := statsmocks.NewMockRepository(s.T())
			mockCache := cachemocks.NewMockProvider(s.T())

			deps := testDeps{
				repo:  mockRepo,
				cache: mockCache,
			}

			uc := stats.NewUseCase(stats.Deps{
				Repo:  mockRepo,
				Cache: mockCache,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.SyncCommentStats(context.Background())

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}
