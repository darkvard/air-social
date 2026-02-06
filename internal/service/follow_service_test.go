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
					Twice()

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
				// Wait for async cache invalidation goroutine
				time.Sleep(50 * time.Millisecond)
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
					Twice()
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
				// Wait for async cache invalidation goroutine
				time.Sleep(50 * time.Millisecond)
			}
		})
	}
}

func (s *followServiceSuite) TestGetFollowings() {

}

func (s *followServiceSuite) TestGetFollowers() {

}
