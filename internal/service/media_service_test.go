package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain"
	"air-social/internal/mocks"
	"air-social/pkg"
)

type mediaServiceSuite struct {
	suite.Suite
	cfg domain.FileConfig
}

func TestMediaServiceSuite(t *testing.T) {
	suite.Run(t, new(mediaServiceSuite))
}

func (s *mediaServiceSuite) SetupSuite() {
	s.cfg = domain.FileConfig{
		BucketPublic:     "test-bucket",
		PublicPathPrefix: "http://cdn.test",
	}
}

func (s *mediaServiceSuite) TestGetPresignedURL() {
	var (
		userID int64 = 1
		input        = domain.PresignedFileParams{
			UserID:   userID,
			Domain:   domain.DomainUser,
			Feature:  domain.FeatureAvatar,
			FileName: "test.jpg",
			FileType: "image/jpeg",
			FileSize: 1024,
		}
		presignedResp = domain.PresignedURLResult{
			UploadURL: "http://s3.upload",
			FormData:  map[string]string{"key": "value"},
		}
	)

	type args struct {
		input domain.PresignedFileParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(storage *mocks.FileStorage, cache *mocks.CacheStorage)
		want      domain.PresignedFileResponse
		wantErr   error
	}{
		{
			name: "unsupported_feature",
			args: args{input: domain.PresignedFileParams{
				Domain: domain.DomainUser, Feature: "invalid",
			}},
			setupMock: func(storage *mocks.FileStorage, cache *mocks.CacheStorage) {},
			wantErr:   pkg.ErrFileUnsupported,
		},
		{
			name: "file_too_large",
			args: args{input: domain.PresignedFileParams{
				Domain: domain.DomainUser, Feature: domain.FeatureAvatar, FileSize: domain.Limit5MB + 1,
			}},
			setupMock: func(storage *mocks.FileStorage, cache *mocks.CacheStorage) {},
			wantErr:   pkg.ErrFileTooLarge,
		},
		{
			name: "invalid_file_type",
			args: args{input: domain.PresignedFileParams{
				Domain: domain.DomainUser, Feature: domain.FeatureAvatar, FileType: "application/pdf",
			}},
			setupMock: func(storage *mocks.FileStorage, cache *mocks.CacheStorage) {},
			wantErr:   pkg.ErrFileTypeInvalid,
		},
		{
			name: "storage_error",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage, cache *mocks.CacheStorage) {
				storage.EXPECT().GetPresignedPostPolicy(mock.Anything, mock.Anything, mock.Anything).Return(domain.PresignedURLResult{}, assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "cache_error",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage, cache *mocks.CacheStorage) {
				storage.EXPECT().GetPresignedPostPolicy(mock.Anything, mock.Anything, mock.Anything).Return(presignedResp, nil).Once()
				cache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage, cache *mocks.CacheStorage) {
				storage.EXPECT().GetPresignedPostPolicy(mock.Anything,
					mock.MatchedBy(func(loc domain.StorageLocation) bool {
						return loc.Bucket == s.cfg.BucketPublic && len(loc.Key) > 0
					}),
					mock.MatchedBy(func(c domain.UploadConstraints) bool {
						return c.MaxSize == domain.Limit5MB && c.ContentType == input.FileType
					})).
					Return(presignedResp, nil).Once()

				cache.EXPECT().Set(mock.Anything, mock.AnythingOfType("string"), userID, domain.UploadExpiry).Return(nil).Once()
			},
			want: domain.PresignedFileResponse{
				UploadURL:     presignedResp.UploadURL,
				FormData:      presignedResp.FormData,
				ExpirySeconds: int64(domain.UploadExpiry.Seconds()),
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockStorage := mocks.NewFileStorage(s.T())
			mockCache := mocks.NewCacheStorage(s.T())
			svc := NewMediaService(mockStorage, mockCache, s.cfg)

			if tc.setupMock != nil {
				tc.setupMock(mockStorage, mockCache)
			}

			got, err := svc.GetPresignedURL(context.Background(), tc.args.input)

			if tc.wantErr != nil {
				s.Error(err)
				if tc.wantErr != assert.AnError {
					s.ErrorIs(err, tc.wantErr)
				}
			} else {
				s.NoError(err)
				s.Equal(tc.want.UploadURL, got.UploadURL)
				s.Equal(tc.want.FormData, got.FormData)
				s.NotEmpty(got.ObjectKey)
				s.NotEmpty(got.PublicURL)
				s.Contains(got.PublicURL, s.cfg.PublicPathPrefix)
			}
		})
	}
}

func (s *mediaServiceSuite) TestConfirmUpload() {
	var (
		userID    int64 = 1
		objectKey       = "user/1/avatar/image.jpg"
		input           = domain.ConfirmFileParams{
			UserID:    userID,
			ObjectKey: objectKey,
			Feature:   domain.FeatureAvatar,
		}
	)

	type args struct {
		input domain.ConfirmFileParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(storage *mocks.FileStorage, cache *mocks.CacheStorage)
		want      string
		wantErr   error
	}{
		{
			name: "success",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage, cache *mocks.CacheStorage) {
				cache.EXPECT().Get(mock.Anything, domain.GetUploadImageKey(objectKey), mock.Anything).
					Run(func(ctx context.Context, key string, dest any) {
						*dest.(*int64) = userID
					}).Return(nil).Once()

				storage.EXPECT().StatFile(mock.Anything, domain.StorageLocation{
					Bucket: s.cfg.BucketPublic,
					Key:    objectKey,
				}).Return(true, nil).Once()

				cache.EXPECT().Delete(mock.Anything, domain.GetUploadImageKey(objectKey)).Return(nil).Once()
			},
			want:    objectKey,
			wantErr: nil,
		},
		{
			name: "cache_get_error",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage, cache *mocks.CacheStorage) {
				cache.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
			want:    "",
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "user_mismatch",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage, cache *mocks.CacheStorage) {
				cache.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
					Run(func(ctx context.Context, key string, dest any) {
						*dest.(*int64) = 999 // Different user
					}).Return(nil).Once()
			},
			want:    "",
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "storage_stat_error",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage, cache *mocks.CacheStorage) {
				cache.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
					Run(func(ctx context.Context, key string, dest any) {
						*dest.(*int64) = userID
					}).Return(nil).Once()

				storage.EXPECT().StatFile(mock.Anything, mock.Anything).Return(false, assert.AnError).Once()
			},
			want:    "",
			wantErr: assert.AnError,
		},
		{
			name: "file_not_found",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage, cache *mocks.CacheStorage) {
				cache.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
					Run(func(ctx context.Context, key string, dest any) {
						*dest.(*int64) = userID
					}).Return(nil).Once()

				storage.EXPECT().StatFile(mock.Anything, mock.Anything).Return(false, nil).Once()
			},
			want:    "",
			wantErr: pkg.ErrNotFound,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockStorage := mocks.NewFileStorage(s.T())
			mockCache := mocks.NewCacheStorage(s.T())
			svc := NewMediaService(mockStorage, mockCache, s.cfg)

			if tc.setupMock != nil {
				tc.setupMock(mockStorage, mockCache)
			}

			got, err := svc.ConfirmUpload(context.Background(), tc.args.input)

			if tc.wantErr != nil {
				s.Error(err)
				if tc.wantErr != assert.AnError {
					s.ErrorIs(err, tc.wantErr)
				}
			} else {
				s.NoError(err)
				s.Equal(tc.want, got)
			}
		})
	}
}

func (s *mediaServiceSuite) TestDeleteFile() {
	objectKey := "file.jpg"

	tests := []struct {
		name      string
		objectKey string
		setupMock func(storage *mocks.FileStorage)
		wantErr   error
	}{
		{
			name:      "success",
			objectKey: objectKey,
			setupMock: func(storage *mocks.FileStorage) {
				storage.EXPECT().DeleteFile(mock.Anything, domain.StorageLocation{
					Bucket: s.cfg.BucketPublic,
					Key:    objectKey,
				}).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name:      "error",
			objectKey: objectKey,
			setupMock: func(storage *mocks.FileStorage) {
				storage.EXPECT().DeleteFile(mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockStorage := mocks.NewFileStorage(s.T())
			svc := NewMediaService(mockStorage, nil, s.cfg)

			if tc.setupMock != nil {
				tc.setupMock(mockStorage)
			}

			err := svc.DeleteFile(context.Background(), tc.objectKey)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *mediaServiceSuite) TestGetPublicURL() {
	tests := []struct {
		name      string
		objectKey string
		want      string
	}{
		{
			name:      "empty",
			objectKey: "",
			want:      "",
		},
		{
			name:      "success",
			objectKey: "image.jpg",
			want:      s.cfg.PublicPathPrefix + "/image.jpg",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			svc := NewMediaService(nil, nil, s.cfg)
			got := svc.GetPublicURL(tc.objectKey)
			s.Equal(tc.want, got)
		})
	}
}
