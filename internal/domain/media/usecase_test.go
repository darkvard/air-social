package media_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/common"
	commonmocks "air-social/internal/domain/common/mocks"
	"air-social/internal/domain/media"
	mediamocks "air-social/internal/domain/media/mocks"
	"air-social/pkg"
)

type mediaUseCaseSuite struct {
	suite.Suite
}

func TestMediaUseCaseSuite(t *testing.T) {
	suite.Run(t, new(mediaUseCaseSuite))
}

func (s *mediaUseCaseSuite) TestGenerateKey() {
	type testDeps struct{}

	type args struct {
		params media.PresignParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		check     func(key string)
	}{
		{
			name: "success",
			args: args{
				params: media.PresignParams{
					Domain:   media.DomainUser,
					EntityID: 1,
					Feature:  media.FeatureAvatar,
					FileName: "image.png",
				},
			},
			setupMock: func(deps testDeps) {},
			check: func(key string) {
				s.Contains(key, "users/1/avatar/")
				s.Contains(key, ".png")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			uc := media.NewUseCase(media.Deps{})
			if tc.setupMock != nil {
				tc.setupMock(testDeps{})
			}
			got := uc.GenerateKey(tc.args.params)
			tc.check(got)
		})
	}
}

func (s *mediaUseCaseSuite) TestGetPresignedURL() {
	var (
		bucketName = "public-bucket"
		baseURL    = "http://api.example.com"
		publicURL  = "http://cdn.example.com/file.jpg"
		uploadURL  = "http://api.example.com/public-bucket"
	)

	validParam := media.PresignParams{
		EntityID: 1,
		FileName: "avatar.jpg",
		FileType: "image/jpeg",
		FileSize: 1024,
		Domain:   media.DomainUser,
		Feature:  media.FeatureAvatar,
	}

	uploadForm := media.UploadForm{
		URL:    "http://minio/upload",
		Fields: map[string]string{"key": "value"},
	}

	type testDeps struct {
		storage *mediamocks.MockStorage
		link    *commonmocks.MockLinkProvider
		route   *commonmocks.MockRouteProvider
	}

	type args struct {
		ctx    context.Context
		params []media.PresignParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantLen   int
		wantErr   error
	}{
		{
			name: "invalid_domain",
			args: args{
				ctx: context.Background(),
				params: []media.PresignParams{{
					Domain: "invalid",
				}},
			},
			setupMock: func(deps testDeps) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "invalid_feature",
			args: args{
				ctx: context.Background(),
				params: []media.PresignParams{{
					Domain:  media.DomainUser,
					Feature: "invalid",
				}},
			},
			setupMock: func(deps testDeps) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "file_too_large",
			args: args{
				ctx: context.Background(),
				params: []media.PresignParams{{
					Domain:   media.DomainUser,
					Feature:  media.FeatureAvatar,
					FileSize: media.Limit5MB + 1,
				}},
			},
			setupMock: func(deps testDeps) {},
			wantErr:   pkg.ErrFileTooLarge,
		},
		{
			name: "invalid_file_type",
			args: args{
				ctx: context.Background(),
				params: []media.PresignParams{{
					Domain:   media.DomainUser,
					Feature:  media.FeatureAvatar,
					FileType: "application/pdf",
				}},
			},
			setupMock: func(deps testDeps) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "storage_error",
			args: args{
				ctx:    context.Background(),
				params: []media.PresignParams{validParam},
			},
			setupMock: func(deps testDeps) {
				deps.storage.EXPECT().
					PresignUpload(mock.Anything, mock.Anything, mock.Anything).
					Return(media.UploadForm{}, assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				params: []media.PresignParams{validParam},
			},
			setupMock: func(deps testDeps) {
				deps.storage.EXPECT().
					PresignUpload(mock.Anything, mock.MatchedBy(func(loc media.Location) bool {
						return loc.Bucket == bucketName
					}), mock.MatchedBy(func(c media.Constraints) bool {
						return c.ContentType == validParam.FileType && c.MaxSize == media.Limit5MB
					})).
					Return(uploadForm, nil).Once()

				deps.route.EXPECT().
					BaseURL().
					Return(baseURL).Once()

				deps.link.EXPECT().
					PublicFile(mock.Anything).
					Return(publicURL).Once()
			},
			wantLen: 1,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockStorage := mediamocks.NewMockStorage(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())
			mockRoute := commonmocks.NewMockRouteProvider(s.T())

			deps := testDeps{
				storage: mockStorage,
				link:    mockLink,
				route:   mockRoute,
			}

			uc := media.NewUseCase(media.Deps{
				Bucket:  media.Bucket{Public: bucketName},
				Storage: mockStorage,
				Link: common.AppLinkManager{
					LinkProvider:  mockLink,
					RouteProvider: mockRoute,
				},
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetPresignedURL(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Len(got, tc.wantLen)
				if tc.wantLen > 0 {
					s.Equal(uploadURL, got[0].UploadURL)
					s.Equal(publicURL, got[0].PublicURL)
					s.Equal(uploadForm.Fields, got[0].FormData)
				}
			}
		})
	}
}

func (s *mediaUseCaseSuite) TestConfirmUpload() {
	var (
		bucketName = "public-bucket"
		entityID   = int64(1)
		objectKey  = "users/1/avatar/file.jpg"
	)

	validParam := media.ConfirmParams{
		EntityID:  entityID,
		ObjectKey: objectKey,
		Domain:    media.DomainUser,
		Feature:   media.FeatureAvatar,
	}

	type testDeps struct {
		storage *mediamocks.MockStorage
	}

	type args struct {
		ctx    context.Context
		params []media.ConfirmParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      []string
		wantErr   error
	}{
		{
			name: "invalid_domain",
			args: args{
				ctx: context.Background(),
				params: []media.ConfirmParams{{
					Domain: "invalid",
				}},
			},
			setupMock: func(deps testDeps) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "invalid_feature",
			args: args{
				ctx: context.Background(),
				params: []media.ConfirmParams{{
					Domain:  media.DomainUser,
					Feature: "invalid",
				}},
			},
			setupMock: func(deps testDeps) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "key_mismatch",
			args: args{
				ctx: context.Background(),
				params: []media.ConfirmParams{{
					EntityID:  entityID,
					ObjectKey: "wrong/path/file.jpg",
					Domain:    media.DomainUser,
					Feature:   media.FeatureAvatar,
				}},
			},
			setupMock: func(deps testDeps) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "storage_error",
			args: args{
				ctx:    context.Background(),
				params: []media.ConfirmParams{validParam},
			},
			setupMock: func(deps testDeps) {
				deps.storage.EXPECT().
					Exists(mock.Anything, media.Location{Bucket: bucketName, Key: objectKey}).
					Return(false, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "file_not_found",
			args: args{
				ctx:    context.Background(),
				params: []media.ConfirmParams{validParam},
			},
			setupMock: func(deps testDeps) {
				deps.storage.EXPECT().
					Exists(mock.Anything, media.Location{Bucket: bucketName, Key: objectKey}).
					Return(false, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				params: []media.ConfirmParams{validParam},
			},
			setupMock: func(deps testDeps) {
				deps.storage.EXPECT().
					Exists(mock.Anything, media.Location{Bucket: bucketName, Key: objectKey}).
					Return(true, nil).Once()
			},
			want:    []string{objectKey},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockStorage := mediamocks.NewMockStorage(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())
			mockRoute := commonmocks.NewMockRouteProvider(s.T())

			deps := testDeps{
				storage: mockStorage,
			}

			uc := media.NewUseCase(media.Deps{
				Bucket:  media.Bucket{Public: bucketName},
				Storage: mockStorage,
				Link: common.AppLinkManager{
					LinkProvider:  mockLink,
					RouteProvider: mockRoute,
				},
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.ConfirmUpload(tc.args.ctx, tc.args.params)

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

func (s *mediaUseCaseSuite) TestVerifyMedia() {
	var (
		bucketName = "public-bucket"
		keys       = []string{"key1", "key2"}
	)

	type testDeps struct {
		storage *mediamocks.MockStorage
	}

	type args struct {
		ctx  context.Context
		keys []string
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "storage_error",
			args: args{ctx: context.Background(), keys: keys},
			setupMock: func(deps testDeps) {
				deps.storage.EXPECT().
					Exists(mock.Anything, media.Location{Bucket: bucketName, Key: keys[0]}).
					Return(false, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "not_found",
			args: args{ctx: context.Background(), keys: keys},
			setupMock: func(deps testDeps) {
				deps.storage.EXPECT().
					Exists(mock.Anything, media.Location{Bucket: bucketName, Key: keys[0]}).
					Return(false, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "success",
			args: args{ctx: context.Background(), keys: keys},
			setupMock: func(deps testDeps) {
				deps.storage.EXPECT().
					Exists(mock.Anything, media.Location{Bucket: bucketName, Key: keys[0]}).
					Return(true, nil).Once()
				deps.storage.EXPECT().
					Exists(mock.Anything, media.Location{Bucket: bucketName, Key: keys[1]}).
					Return(true, nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockStorage := mediamocks.NewMockStorage(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())
			mockRoute := commonmocks.NewMockRouteProvider(s.T())

			deps := testDeps{storage: mockStorage}
			uc := media.NewUseCase(media.Deps{
				Bucket:  media.Bucket{Public: bucketName},
				Storage: mockStorage,
				Link: common.AppLinkManager{
					LinkProvider:  mockLink,
					RouteProvider: mockRoute,
				},
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.VerifyMedia(tc.args.ctx, tc.args.keys)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *mediaUseCaseSuite) TestDeleteMedia() {
	var (
		bucketName = "public-bucket"
		keys       = []string{"key1"}
	)

	type testDeps struct {
		storage *mediamocks.MockStorage
	}

	type args struct {
		ctx  context.Context
		keys []string
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "storage_error",
			args: args{ctx: context.Background(), keys: keys},
			setupMock: func(deps testDeps) {
				deps.storage.EXPECT().
					Delete(mock.Anything, media.Location{Bucket: bucketName, Key: keys[0]}).
					Return(assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: args{ctx: context.Background(), keys: keys},
			setupMock: func(deps testDeps) {
				deps.storage.EXPECT().
					Delete(mock.Anything, media.Location{Bucket: bucketName, Key: keys[0]}).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockStorage := mediamocks.NewMockStorage(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())
			mockRoute := commonmocks.NewMockRouteProvider(s.T())

			deps := testDeps{storage: mockStorage}
			uc := media.NewUseCase(media.Deps{
				Bucket:  media.Bucket{Public: bucketName},
				Storage: mockStorage,
				Link: common.AppLinkManager{
					LinkProvider:  mockLink,
					RouteProvider: mockRoute,
				},
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.DeleteMedia(tc.args.ctx, tc.args.keys)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}
