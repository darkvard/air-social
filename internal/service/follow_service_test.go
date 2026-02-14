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
		setupMocks func(followRepo *mocks.FollowRepository, userSvc *mocks.UserService, cache *mocks.CacheStorage, a args)
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
			setupMocks: func(followRepo *mocks.FollowRepository, userSvc *mocks.UserService, cache *mocks.CacheStorage, a args) {
				userSvc.
					EXPECT().
					GetSummary(mock.Anything, a.followeeID).
					Return(nil, assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "get user not found",
			args: validPayload,
			setupMocks: func(followRepo *mocks.FollowRepository, userSvc *mocks.UserService, cache *mocks.CacheStorage, a args) {
				userSvc.
					EXPECT().
					GetSummary(mock.Anything, a.followeeID).
					Return(nil, nil).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "create follow error",
			args: validPayload,
			setupMocks: func(followRepo *mocks.FollowRepository, userSvc *mocks.UserService, cache *mocks.CacheStorage, a args) {
				userSvc.
					EXPECT().
					GetSummary(mock.Anything, a.followeeID).
					Return(&domain.UserSummary{}, nil).
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
			setupMocks: func(followRepo *mocks.FollowRepository, userSvc *mocks.UserService, cache *mocks.CacheStorage, a args) {
				userSvc.
					EXPECT().
					GetSummary(mock.Anything, a.followeeID).
					Return(&domain.UserSummary{}, nil).
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
			mockUserSvc := mocks.NewUserService(s.T())
			mockFollowRepo := mocks.NewFollowRepository(s.T())
			mockCache := mocks.NewCacheStorage(s.T())

			followSvc := NewFollowService(mockFollowRepo, mockCache, mockUserSvc)

			if tc.setupMocks != nil {
				tc.setupMocks(mockFollowRepo, mockUserSvc, mockCache, tc.args)
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
			mockCache := mocks.NewCacheStorage(s.T())

			followSvc := NewFollowService(mockFollowRepo, mockCache, nil)

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
	var (
		targetID      int64 = 1
		currentUserID int64 = 99
		params              = domain.FollowParams{
			TargetUserID:  targetID,
			CurrentUserID: currentUserID,
			QueryParams:   domain.QueryParams{Page: 1, Limit: 10},
		}
		users = []domain.User{
			{ID: 2, Username: "following1"},
			{ID: 3, Username: "following2"},
		}
		total int64 = 5
	)

	type want struct {
		result domain.PaginatedResult[domain.SocialUser]
		err    error
	}

	tests := []struct {
		name       string
		args       domain.FollowParams
		setupMocks func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams)
		want       want
	}{
		{
			name: "success_with_enrichment",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.
					EXPECT().
					GetFollowings(mock.Anything, input).
					Return(users, nil).
					Once()

				cache.
					EXPECT().
					Get(mock.Anything, domain.GetFollowingCountKey(input.TargetUserID), mock.Anything).
					Run(func(_ context.Context, _ string, dest any) {
						*(dest.(*int64)) = total
					}).
					Return(nil).
					Once()

				targetIDs := []int64{2, 3}

				followRepo.
					EXPECT().
					IsFollowing(mock.Anything, input.CurrentUserID, targetIDs).
					Return(map[int64]bool{2: true}, nil).
					Once()

				followRepo.
					EXPECT().
					IsFollowedBy(mock.Anything, input.CurrentUserID, targetIDs).
					Return(map[int64]bool{3: true}, nil).
					Once()
			},
			want: want{
				result: domain.NewPaginatedResult([]domain.SocialUser{
					{User: users[0], Relation: domain.Relationship{IsFollowing: true, IsFollowedBy: false}},
					{User: users[1], Relation: domain.Relationship{IsFollowing: false, IsFollowedBy: true}},
				}, total, params.Page, params.Limit),
				err: nil,
			},
		},
		{
			name: "success_guest_mode_cache_miss",
			args: domain.FollowParams{
				TargetUserID:  targetID,
				CurrentUserID: 0,
				QueryParams:   domain.QueryParams{Page: 1, Limit: 10},
			},
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.
					EXPECT().
					GetFollowings(mock.Anything, input).
					Return(users, nil).
					Once()

				cache.
					EXPECT().
					Get(mock.Anything, domain.GetFollowingCountKey(input.TargetUserID), mock.Anything).
					Return(assert.AnError).
					Once()

				followRepo.
					EXPECT().
					CountFollowings(mock.Anything, input.TargetUserID).
					Return(total, nil).
					Once()

				cache.
					EXPECT().
					Set(mock.Anything, domain.GetFollowingCountKey(input.TargetUserID), total, time.Hour).
					Return(nil).
					Maybe()
			},
			want: want{
				result: domain.NewPaginatedResult([]domain.SocialUser{
					{User: users[0]}, {User: users[1]},
				}, total, params.Page, params.Limit),
				err: nil,
			},
		},
		{
			name: "fetch_list_error",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.
					EXPECT().
					GetFollowings(mock.Anything, input).
					Return(nil, assert.AnError).
					Once()

				cache.
					EXPECT().
					Get(mock.Anything, domain.GetFollowingCountKey(input.TargetUserID), mock.Anything).
					Return(assert.AnError).
					Maybe()

				followRepo.
					EXPECT().
					CountFollowings(mock.Anything, input.TargetUserID).Return(int64(0), nil).
					Maybe()

				cache.
					EXPECT().
					Set(mock.Anything, domain.GetFollowingCountKey(input.TargetUserID), int64(0), time.Hour).
					Return(nil).
					Maybe()
			},
			want: want{
				result: domain.PaginatedResult[domain.SocialUser]{},
				err:    assert.AnError,
			},
		},
		{
			name: "fetch_count_error",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.
					EXPECT().
					GetFollowings(mock.Anything, input).
					Return(users, nil).
					Once()

				cache.
					EXPECT().
					Get(mock.Anything, domain.GetFollowingCountKey(input.TargetUserID), mock.Anything).
					Return(assert.AnError).
					Once()

				followRepo.
					EXPECT().
					CountFollowings(mock.Anything, input.TargetUserID).
					Return(int64(0), assert.AnError).
					Once()

				followRepo.
					EXPECT().
					IsFollowing(mock.Anything, input.CurrentUserID, mock.Anything).
					Return(nil, nil).
					Maybe()

				followRepo.
					EXPECT().
					IsFollowedBy(mock.Anything, input.CurrentUserID, mock.Anything).
					Return(nil, nil).
					Maybe()
			},
			want: want{
				result: domain.PaginatedResult[domain.SocialUser]{},
				err:    assert.AnError,
			},
		},
		{
			name: "enrichment_error",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.
					EXPECT().
					GetFollowings(mock.Anything, input).
					Return(users, nil).
					Once()

				cache.
					EXPECT().
					Get(mock.Anything, domain.GetFollowingCountKey(input.TargetUserID), mock.Anything).
					Run(func(_ context.Context, _ string, dest interface{}) {
						*(dest.(*int64)) = total
					}).
					Return(nil).
					Maybe()

				targetIDs := []int64{2, 3}
				followRepo.
					EXPECT().
					IsFollowing(mock.Anything, input.CurrentUserID, targetIDs).
					Return(nil, assert.AnError).
					Once()
			},
			want: want{
				result: domain.PaginatedResult[domain.SocialUser]{},
				err:    assert.AnError,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockFollowRepo := mocks.NewFollowRepository(s.T())
			mockCache := mocks.NewCacheStorage(s.T())

			followSvc := NewFollowService(mockFollowRepo, mockCache, nil)

			if tc.setupMocks != nil {
				tc.setupMocks(mockFollowRepo, mockCache, tc.args)
			}

			got, err := followSvc.GetFollowings(context.Background(), tc.args)

			if tc.want.err != nil {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tc.want.result, got)
			}
		})
	}
}

func (s *followServiceSuite) TestGetFollowers() {
	var (
		targetID      int64 = 1
		currentUserID int64 = 99
		params              = domain.FollowParams{
			TargetUserID:  targetID,
			CurrentUserID: currentUserID,
			QueryParams:   domain.QueryParams{Page: 1, Limit: 10},
		}
		users = []domain.User{
			{ID: 2, Username: "follower1"},
			{ID: 3, Username: "follower2"},
		}
		total int64 = 5
	)

	type want struct {
		result domain.PaginatedResult[domain.SocialUser]
		err    error
	}

	tests := []struct {
		name       string
		args       domain.FollowParams
		setupMocks func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams)
		want       want
	}{
		{
			name: "success_with_enrichment",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.
					EXPECT().
					GetFollowers(mock.Anything, input).
					Return(users, nil).
					Once()

				cache.
					EXPECT().
					Get(mock.Anything, domain.GetFollowerCountKey(targetID), mock.Anything).
					Run(func(_ context.Context, _ string, dest interface{}) {
						*(dest.(*int64)) = total
					}).
					Return(nil).
					Once()

				targetIDs := []int64{2, 3}

				followRepo.
					EXPECT().
					IsFollowing(mock.Anything, input.CurrentUserID, targetIDs).
					Return(map[int64]bool{2: true}, nil).
					Once()

				followRepo.
					EXPECT().
					IsFollowedBy(mock.Anything, input.CurrentUserID, targetIDs).
					Return(map[int64]bool{3: true}, nil).
					Once()
			},
			want: want{
				result: domain.NewPaginatedResult([]domain.SocialUser{
					{User: users[0], Relation: domain.Relationship{IsFollowing: true, IsFollowedBy: false}},
					{User: users[1], Relation: domain.Relationship{IsFollowing: false, IsFollowedBy: true}},
				}, total, params.Page, params.Limit),
				err: nil,
			},
		},
		{
			name: "success_guest_mode_cache_miss",
			args: domain.FollowParams{
				TargetUserID:  targetID,
				CurrentUserID: 0,
				QueryParams:   domain.QueryParams{Page: 1, Limit: 10},
			},
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.
					EXPECT().
					GetFollowers(mock.Anything, input).
					Return(users, nil).
					Once()

				cache.
					EXPECT().
					Get(mock.Anything, domain.GetFollowerCountKey(targetID), mock.Anything).
					Return(assert.AnError).
					Once()

				followRepo.
					EXPECT().
					CountFollowers(mock.Anything, targetID).
					Return(total, nil).
					Once()

				cache.
					EXPECT().
					Set(mock.Anything, domain.GetFollowerCountKey(targetID), total, time.Hour).
					Return(nil).
					Maybe()
			},
			want: want{
				result: domain.NewPaginatedResult([]domain.SocialUser{
					{User: users[0]}, {User: users[1]},
				}, total, params.Page, params.Limit),
				err: nil,
			},
		},
		{
			name: "fetch_list_error",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.
					EXPECT().
					GetFollowers(mock.Anything, input).
					Return(nil, assert.AnError).
					Once()

				cache.
					EXPECT().
					Get(mock.Anything, domain.GetFollowerCountKey(targetID), mock.Anything).
					Return(assert.AnError).
					Maybe()

				followRepo.
					EXPECT().
					CountFollowers(mock.Anything, targetID).
					Return(int64(0), nil).
					Maybe()

				cache.
					EXPECT().
					Set(mock.Anything, domain.GetFollowerCountKey(targetID), int64(0), time.Hour).
					Return(nil).
					Maybe()
			},
			want: want{
				result: domain.PaginatedResult[domain.SocialUser]{},
				err:    assert.AnError,
			},
		},
		{
			name: "fetch_count_error",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.
					EXPECT().
					GetFollowers(mock.Anything, input).
					Return(users, nil).
					Once()

				cache.
					EXPECT().
					Get(mock.Anything, domain.GetFollowerCountKey(targetID), mock.Anything).
					Return(assert.AnError).
					Once()

				followRepo.
					EXPECT().
					CountFollowers(mock.Anything, targetID).
					Return(int64(0), assert.AnError).
					Once()

				followRepo.
					EXPECT().
					IsFollowing(mock.Anything, input.CurrentUserID, mock.Anything).
					Return(nil, nil).
					Maybe()

				followRepo.
					EXPECT().
					IsFollowedBy(mock.Anything, input.CurrentUserID, mock.Anything).
					Return(nil, nil).
					Maybe()
			},
			want: want{
				result: domain.PaginatedResult[domain.SocialUser]{},
				err:    assert.AnError,
			},
		},
		{
			name: "enrichment_error",
			args: params,
			setupMocks: func(followRepo *mocks.FollowRepository, cache *mocks.CacheStorage, input domain.FollowParams) {
				followRepo.
					EXPECT().
					GetFollowers(mock.Anything, input).
					Return(users, nil).
					Once()

				cache.
					EXPECT().
					Get(mock.Anything, domain.GetFollowerCountKey(targetID), mock.Anything).
					Run(func(_ context.Context, _ string, dest interface{}) {
						*(dest.(*int64)) = total
					}).Return(nil).
					Maybe()

				targetIDs := []int64{2, 3}

				followRepo.
					EXPECT().
					IsFollowing(mock.Anything, input.CurrentUserID, targetIDs).
					Return(nil, assert.AnError).
					Once()
			},
			want: want{
				result: domain.PaginatedResult[domain.SocialUser]{},
				err:    assert.AnError,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockFollowRepo := mocks.NewFollowRepository(s.T())
			mockCache := mocks.NewCacheStorage(s.T())

			followSvc := NewFollowService(mockFollowRepo, mockCache, nil)

			if tc.setupMocks != nil {
				tc.setupMocks(mockFollowRepo, mockCache, tc.args)
			}

			got, err := followSvc.GetFollowers(context.Background(), tc.args)

			if tc.want.err != nil {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tc.want.result, got)
			}
		})
	}
}
