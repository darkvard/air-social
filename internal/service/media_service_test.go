package service

// import (
// 	"context"
// 	"testing"
//
//

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/suite"

// 	"air-social/internal/config"
// 	"air-social/internal/domain"
// 	"air-social/internal/mocks"
// 	"air-social/pkg"
// )

// type mediaServiceSuite struct {
// 	suite.Suite
// 	cfg config.MinioStorageConfig
// }

// func TestMediaServiceSuite(t *testing.T) {
// 	suite.Run(t, new(mediaServiceSuite))
// }

// func (s *mediaServiceSuite) SetupSuite() {
// 	s.cfg = config.MinioStorageConfig{
// 		BucketPublic: "test-bucket",
// 	}
// }

// func (s *mediaServiceSuite) TestGetPresignedURL() {
// 	var (
// 		userID int64 = 1
// 		input        = domain.PresignedFileParams{
// 			EntityID: userID,
// 			Domain:   domain.DomainUser,
// 			Feature:  domain.FeatureAvatar,
// 			FileName: "test.jpg",
// 			FileType: "image/jpeg",
// 			FileSize: 1024,
// 		}
// 		presignedResp = domain.PresignedURLResult{
// 			UploadURL: "http://s3.upload",
// 			FormData:  map[string]string{"key": "value"},
// 		}
// 		expectedUploadURL = "http://cdn.test/test-bucket"
// 	)

// 	type args struct {
// 		input []domain.PresignedFileParams
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(storage *mocks.FileStorage, url *mocks.URLFactory)
// 		want      []domain.PresignedFile
// 		wantErr   error
// 	}{
// 		{
// 			name: "unsupported_feature",
// 			args: args{input: []domain.PresignedFileParams{{
// 				Domain: domain.DomainUser, Feature: "invalid",
// 			}}},
// 			setupMock: func(storage *mocks.FileStorage, url *mocks.URLFactory) {},
// 			wantErr:   pkg.ErrBadRequest,
// 		},
// 		{
// 			name: "file_too_large",
// 			args: args{input: []domain.PresignedFileParams{{
// 				Domain: domain.DomainUser, Feature: domain.FeatureAvatar, FileSize: domain.Limit5MB + 1,
// 			}}},
// 			setupMock: func(storage *mocks.FileStorage, url *mocks.URLFactory) {},
// 			wantErr:   pkg.ErrFileTooLarge,
// 		},
// 		{
// 			name: "invalid_file_type",
// 			args: args{input: []domain.PresignedFileParams{{
// 				Domain: domain.DomainUser, Feature: domain.FeatureAvatar, FileType: "application/pdf",
// 			}}},
// 			setupMock: func(storage *mocks.FileStorage, url *mocks.URLFactory) {},
// 			wantErr:   pkg.ErrBadRequest,
// 		},
// 		{
// 			name: "storage_error",
// 			args: args{input: []domain.PresignedFileParams{input}},
// 			setupMock: func(storage *mocks.FileStorage, url *mocks.URLFactory) {
// 				storage.EXPECT().GetPresignedPostPolicy(mock.Anything, mock.Anything, mock.Anything).Return(domain.PresignedURLResult{}, assert.AnError).Once()
// 			},
// 			wantErr: assert.AnError,
// 		},
// 		{
// 			name: "success",
// 			args: args{input: []domain.PresignedFileParams{input}},
// 			setupMock: func(storage *mocks.FileStorage, url *mocks.URLFactory) {
// 				storage.EXPECT().GetPresignedPostPolicy(mock.Anything,
// 					mock.MatchedBy(func(loc domain.StorageLocation) bool {
// 						return loc.Bucket == s.cfg.BucketPublic && len(loc.Key) > 0
// 					}),
// 					mock.MatchedBy(func(c domain.UploadConstraints) bool {
// 						return c.MaxSize == domain.Limit5MB && c.ContentType == input.FileType
// 					})).
// 					Return(presignedResp, nil).Once()
// 				url.EXPECT().BaseURL().Return("http://cdn.test").Once()
// 				url.EXPECT().PublicFileURL(mock.Anything).Return(expectedUploadURL + "/somekey").Once()
// 			},
// 			want: []domain.PresignedFile{{
// 				FileName:  input.FileName,
// 				UploadURL: expectedUploadURL,
// 				FormData:  presignedResp.FormData,
// 				ExpireAt:  pkg.TimeNowUTC().Add(domain.PresignedUploadExpiry),
// 			}},
// 			wantErr: nil,
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			mockStorage := mocks.NewFileStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			svc := NewMediaService(mockStorage, s.cfg, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(mockStorage, mockURL)
// 			}

// 			got, err := svc.GetPresignedURL(context.Background(), tc.args.input)

// 			if tc.wantErr != nil {
// 				s.Error(err)
// 				if tc.wantErr != assert.AnError {
// 					s.ErrorIs(err, tc.wantErr)
// 				}
// 			} else {
// 				s.NoError(err)
// 				s.Equal(tc.want[0].UploadURL, got[0].UploadURL)
// 				s.Equal(tc.want[0].FileName, got[0].FileName)
// 				s.Equal(tc.want[0].FormData, got[0].FormData)
// 				s.NotEmpty(got[0].ObjectKey)
// 				s.NotEmpty(got[0].PublicURL)
// 				// s.Contains(got.PublicURL, s.cfg.DomainPublic) // Mock returns specific string
// 			}
// 		})
// 	}
// }

// func (s *mediaServiceSuite) TestConfirmUpload() {
// 	var (
// 		userID    int64 = 1
// 		objectKey       = "users/1/avatar/image.jpg"
// 		input           = domain.ConfirmFileParams{
// 			EntityID:  userID,
// 			ObjectKey: objectKey,
// 			Domain:    domain.DomainUser,
// 			Feature:   domain.FeatureAvatar,
// 		}
// 	)

// 	type args struct {
// 		input []domain.ConfirmFileParams
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(storage *mocks.FileStorage)
// 		want      []string
// 		wantErr   error
// 	}{
// 		{
// 			name: "success",
// 			args: args{input: []domain.ConfirmFileParams{input}},
// 			setupMock: func(storage *mocks.FileStorage) {

// 				storage.EXPECT().StatFile(mock.Anything, domain.StorageLocation{
// 					Bucket: s.cfg.BucketPublic,
// 					Key:    objectKey,
// 				}).Return(true, nil).Once()

// 			},
// 			want:    []string{objectKey},
// 			wantErr: nil,
// 		},

// 		{
// 			name: "user_mismatch",
// 			args: args{input: []domain.ConfirmFileParams{{
// 				EntityID:  userID + 1,
// 				ObjectKey: objectKey,
// 				Domain:    domain.DomainUser,
// 				Feature:   domain.FeatureAvatar,
// 			}}},
// 			setupMock: func(storage *mocks.FileStorage) {

// 			},
// 			want:    nil,
// 			wantErr: pkg.ErrForbidden,
// 		},
// 		{
// 			name: "storage_stat_error",
// 			args: args{input: []domain.ConfirmFileParams{input}},
// 			setupMock: func(storage *mocks.FileStorage) {

// 				storage.EXPECT().StatFile(mock.Anything, mock.Anything).Return(false, assert.AnError).Once()
// 			},
// 			want:    nil,
// 			wantErr: assert.AnError,
// 		},
// 		{
// 			name: "file_not_found",
// 			args: args{input: []domain.ConfirmFileParams{input}},
// 			setupMock: func(storage *mocks.FileStorage) {
// 				storage.EXPECT().StatFile(mock.Anything, mock.Anything).Return(false, nil).Once()
// 			},
// 			want:    nil,
// 			wantErr: pkg.ErrNotFound,
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			mockStorage := mocks.NewFileStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			svc := NewMediaService(mockStorage, s.cfg, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(mockStorage)
// 			}

// 			got, err := svc.ConfirmUpload(context.Background(), tc.args.input)

// 			if tc.wantErr != nil {
// 				s.Error(err)
// 				if tc.wantErr != assert.AnError {
// 					s.ErrorIs(err, tc.wantErr)
// 				}
// 			} else {
// 				s.NoError(err)
// 				s.Equal(tc.want, got)
// 			}
// 		})
// 	}
// }

// func (s *mediaServiceSuite) TestDeleteFile() {
// 	objectKeys := []string{"file.jpg"}

// 	tests := []struct {
// 		name       string
// 		objectKeys []string
// 		setupMock  func(storage *mocks.FileStorage)
// 		wantErr    error
// 	}{
// 		{
// 			name:       "success",
// 			objectKeys: objectKeys,
// 			setupMock: func(storage *mocks.FileStorage) {
// 				storage.EXPECT().DeleteFile(mock.Anything, domain.StorageLocation{
// 					Bucket: s.cfg.BucketPublic,
// 					Key:    objectKeys[0],
// 				}).Return(nil).Once()
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:       "error",
// 			objectKeys: objectKeys,
// 			setupMock: func(storage *mocks.FileStorage) {
// 				storage.EXPECT().DeleteFile(mock.Anything, mock.Anything).Return(assert.AnError).Once()
// 			},
// 			wantErr: assert.AnError,
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			mockStorage := mocks.NewFileStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			svc := NewMediaService(mockStorage, s.cfg, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(mockStorage)
// 			}

// 			err := svc.DeleteFile(context.Background(), tc.objectKeys)

// 			if tc.wantErr != nil {
// 				s.ErrorIs(err, tc.wantErr)
// 			} else {
// 				s.NoError(err)
// 			}
// 		})
// 	}
// }

// func (s *mediaServiceSuite) TestVerifyMedia() {
// 	objectKeys := []string{"key1", "key2"}

// 	tests := []struct {
// 		name       string
// 		objectKeys []string
// 		setupMock  func(storage *mocks.FileStorage)
// 		wantErr    error
// 	}{
// 		{
// 			name:       "success",
// 			objectKeys: objectKeys,
// 			setupMock: func(storage *mocks.FileStorage) {
// 				storage.EXPECT().StatFile(mock.Anything, domain.StorageLocation{
// 					Bucket: s.cfg.BucketPublic,
// 					Key:    objectKeys[0],
// 				}).Return(true, nil).Once()
// 				storage.EXPECT().StatFile(mock.Anything, domain.StorageLocation{
// 					Bucket: s.cfg.BucketPublic,
// 					Key:    objectKeys[1],
// 				}).Return(true, nil).Once()
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:       "storage_error",
// 			objectKeys: objectKeys,
// 			setupMock: func(storage *mocks.FileStorage) {
// 				storage.EXPECT().StatFile(mock.Anything, mock.Anything).Return(false, assert.AnError).Once()
// 			},
// 			wantErr: assert.AnError,
// 		},
// 		{
// 			name:       "not_found",
// 			objectKeys: objectKeys,
// 			setupMock: func(storage *mocks.FileStorage) {
// 				storage.EXPECT().StatFile(mock.Anything, mock.Anything).Return(false, nil).Once()
// 			},
// 			wantErr: pkg.ErrNotFound,
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			mockStorage := mocks.NewFileStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			svc := NewMediaService(mockStorage, s.cfg, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(mockStorage)
// 			}

// 			err := svc.VerifyMedia(context.Background(), tc.objectKeys)

// 			if tc.wantErr != nil {
// 				s.ErrorIs(err, tc.wantErr)
// 			} else {
// 				s.NoError(err)
// 			}
// 		})
// 	}
// }
