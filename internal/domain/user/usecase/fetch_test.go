package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appcache "air-social/internal/cache"
	cachemocks "air-social/internal/cache/mocks"
	commonmocks "air-social/internal/domain/common/mocks"
	"air-social/internal/domain/user"
	usermocks "air-social/internal/domain/user/mocks"
	"air-social/internal/domain/user/usecase"
	"air-social/pkg"
)

type fetchUseCaseSuite struct {
	suite.Suite
}

func TestFetchUseCaseSuite(t *testing.T) {
	suite.Run(t, new(fetchUseCaseSuite))
}

func (s *fetchUseCaseSuite) TestGetByID() {
	var (
		userID int64 = 1
	)

	expectedUser := &user.User{
		ID:    userID,
		Email: "test@example.com",
	}

	type testDeps struct {
		repo  *usermocks.MockRepository
		cache *cachemocks.MockCache[*user.UserSummary]
		link  *commonmocks.MockLinkProvider
	}

	type args struct {
		ctx context.Context
		id  int64
	}

	type want struct {
		user *user.User
		err  error
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      want
	}{
		{
			name: "repo_error",
			args: args{
				ctx: context.Background(),
				id:  userID,
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			want: want{
				user: nil,
				err:  pkg.ErrNotFound,
			},
		},
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				id:  userID,
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(expectedUser, nil).Once()

				deps.link.EXPECT().
					PublicFile(mock.Anything).
					Return("").Times(2)

				deps.cache.EXPECT().
					Set(mock.Anything, usecase.GetKey(userID), mock.Anything, mock.Anything).
					Return(nil).Once()
			},
			want: want{
				user: expectedUser,
				err:  nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := usermocks.NewMockRepository(s.T())
			mockCache, userCache := newTestCache(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())

			deps := testDeps{
				repo:  mockRepo,
				cache: mockCache,
				link:  mockLink,
			}
			uc := usecase.NewFetchUseCase(usecase.Deps{
				Repo:  mockRepo,
				Cache: userCache,
				Link:  mockLink,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetByID(tc.args.ctx, tc.args.id)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.user, got)
			}
		})
	}
}

func (s *fetchUseCaseSuite) TestGetByEmail() {
	var (
		email = "test@example.com"
	)

	expectedUser := &user.User{
		ID:    1,
		Email: email,
	}

	type testDeps struct {
		repo  *usermocks.MockRepository
		cache *cachemocks.MockCache[*user.UserSummary]
		link  *commonmocks.MockLinkProvider
	}

	type args struct {
		ctx   context.Context
		email string
	}

	type want struct {
		user *user.User
		err  error
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      want
	}{
		{
			name: "repo_error",
			args: args{
				ctx:   context.Background(),
				email: email,
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(nil, pkg.ErrNotFound).Once()
			},
			want: want{
				user: nil,
				err:  pkg.ErrNotFound,
			},
		},
		{
			name: "success",
			args: args{
				ctx:   context.Background(),
				email: email,
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(expectedUser, nil).Once()

				deps.link.EXPECT().
					PublicFile(mock.Anything).
					Return("").Times(2)

				deps.cache.EXPECT().
					Set(mock.Anything, usecase.GetKey(expectedUser.ID), mock.Anything, mock.Anything).
					Return(nil).Once()
			},
			want: want{
				user: expectedUser,
				err:  nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := usermocks.NewMockRepository(s.T())
			mockCache, userCache := newTestCache(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())

			deps := testDeps{
				repo:  mockRepo,
				cache: mockCache,
				link:  mockLink,
			}
			uc := usecase.NewFetchUseCase(usecase.Deps{
				Repo:  mockRepo,
				Cache: userCache,
				Link:  mockLink,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetByEmail(tc.args.ctx, tc.args.email)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.user, got)
			}
		})
	}
}

func (s *fetchUseCaseSuite) TestGetSummary() {
	var (
		userID    int64 = 1
		avatarKey       = "avatar.jpg"
		coverKey        = "cover.jpg"
		avatarURL       = "http://cdn/avatar.jpg"
		coverURL        = "http://cdn/cover.jpg"
	)

	dbUser := &user.User{
		ID: userID,
		Profile: user.Profile{
			FullName:   "Test User",
			Avatar:     avatarKey,
			CoverImage: coverKey,
		},
		Status: user.Status{
			Verified: true,
		},
	}

	expectedSummary := &user.UserSummary{
		ID:         userID,
		FullName:   "Test User",
		Avatar:     avatarURL,
		CoverImage: coverURL,
		Verified:   true,
	}

	type testDeps struct {
		repo  *usermocks.MockRepository
		cache *cachemocks.MockCache[*user.UserSummary]
		link  *commonmocks.MockLinkProvider
	}

	type args struct {
		ctx context.Context
		id  int64
	}

	type want struct {
		summary *user.UserSummary
		err     error
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      want
	}{
		{
			name: "cache_hit",
			args: args{
				ctx: context.Background(),
				id:  userID,
			},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					Get(mock.Anything, usecase.GetKey(userID)).
					Return(expectedSummary, nil).Once()
			},
			want: want{
				summary: expectedSummary,
				err:     nil,
			},
		},
		{
			name: "cache_miss_repo_error",
			args: args{
				ctx: context.Background(),
				id:  userID,
			},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					Get(mock.Anything, usecase.GetKey(userID)).
					Return(nil, appcache.ErrCacheMiss).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			want: want{
				summary: nil,
				err:     pkg.ErrNotFound,
			},
		},
		{
			name: "cache_miss_success",
			args: args{
				ctx: context.Background(),
				id:  userID,
			},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					Get(mock.Anything, usecase.GetKey(userID)).
					Return(nil, appcache.ErrCacheMiss).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(dbUser, nil).Once()

				deps.link.EXPECT().
					PublicFile(avatarKey).
					Return(avatarURL).Once()

				deps.link.EXPECT().
					PublicFile(coverKey).
					Return(coverURL).Once()

				deps.cache.EXPECT().
					Set(mock.Anything, usecase.GetKey(userID), mock.Anything, 12*time.Hour).
					Return(nil).Once()
			},
			want: want{
				summary: expectedSummary,
				err:     nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := usermocks.NewMockRepository(s.T())
			mockCache, userCache := newTestCache(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())

			deps := testDeps{
				repo:  mockRepo,
				cache: mockCache,
				link:  mockLink,
			}
			uc := usecase.NewFetchUseCase(usecase.Deps{
				Repo:  mockRepo,
				Cache: userCache,
				Link:  mockLink,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetSummary(tc.args.ctx, tc.args.id)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.summary, got)
			}
		})
	}
}
