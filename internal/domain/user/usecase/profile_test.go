package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	commonmocks "air-social/internal/domain/common/mocks"
	"air-social/internal/domain/media"
	"air-social/internal/domain/user"
	usermocks "air-social/internal/domain/user/mocks"
	"air-social/internal/domain/user/usecase"
	usecasemocks "air-social/internal/domain/user/usecase/mocks"
	"air-social/pkg"
)

type profileUseCaseSuite struct {
	suite.Suite
}

func TestProfileUseCaseSuite(t *testing.T) {
	suite.Run(t, new(profileUseCaseSuite))
}

func (s *profileUseCaseSuite) TestUpdateProfile() {
	var (
		userID   int64 = 1
		fullName       = "New Name"
		bio            = "New Bio"
	)

	baseInput := user.UpdateParams{
		UserID:   userID,
		FullName: &fullName,
		Bio:      &bio,
	}

	existingUser := &user.User{
		ID: userID,
		Profile: user.Profile{
			FullName: "Old Name",
			Bio:      "Old Bio",
		},
	}

	type testDeps struct {
		repo  *usermocks.MockRepository
		cache *commonmocks.MockCache
		media *usecasemocks.MockMediaConfirmer
	}

	type args struct {
		ctx    context.Context
		params user.UpdateParams
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
			name: "get_user_error",
			args: args{
				ctx:    context.Background(),
				params: baseInput,
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
			name: "update_repo_error",
			args: args{
				ctx:    context.Background(),
				params: baseInput,
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(existingUser, nil).Once()

				deps.repo.EXPECT().
					UpdateProfile(mock.Anything, mock.MatchedBy(func(u *user.User) bool {
						return u.Profile.FullName == fullName && u.Profile.Bio == bio
					})).
					Return(assert.AnError).Once()
			},
			want: want{
				user: nil,
				err:  pkg.ErrInternal,
			},
		},
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				params: baseInput,
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(existingUser, nil).Once()

				deps.repo.EXPECT().
					UpdateProfile(mock.Anything, mock.MatchedBy(func(u *user.User) bool {
						return u.Profile.FullName == fullName && u.Profile.Bio == bio
					})).
					Return(nil).Once()

				deps.cache.EXPECT().
					Delete(mock.Anything, usecase.GetKey(userID)).
					Return(nil).Once()
			},
			want: want{
				user: existingUser,
				err:  nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := usermocks.NewMockRepository(s.T())
			mockCache := commonmocks.NewMockCache(s.T())
			mockMedia := usecasemocks.NewMockMediaConfirmer(s.T())

			deps := testDeps{
				repo:  mockRepo,
				cache: mockCache,
				media: mockMedia,
			}
			uc := usecase.NewProfileUseCase(usecase.Deps{
				Repo:  mockRepo,
				Cache: mockCache,
				Media: mockMedia,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.UpdateProfile(tc.args.ctx, tc.args.params)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.user.Profile.FullName, got.Profile.FullName)
				s.Equal(tc.want.user.Profile.Bio, got.Profile.Bio)
			}
		})
	}
}

func (s *profileUseCaseSuite) TestUpdateAvatar() {
	var (
		userID    int64 = 1
		objectKey       = "users/1/avatar/img.jpg"
	)

	validParams := media.ConfirmParams{
		EntityID:  userID,
		ObjectKey: objectKey,
		Domain:    media.DomainUser,
		Feature:   media.FeatureAvatar,
	}

	updatedUser := &user.User{
		ID: userID,
		Profile: user.Profile{
			Avatar: objectKey,
		},
	}

	type testDeps struct {
		repo  *usermocks.MockRepository
		cache *commonmocks.MockCache
		media *usecasemocks.MockMediaConfirmer
	}

	type args struct {
		ctx    context.Context
		params media.ConfirmParams
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
			name: "invalid_domain_feature",
			args: args{
				ctx: context.Background(),
				params: media.ConfirmParams{
					Domain:  "invalid",
					Feature: "invalid",
				},
			},
			setupMock: func(deps testDeps) {},
			want: want{
				err: pkg.ErrBadRequest,
			},
		},
		{
			name: "confirm_upload_fail",
			args: args{
				ctx:    context.Background(),
				params: validParams,
			},
			setupMock: func(deps testDeps) {
				deps.media.EXPECT().
					ConfirmUpload(mock.Anything, []media.ConfirmParams{validParams}).
					Return(nil, assert.AnError).Once()
			},
			want: want{
				err: pkg.ErrInternal,
			},
		},
		{
			name: "repo_update_fail",
			args: args{
				ctx:    context.Background(),
				params: validParams,
			},
			setupMock: func(deps testDeps) {
				deps.media.EXPECT().
					ConfirmUpload(mock.Anything, []media.ConfirmParams{validParams}).
					Return([]string{objectKey}, nil).Once()

				deps.repo.EXPECT().
					UpdateAvatar(mock.Anything, userID, objectKey).
					Return(assert.AnError).Once()
			},
			want: want{
				err: pkg.ErrInternal,
			},
		},
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				params: validParams,
			},
			setupMock: func(deps testDeps) {
				deps.media.EXPECT().
					ConfirmUpload(mock.Anything, []media.ConfirmParams{validParams}).
					Return([]string{objectKey}, nil).Once()

				deps.repo.EXPECT().
					UpdateAvatar(mock.Anything, userID, objectKey).
					Return(nil).Once()

				deps.cache.EXPECT().
					Delete(mock.Anything, usecase.GetKey(userID)).
					Return(nil).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(updatedUser, nil).Once()

				// This is a duplicate mock call in the original code, I'll keep it but it should be reviewed.
				// deps.cache.EXPECT().
				// 	Delete(mock.Anything, usecase.GetKey(userID)).
				// 	Return(nil).Once()
			},
			want: want{
				user: updatedUser,
				err:  nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := usermocks.NewMockRepository(s.T())
			mockCache := commonmocks.NewMockCache(s.T())
			mockMedia := usecasemocks.NewMockMediaConfirmer(s.T())

			deps := testDeps{
				repo:  mockRepo,
				cache: mockCache,
				media: mockMedia,
			}
			uc := usecase.NewProfileUseCase(usecase.Deps{Repo: mockRepo, Cache: mockCache, Media: mockMedia})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.UpdateAvatar(tc.args.ctx, tc.args.params)

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

func (s *profileUseCaseSuite) TestUpdateCover() {
	var (
		userID    int64 = 1
		objectKey       = "users/1/cover/img.jpg"
	)

	validParams := media.ConfirmParams{
		EntityID:  userID,
		ObjectKey: objectKey,
		Domain:    media.DomainUser,
		Feature:   media.FeatureCover,
	}

	updatedUser := &user.User{
		ID: userID,
		Profile: user.Profile{
			CoverImage: objectKey,
		},
	}

	type testDeps struct {
		repo  *usermocks.MockRepository
		cache *commonmocks.MockCache
		media *usecasemocks.MockMediaConfirmer
	}

	type args struct {
		ctx    context.Context
		params media.ConfirmParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantUser  *user.User
		wantErr   error
	}{
		{
			name: "success",
			args: args{ctx: context.Background(), params: validParams},
			setupMock: func(deps testDeps) {
				deps.media.EXPECT().
					ConfirmUpload(mock.Anything, []media.ConfirmParams{validParams}).
					Return([]string{objectKey}, nil).Once()

				deps.repo.EXPECT().
					UpdateCover(mock.Anything, userID, objectKey).
					Return(nil).Once()

				deps.cache.EXPECT().
					Delete(mock.Anything, usecase.GetKey(userID)).
					Return(nil).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(updatedUser, nil).Once()
			},
			wantUser: updatedUser,
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := usermocks.NewMockRepository(s.T())
			mockCache := commonmocks.NewMockCache(s.T())
			mockMedia := usecasemocks.NewMockMediaConfirmer(s.T())

			deps := testDeps{
				repo:  mockRepo,
				cache: mockCache,
				media: mockMedia,
			}
			uc := usecase.NewProfileUseCase(usecase.Deps{Repo: mockRepo, Cache: mockCache, Media: mockMedia})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}
			got, err := uc.UpdateCover(tc.args.ctx, tc.args.params)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
				s.Equal(tc.wantUser, got)
			}
		})
	}
}
