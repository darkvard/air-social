package follow_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/common"
	"air-social/internal/domain/follow"
	"air-social/internal/domain/follow/mocks"
	"air-social/internal/domain/user"
	"air-social/pkg"
)

type followUseCaseSuite struct {
	suite.Suite
}

func TestFollowUseCaseSuite(t *testing.T) {
	suite.Run(t, new(followUseCaseSuite))
}

func (s *followUseCaseSuite) TestFollow() {
	var (
		followerID int64 = 1
		followeeID int64 = 2
	)

	type testDeps struct {
		repo    *mocks.MockRepository
		fetcher *mocks.MockUserFetcher
	}

	tests := []struct {
		name string
		args struct {
			ctx        context.Context
			followerID int64
			followeeID int64
		}
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "self_follow",
			args: struct {
				ctx        context.Context
				followerID int64
				followeeID int64
			}{context.Background(), followerID, followerID},
			setupMock: func(deps testDeps) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "target_user_not_found",
			args: struct {
				ctx        context.Context
				followerID int64
				followeeID int64
			}{context.Background(), followerID, followeeID},
			setupMock: func(deps testDeps) {
				deps.fetcher.EXPECT().
					GetSummary(mock.Anything, followeeID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "repo_error",
			args: struct {
				ctx        context.Context
				followerID int64
				followeeID int64
			}{context.Background(), followerID, followeeID},
			setupMock: func(deps testDeps) {
				deps.fetcher.EXPECT().
					GetSummary(mock.Anything, followeeID).
					Return(&user.UserSummary{ID: followeeID}, nil).Once()

				deps.repo.EXPECT().
					Create(mock.Anything, followerID, followeeID).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			args: struct {
				ctx        context.Context
				followerID int64
				followeeID int64
			}{context.Background(), followerID, followeeID},
			setupMock: func(deps testDeps) {
				deps.fetcher.EXPECT().
					GetSummary(mock.Anything, followeeID).
					Return(&user.UserSummary{ID: followeeID}, nil).Once()

				deps.repo.EXPECT().
					Create(mock.Anything, followerID, followeeID).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewMockRepository(s.T())
			mockFetcher := mocks.NewMockUserFetcher(s.T())

			deps := testDeps{
				repo:    mockRepo,
				fetcher: mockFetcher,
			}
			uc := follow.NewUseCase(follow.Deps{
				FollowRepo:  mockRepo,
				UserFetcher: mockFetcher,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.Follow(tc.args.ctx, tc.args.followerID, tc.args.followeeID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *followUseCaseSuite) TestUnfollow() {
	var (
		followerID int64 = 1
		followeeID int64 = 2
	)

	type testDeps struct {
		repo    *mocks.MockRepository
		fetcher *mocks.MockUserFetcher
	}

	tests := []struct {
		name string
		args struct {
			ctx        context.Context
			followerID int64
			followeeID int64
		}
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "self_unfollow",
			args: struct {
				ctx        context.Context
				followerID int64
				followeeID int64
			}{context.Background(), followerID, followerID},
			setupMock: func(deps testDeps) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "target_user_not_found",
			args: struct {
				ctx        context.Context
				followerID int64
				followeeID int64
			}{context.Background(), followerID, followeeID},
			setupMock: func(deps testDeps) {
				deps.fetcher.EXPECT().
					GetSummary(mock.Anything, followeeID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "repo_error",
			args: struct {
				ctx        context.Context
				followerID int64
				followeeID int64
			}{context.Background(), followerID, followeeID},
			setupMock: func(deps testDeps) {
				deps.fetcher.EXPECT().
					GetSummary(mock.Anything, followeeID).
					Return(&user.UserSummary{ID: followeeID}, nil).Once()

				deps.repo.EXPECT().
					Delete(mock.Anything, followerID, followeeID).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			args: struct {
				ctx        context.Context
				followerID int64
				followeeID int64
			}{context.Background(), followerID, followeeID},
			setupMock: func(deps testDeps) {
				deps.fetcher.EXPECT().
					GetSummary(mock.Anything, followeeID).
					Return(&user.UserSummary{ID: followeeID}, nil).Once()

				deps.repo.EXPECT().
					Delete(mock.Anything, followerID, followeeID).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewMockRepository(s.T())
			mockFetcher := mocks.NewMockUserFetcher(s.T())

			deps := testDeps{
				repo:    mockRepo,
				fetcher: mockFetcher,
			}
			uc := follow.NewUseCase(follow.Deps{
				FollowRepo:  mockRepo,
				UserFetcher: mockFetcher,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.Unfollow(tc.args.ctx, tc.args.followerID, tc.args.followeeID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *followUseCaseSuite) TestGetFollowings() {
	var (
		userID int64 = 1
	)
	params := follow.GetFollowsParams{
		ViewerID: userID,
		Paging:   common.OffsetQueryParams{Page: 1, Limit: 10},
	}

	mockResult := []follow.FollowUser{
		{ID: 2, FullName: "User 2"},
	}
	mockTotal := int64(1)

	type testDeps struct {
		repo    *mocks.MockRepository
		fetcher *mocks.MockUserFetcher
	}

	tests := []struct {
		name string
		args struct {
			ctx    context.Context
			params follow.GetFollowsParams
		}
		setupMock func(deps testDeps)
		wantLen   int
		wantErr   error
	}{
		{
			name: "repo_error",
			args: struct {
				ctx    context.Context
				params follow.GetFollowsParams
			}{context.Background(), params},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetFollowings(mock.Anything, params).
					Return(nil, int64(0), assert.AnError).Once()
			},
			wantLen: 0,
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: struct {
				ctx    context.Context
				params follow.GetFollowsParams
			}{context.Background(), params},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetFollowings(mock.Anything, params).
					Return(mockResult, mockTotal, nil).Once()
			},
			wantLen: 1,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewMockRepository(s.T())
			mockFetcher := mocks.NewMockUserFetcher(s.T())

			deps := testDeps{
				repo:    mockRepo,
				fetcher: mockFetcher,
			}
			uc := follow.NewUseCase(follow.Deps{
				FollowRepo:  mockRepo,
				UserFetcher: mockFetcher,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetFollowings(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
				s.Equal(tc.wantLen, len(got.Data))
				s.Equal(mockTotal, got.Total)
			}
		})
	}
}

func (s *followUseCaseSuite) TestGetFollowers() {
	var (
		userID int64 = 1
	)
	params := follow.GetFollowsParams{
		ViewerID: userID,
		Paging:   common.OffsetQueryParams{Page: 1, Limit: 10},
	}

	mockResult := []follow.FollowUser{
		{ID: 2, FullName: "User 2"},
	}
	mockTotal := int64(1)

	type testDeps struct {
		repo    *mocks.MockRepository
		fetcher *mocks.MockUserFetcher
	}

	tests := []struct {
		name string
		args struct {
			ctx    context.Context
			params follow.GetFollowsParams
		}
		setupMock func(deps testDeps)
		wantLen   int
		wantErr   error
	}{
		{
			name: "repo_error",
			args: struct {
				ctx    context.Context
				params follow.GetFollowsParams
			}{context.Background(), params},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetFollowers(mock.Anything, params).
					Return(nil, int64(0), assert.AnError).Once()
			},
			wantLen: 0,
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: struct {
				ctx    context.Context
				params follow.GetFollowsParams
			}{context.Background(), params},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetFollowers(mock.Anything, params).
					Return(mockResult, mockTotal, nil).Once()
			},
			wantLen: 1,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewMockRepository(s.T())
			mockFetcher := mocks.NewMockUserFetcher(s.T())

			deps := testDeps{
				repo:    mockRepo,
				fetcher: mockFetcher,
			}
			uc := follow.NewUseCase(follow.Deps{
				FollowRepo:  mockRepo,
				UserFetcher: mockFetcher,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetFollowers(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
				s.Equal(tc.wantLen, len(got.Data))
				s.Equal(mockTotal, got.Total)
			}
		})
	}
}
