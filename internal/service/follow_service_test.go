package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain"
	"air-social/internal/mocks"
)

type followServiceSuite struct {
	suite.Suite
}

func TestFollowServiceSuite(t *testing.T) {
	suite.Run(t, new(followServiceSuite))
}

func (s *followServiceSuite) TestFollow() {
	type args struct {
		followerID int64
		followeeID int64
	}

	invalidPayload := args{
		followerID: 1,
		followeeID: 1,
	}

	validPayload := args{
		followerID: 1,
		followeeID: 2,
	}

	tests := []struct {
		name       string
		args       args
		setupMocks func(followRepo *mocks.FollowRepository, userRepo *mocks.UserRepository, cache *mocks.CacheStorage, a args)
		wantErr    error
	}{
		{
			name:    "same id",
			args:    invalidPayload,
			wantErr: assert.AnError,
		},
		{
			name: "get user error",
			args: validPayload,
			setupMocks: func(followRepo *mocks.FollowRepository, userRepo *mocks.UserRepository, cache *mocks.CacheStorage, a args) {
				userRepo.
					EXPECT().
					GetByID(mock.Anything, a.followeeID).
					Return(nil, assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "get user not found",
			args: validPayload,
			setupMocks: func(followRepo *mocks.FollowRepository, userRepo *mocks.UserRepository, cache *mocks.CacheStorage, a args) {
				userRepo.
					EXPECT().
					GetByID(mock.Anything, a.followeeID).
					Return(nil, nil).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "create follow error",
			args: validPayload,
			setupMocks: func(followRepo *mocks.FollowRepository, userRepo *mocks.UserRepository, cache *mocks.CacheStorage, a args) {
				userRepo.
					EXPECT().
					GetByID(mock.Anything, a.followeeID).
					Return(&domain.User{}, nil).
					Once()

				followRepo.
					EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "create follow success",
			args: validPayload,
			setupMocks: func(followRepo *mocks.FollowRepository, userRepo *mocks.UserRepository, cache *mocks.CacheStorage, a args) {
				userRepo.
					EXPECT().
					GetByID(mock.Anything, a.followeeID).
					Return(&domain.User{}, nil).
					Once()
				followRepo.
					EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil).
					Once()
				cache.
					EXPECT().
					Delete(mock.Anything, mock.Anything).
					Return(nil).
					Maybe()

			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockUserRepo := mocks.NewUserRepository(s.T())
			mockFollowRepo := mocks.NewFollowRepository(s.T())
			mockCache := mocks.NewCacheStorage(s.T())

			followSvc := NewFollowService(mockFollowRepo, mockUserRepo, mockCache)

			if tc.setupMocks != nil {
				tc.setupMocks(mockFollowRepo, mockUserRepo, mockCache, tc.args)
			}

			err := followSvc.Follow(context.Background(), tc.args.followerID, tc.args.followeeID)
			if tc.wantErr != nil {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}

}

func (s *followServiceSuite) TestUnfollow() {
	type args struct {
		followerID int64
		followeeID int64
	}

	invalidPayload := args{
		followerID: 1,
		followeeID: 1,
	}

	validPayload := args{
		followerID: 1,
		followeeID: 2,
	}

	tests := []struct {
		name       string
		args       args
		setupMocks func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, a args)
		wantErr    error
	}{
		{
			name:       "same id",
			args:       invalidPayload,
			setupMocks: nil,
			wantErr:    assert.AnError,
		},
		{
			name: "delete error",
			args: validPayload,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, a args) {
				followRepo.
					EXPECT().
					Delete(mock.Anything, a.followerID, a.followeeID).
					Return(assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "unfollow success",
			args: validPayload,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, a args) {
				followRepo.
					EXPECT().
					Delete(mock.Anything, a.followerID, a.followeeID).
					Return(nil).
					Once()
				cache.
					EXPECT().
					Delete(mock.Anything, mock.Anything).
					Return(nil).
					Maybe()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockFollowRepo := mocks.NewFollowRepository(s.T())
			mockUserRepo := mocks.NewUserRepository(s.T())
			mockCache := mocks.NewCacheStorage(s.T())

			followSvc := NewFollowService(mockFollowRepo, mockUserRepo, mockCache)

			if tc.setupMocks != nil {
				tc.setupMocks(mockFollowRepo, mockCache, tc.args)
			}

			err := followSvc.Unfollow(context.Background(), tc.args.followerID, tc.args.followeeID)

			if tc.wantErr != nil {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *followServiceSuite) TestGetFollowings() {

}

func (s *followServiceSuite) TestGetFollowers() {
	var (
		userID int64 = 1
		params       = domain.FollowParams{
			UserID: userID,
			Page:   1,
			Limit:  10,
		}
		users = []domain.User{
			{ID: 2, Username: "follower1"},
			{ID: 3, Username: "follower2"},
		}
		total int64 = 5
	)

	type want struct {
		result domain.FollowResult
		err    error
	}

	tests := []struct {
		name       string
		args       domain.FollowParams
		setupMocks func(
			followRepo *mocks.FollowRepository,
			cache *mocks.CacheStorage,
			input domain.FollowParams,
		)
		want want
	}{
		{
			name: "success_cache_hit",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.EXPECT().
					GetFollowers(mock.Anything, input).
					Return(users, nil).
					Once()

				cache.EXPECT().
					Get(
						mock.Anything,
						domain.GetFollowerCountKey(input.UserID),
						mock.Anything,
					).
					Run(func(_ context.Context, _ string, dest interface{}) {
						*(dest.(*int64)) = total
					}).
					Return(nil).
					Once()
			},
			want: want{
				result: domain.FollowResult{
					Users: users,
					Total: total,
					Page:  params.Page,
					Limit: params.Limit,
				},
				err: nil,
			},
		},
		{
			name: "success_cache_miss",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.EXPECT().
					GetFollowers(mock.Anything, input).
					Return(users, nil).
					Once()

				cache.EXPECT().
					Get(
						mock.Anything,
						domain.GetFollowerCountKey(input.UserID),
						mock.Anything,
					).
					Return(assert.AnError).
					Once()

				followRepo.EXPECT().
					CountFollowers(mock.Anything, input.UserID).
					Return(total, nil).
					Once()

				cache.EXPECT().
					Set(
						mock.Anything,
						domain.GetFollowerCountKey(input.UserID),
						total,
						time.Hour,
					).
					Return(nil).
					Once()
			},
			want: want{
				result: domain.FollowResult{
					Users: users,
					Total: total,
					Page:  params.Page,
					Limit: params.Limit,
				},
				err: nil,
			},
		},
		{
			name: "get_followers_error",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.EXPECT().
					GetFollowers(mock.Anything, input).
					Return(nil, assert.AnError).
					Once()

				cache.EXPECT().
					Get(
						mock.Anything,
						domain.GetFollowerCountKey(input.UserID),
						mock.Anything,
					).
					Return(assert.AnError).
					Maybe()

				followRepo.EXPECT().
					CountFollowers(mock.Anything, input.UserID).
					Return(int64(0), nil).
					Maybe()

				cache.EXPECT().
					Set(
						mock.Anything,
						domain.GetFollowerCountKey(input.UserID),
						mock.Anything,
						time.Hour,
					).
					Return(nil).
					Maybe()

			},
			want: want{
				result: domain.FollowResult{},
				err:    assert.AnError,
			},
		},
		{
			name: "count_followers_error",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.EXPECT().
					GetFollowers(mock.Anything, input).
					Return(users, nil).
					Once()

				cache.EXPECT().
					Get(
						mock.Anything,
						domain.GetFollowerCountKey(input.UserID),
						mock.Anything,
					).
					Return(assert.AnError).
					Once()

				followRepo.EXPECT().
					CountFollowers(mock.Anything, input.UserID).
					Return(int64(0), assert.AnError).
					Once()
			},
			want: want{
				result: domain.FollowResult{},
				err:    assert.AnError,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockFollowRepo := mocks.NewFollowRepository(s.T())
			mockCache := mocks.NewCacheStorage(s.T())

			followSvc := NewFollowService(mockFollowRepo, nil, mockCache)

			if tc.setupMocks != nil {
				tc.setupMocks(mockFollowRepo, mockCache, tc.args)
			}

			got, err := followSvc.GetFollowers(context.Background(), tc.args)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Empty(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.result, got)
			}

			mockFollowRepo.AssertExpectations(s.T())
			mockCache.AssertExpectations(s.T())
		})
	}
}
