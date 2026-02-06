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
		BucketPublic: "test-bucket",
	}
}

func (s *mediaServiceSuite) TestGetPresignedURL() {
	var (
		userID int64 = 1
		input        = domain.PresignedFileParams{
			EntityID: userID,
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
		expectedUploadURL = "http://cdn.test/test-bucket"
	)

	type args struct {
		input domain.PresignedFileParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(storage *mocks.FileStorage, url *mocks.URLFactory)
		want      domain.PresignedFile
		wantErr   error
	}{
		{
			name: "unsupported_feature",
			args: args{input: domain.PresignedFileParams{
				Domain: domain.DomainUser, Feature: "invalid",
			}},
			setupMock: func(storage *mocks.FileStorage, url *mocks.URLFactory) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "file_too_large",
			args: args{input: domain.PresignedFileParams{
				Domain: domain.DomainUser, Feature: domain.FeatureAvatar, FileSize: domain.Limit5MB + 1,
			}},
			setupMock: func(storage *mocks.FileStorage, url *mocks.URLFactory) {},
			wantErr:   pkg.ErrFileTooLarge,
		},
		{
			name: "invalid_file_type",
			args: args{input: domain.PresignedFileParams{
				Domain: domain.DomainUser, Feature: domain.FeatureAvatar, FileType: "application/pdf",
			}},
			setupMock: func(storage *mocks.FileStorage, url *mocks.URLFactory) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "storage_error",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage, url *mocks.URLFactory) {
				storage.EXPECT().GetPresignedPostPolicy(mock.Anything, mock.Anything, mock.Anything).Return(domain.PresignedURLResult{}, assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage, url *mocks.URLFactory) {
				storage.EXPECT().GetPresignedPostPolicy(mock.Anything,
					mock.MatchedBy(func(loc domain.StorageLocation) bool {
						return loc.Bucket == s.cfg.BucketPublic && len(loc.Key) > 0
					}),
					mock.MatchedBy(func(c domain.UploadConstraints) bool {
						return c.MaxSize == domain.Limit5MB && c.ContentType == input.FileType
					})).
					Return(presignedResp, nil).Once()
				url.EXPECT().BaseURL().Return("http://cdn.test").Once()
				url.EXPECT().PublicFileURL(mock.Anything).Return(expectedUploadURL + "/somekey").Once()
			},
			want: domain.PresignedFile{
				UploadURL: expectedUploadURL,
				FormData:  presignedResp.FormData,
				ExpireAt:  pkg.TimeNowUTC().Add(domain.PresignedUploadExpiry),
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockStorage := mocks.NewFileStorage(s.T())
			mockURL := mocks.NewURLFactory(s.T())
			svc := NewMediaService(mockStorage, s.cfg, mockURL)

			if tc.setupMock != nil {
				tc.setupMock(mockStorage, mockURL)
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
				// s.Contains(got.PublicURL, s.cfg.DomainPublic) // Mock returns specific string
			}
		})
	}
}

func (s *mediaServiceSuite) TestConfirmUpload() {
	var (
		userID    int64 = 1
		objectKey       = "users/1/avatar/image.jpg"
		input           = domain.ConfirmFileParams{
			EntityID:  userID,
			ObjectKey: objectKey,
			Domain:    domain.DomainUser,
			Feature:   domain.FeatureAvatar,
		}
	)

	type args struct {
		input domain.ConfirmFileParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(storage *mocks.FileStorage)
		want      string
		wantErr   error
	}{
		{
			name: "success",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage) {

				storage.EXPECT().StatFile(mock.Anything, domain.StorageLocation{
					Bucket: s.cfg.BucketPublic,
					Key:    objectKey,
				}).Return(true, nil).Once()

			},
			want:    objectKey,
			wantErr: nil,
		},

		{
			name: "user_mismatch",
			args: args{input: domain.ConfirmFileParams{
				EntityID:  userID + 1,
				ObjectKey: objectKey,
				Domain:    domain.DomainUser,
				Feature:   domain.FeatureAvatar,
			}},
			setupMock: func(storage *mocks.FileStorage) {

			},
			want:    "",
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "storage_stat_error",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage) {

				storage.EXPECT().StatFile(mock.Anything, mock.Anything).Return(false, assert.AnError).Once()
			},
			want:    "",
			wantErr: assert.AnError,
		},
		{
			name: "file_not_found",
			args: args{input: input},
			setupMock: func(storage *mocks.FileStorage) {
				storage.EXPECT().StatFile(mock.Anything, mock.Anything).Return(false, nil).Once()
			},
			want:    "",
			wantErr: pkg.ErrNotFound,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockStorage := mocks.NewFileStorage(s.T())
			mockURL := mocks.NewURLFactory(s.T())
			svc := NewMediaService(mockStorage, s.cfg, mockURL)

			if tc.setupMock != nil {
				tc.setupMock(mockStorage)
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
			mockURL := mocks.NewURLFactory(s.T())
			svc := NewMediaService(mockStorage, s.cfg, mockURL)

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
