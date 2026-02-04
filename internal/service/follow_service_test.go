package service

import (
	"context"
	"testing"

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
		setupMocks func(followRepo *mocks.FollowRepository, userRepo *mocks.UserRepository, a args)
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
			setupMocks: func(followRepo *mocks.FollowRepository, userRepo *mocks.UserRepository, a args) {
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
			setupMocks: func(followRepo *mocks.FollowRepository, userRepo *mocks.UserRepository, a args) {
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
			setupMocks: func(followRepo *mocks.FollowRepository, userRepo *mocks.UserRepository, a args) {
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
			setupMocks: func(followRepo *mocks.FollowRepository, userRepo *mocks.UserRepository, a args) {
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

			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockUserRepo := mocks.NewUserRepository(s.T())
			mockFollowRepo := mocks.NewFollowRepository(s.T())
			followSvc := NewFollowService(mockFollowRepo, mockUserRepo)

			if tc.setupMocks != nil {
				tc.setupMocks(mockFollowRepo, mockUserRepo, tc.args)
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
		setupMocks func(followRepo *mocks.FollowRepository, a args)
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
			setupMocks: func(followRepo *mocks.FollowRepository, a args) {
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
			setupMocks: func(followRepo *mocks.FollowRepository, a args) {
				followRepo.
					EXPECT().
					Delete(mock.Anything, a.followerID, a.followeeID).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockFollowRepo := mocks.NewFollowRepository(s.T())
			mockUserRepo := mocks.NewUserRepository(s.T())

			followSvc := NewFollowService(mockFollowRepo, mockUserRepo)

			if tc.setupMocks != nil {
				tc.setupMocks(mockFollowRepo, tc.args)
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

}
