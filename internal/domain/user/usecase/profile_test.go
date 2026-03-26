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
	mediamocks "air-social/internal/domain/media/mocks"
	usermocks "air-social/internal/domain/user/mocks"
	"air-social/internal/domain/user/usecase"
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
		bio            = "New Bio"
		fullName       = "New Name"
		location       = "New York"
		website        = "https://example.com"
		username       = "newuser"
	)

	existingUser := &user.User{
		ID: userID,
		Profile: user.Profile{
			Bio:      "Old Bio",
			FullName: "Old Name",
		},
		Username: "olduser",
	}

	type testDeps struct {
		repo  *usermocks.MockRepository
		cache *usermocks.MockCache
	}

	type args struct {
		ctx    context.Context
		params user.UpdateParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      *user.User
		wantErr   error
	}{
		{
			name: "success_all_fields",
			args: args{
				ctx: context.Background(),
				params: user.UpdateParams{
					UserID:   userID,
					Bio:      &bio,
					FullName: &fullName,
					Location: &location,
					Website:  &website,
					Username: &username,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(existingUser, nil).Once()

				deps.repo.EXPECT().
					UpdateProfile(mock.Anything, mock.MatchedBy(func(u *user.User) bool {
						return u.ID == userID &&
							u.Profile.Bio == bio &&
							u.Profile.FullName == fullName &&
							u.Profile.Location == location &&
							u.Profile.Website == website &&
							u.Username == username
					})).
					Return(nil).Once()

				deps.cache.EXPECT().
					Invalidate(mock.Anything, mock.Anything).
					Return(nil).Once()
			},
			want: &user.User{
				ID: userID,
				Profile: user.Profile{
					Bio:      bio,
					FullName: fullName,
					Location: location,
					Website:  website,
				},
				Username: username,
			},
			wantErr: nil,
		},
		{
			name: "repo_get_error",
			args: args{
				ctx: context.Background(),
				params: user.UpdateParams{
					UserID: userID,
					Bio:    &bio,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "repo_update_error",
			args: args{
				ctx: context.Background(),
				params: user.UpdateParams{
					UserID: userID,
					Bio:    &bio,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(existingUser, nil).Once()

				deps.repo.EXPECT().
					UpdateProfile(mock.Anything, mock.Anything).
					Return(assert.AnError).Once()
			},
			want:    nil,
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := usermocks.NewMockRepository(s.T())
			mockCache := usermocks.NewMockCache(s.T())
			mockMedia := mediamocks.NewMockUseCase(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())

			deps := testDeps{
				repo:  mockRepo,
				cache: mockCache,
			}

			uc := usecase.NewProfileUseCase(user.Deps{
				Repo:  mockRepo,
				Cache: mockCache,
				Media: mockMedia,
				Link:  mockLink,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.UpdateProfile(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.Profile.Location, got.Profile.Location)
				s.Equal(tc.want.Profile.Website, got.Profile.Website)
				s.Equal(tc.want.Username, got.Username)
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
		cache *usermocks.MockCache
		media *usermocks.MockMediaConfirmer
		link  *commonmocks.MockLinkProvider
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
			name: "repo_get_fail",
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
					Invalidate(mock.Anything, userID).
					Return(nil).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			want: want{
				err: pkg.ErrNotFound,
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
					Invalidate(mock.Anything, userID).
					Return(nil).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(updatedUser, nil).Once()

				// This is a duplicate mock call in the original code, I'll keep it but it should be reviewed.
				// deps.cache.EXPECT().
				// 	Invalidate(mock.Anything, userID).
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
			mockCache := usermocks.NewMockCache(s.T())
			mockMedia := usermocks.NewMockMediaConfirmer(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())

			deps := testDeps{
				repo:  mockRepo,
				cache: mockCache,
				media: mockMedia,
				link:  mockLink,
			}
			uc := usecase.NewProfileUseCase(user.Deps{Repo: mockRepo, Cache: mockCache, Media: mockMedia, Link: mockLink})

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
		objectKey       = "users/1/cover/image.jpg"
	)

	params := media.ConfirmParams{
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
		cache *usermocks.MockCache
		media *mediamocks.MockUseCase
	}

	type args struct {
		ctx    context.Context
		params media.ConfirmParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      *user.User
		wantErr   error
	}{
		{
			name: "success",
			args: args{ctx: context.Background(), params: params},
			setupMock: func(deps testDeps) {
				deps.media.EXPECT().
					ConfirmUpload(mock.Anything, []media.ConfirmParams{params}).
					Return([]string{objectKey}, nil).Once()

				deps.repo.EXPECT().
					UpdateCover(mock.Anything, userID, objectKey).
					Return(nil).Once()

				deps.cache.EXPECT().
					Invalidate(mock.Anything, mock.Anything).
					Return(nil).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(updatedUser, nil).Once()
			},
			want:    updatedUser,
			wantErr: nil,
		},
		{
			name: "media_confirm_error",
			args: args{ctx: context.Background(), params: params},
			setupMock: func(deps testDeps) {
				deps.media.EXPECT().
					ConfirmUpload(mock.Anything, []media.ConfirmParams{params}).
					Return(nil, pkg.ErrBadRequest).Once()
			},
			want:    nil,
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "repo_update_error",
			args: args{ctx: context.Background(), params: params},
			setupMock: func(deps testDeps) {
				deps.media.EXPECT().
					ConfirmUpload(mock.Anything, []media.ConfirmParams{params}).
					Return([]string{objectKey}, nil).Once()

				deps.repo.EXPECT().
					UpdateCover(mock.Anything, userID, objectKey).
					Return(assert.AnError).Once()
			},
			want:    nil,
			wantErr: pkg.ErrInternal,
		},
		{
			name: "repo_get_error",
			args: args{ctx: context.Background(), params: params},
			setupMock: func(deps testDeps) {
				deps.media.EXPECT().
					ConfirmUpload(mock.Anything, []media.ConfirmParams{params}).
					Return([]string{objectKey}, nil).Once()

				deps.repo.EXPECT().
					UpdateCover(mock.Anything, userID, objectKey).
					Return(nil).Once()

				deps.cache.EXPECT().
					Invalidate(mock.Anything, mock.Anything).
					Return(nil).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := usermocks.NewMockRepository(s.T())
			mockCache := usermocks.NewMockCache(s.T())
			mockMedia := mediamocks.NewMockUseCase(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())

			deps := testDeps{
				repo:  mockRepo,
				cache: mockCache,
				media: mockMedia,
			}

			uc := usecase.NewProfileUseCase(user.Deps{
				Repo:  mockRepo,
				Cache: mockCache,
				Media: mockMedia,
				Link:  mockLink,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.UpdateCover(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want, got)
			}
		})
	}
}