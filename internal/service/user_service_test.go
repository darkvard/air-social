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

type userServiceSuite struct {
	suite.Suite
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(userServiceSuite))
}

func (s *userServiceSuite) TestCreateUser() {
	baseInput := domain.CreateUserParams{
		Email:          "email@example.com",
		Username:       "test",
		PasswordHashed: "hash",
	}

	type args struct {
		input domain.CreateUserParams
	}

	type want struct {
		response domain.UserResponse
		err      error
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args)
		want      want
	}{
		{
			name: "repo_get_error",
			args: args{
				input: baseInput,
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().
					GetByEmail(mock.Anything, a.input.Email).
					Return(nil, assert.AnError).
					Once()
			},
			want: want{
				err:      pkg.ErrInternal,
				response: domain.UserResponse{},
			},
		},
		{
			name: "repo_get_exists",
			args: args{
				input: baseInput,
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				foundUser := &domain.User{Email: a.input.Email}
				userRepo.EXPECT().
					GetByEmail(mock.Anything, a.input.Email).
					Return(foundUser, nil).
					Once()
			},
			want: want{
				err:      pkg.ErrAlreadyExists,
				response: domain.UserResponse{},
			},
		},
		{
			name: "success",
			args: args{
				input: baseInput,
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().
					GetByEmail(mock.Anything, a.input.Email).
					Return(nil, pkg.ErrNotFound).
					Once()

				userRepo.EXPECT().
					Create(
						mock.Anything,
						mock.MatchedBy(func(u *domain.User) bool {
							return u.Email == a.input.Email &&
								u.Username == a.input.Username &&
								u.PasswordHash == a.input.PasswordHashed
						}),
					).
					Return(nil).
					Once()
			},
			want: want{
				err: nil,
				response: domain.UserResponse{
					Email:    baseInput.Email,
					Username: baseInput.Username,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewUserRepository(s.T())
			mockMedia := mocks.NewMediaService(s.T())
			userSvc := NewUserService(mockRepo, mockMedia)

			if tc.setupMock != nil {
				tc.setupMock(mockRepo, mockMedia, tc.args)
			}

			got, err := userSvc.CreateUser(context.Background(), tc.args.input)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Empty(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.response, got)
			}
		})
	}
}

func (s *userServiceSuite) TestGetByID() {
	expectedUser := &domain.User{
		ID:       1,
		Email:    "email@example.com",
		Verified: false,
	}

	type args struct {
		ctx context.Context
		id  int64
	}

	type want struct {
		user *domain.User
		err  error
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args)
		want      want
	}{
		{
			name: "error_internal",
			args: args{
				id: 3,
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.id).Return(nil, pkg.ErrInternal).Once()
			},
			want: want{
				user: nil,
				err:  pkg.ErrInternal,
			},
		},
		{
			name: "error_notfound",
			args: args{
				id: 2,
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.id).Return(nil, pkg.ErrNotFound).Once()
			},
			want: want{
				user: nil,
				err:  pkg.ErrNotFound,
			},
		},
		{
			name: "success",
			args: args{
				id: expectedUser.ID,
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.id).Return(expectedUser, nil).Once()
			},
			want: want{
				user: expectedUser,
				err:  nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			userRepo := mocks.NewUserRepository(s.T())
			mediaSvc := mocks.NewMediaService(s.T())
			userSvc := NewUserService(userRepo, mediaSvc)

			if tc.setupMock != nil {
				tc.setupMock(userRepo, mediaSvc, tc.args)
			}
			got, err := userSvc.GetByID(context.Background(), tc.args.id)

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

func (s *userServiceSuite) TestGetByEmail() {
	expectedUser := &domain.User{
		ID:       1,
		Email:    "email@example.com",
		Verified: false,
	}

	type args struct {
		email string
	}

	type want struct {
		user *domain.User
		err  error
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args)
		want      want
	}{
		{
			name: "error_internal",
			args: args{
				email: "error@example.com",
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(nil, pkg.ErrInternal).Once()
			},
			want: want{
				user: nil,
				err:  pkg.ErrInternal,
			},
		},
		{
			name: "error_notfound",
			args: args{
				email: "notfound@example.com",
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(nil, pkg.ErrNotFound).Once()
			},
			want: want{
				user: nil,
				err:  pkg.ErrNotFound,
			},
		},
		{
			name: "success",
			args: args{
				email: expectedUser.Email,
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(expectedUser, nil).Once()
			},
			want: want{
				user: expectedUser,
				err:  nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			userRepo := mocks.NewUserRepository(s.T())
			mediaSvc := mocks.NewMediaService(s.T())
			userSvc := NewUserService(userRepo, mediaSvc)

			if tc.setupMock != nil {
				tc.setupMock(userRepo, mediaSvc, tc.args)
			}
			got, err := userSvc.GetByEmail(context.Background(), tc.args.email)

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

func (s *userServiceSuite) TestGetProfile() {
	user := &domain.User{
		ID:       1,
		Email:    "email@example.com",
		Verified: false,
		Profile: domain.Profile{
			Avatar:     "user/1/avatar/ab12dgh31.jpg",
			CoverImage: "user/1/cover/oik98anc.png",
		},
	}

	prefixDomainPublic := "http://localhost/air-social-public/"
	userResponse := user.ToResponse()
	userResponse.Avatar = prefixDomainPublic + userResponse.Avatar
	userResponse.CoverImage = prefixDomainPublic + userResponse.CoverImage

	type args struct {
		id int64
	}

	type want struct {
		user domain.UserResponse
		err  error
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args)
		want      want
	}{
		{
			name: "error",
			args: args{
				id: 3,
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.id).Return(nil, assert.AnError).Once()
			},
			want: want{
				user: domain.UserResponse{},
				err:  assert.AnError,
			},
		},
		{
			name: "success",
			args: args{
				id: userResponse.ID,
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.id).Return(user, nil).Once()
				mediaSvc.EXPECT().GetPublicURL(mock.Anything).
					RunAndReturn(func(objectKey string) string {
						switch objectKey {
						case user.Avatar:
							return userResponse.Avatar
						case user.CoverImage:
							return userResponse.CoverImage
						default:
							return ""
						}
					}).
					Twice()
			},
			want: want{
				user: userResponse,
				err:  nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			userRepo := mocks.NewUserRepository(s.T())
			mediaSvc := mocks.NewMediaService(s.T())
			userSvc := NewUserService(userRepo, mediaSvc)

			if tc.setupMock != nil {
				tc.setupMock(userRepo, mediaSvc, tc.args)
			}
			got, err := userSvc.GetProfile(context.Background(), tc.args.id)

			if tc.want.err != nil {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tc.want.user, got)
			}
		})
	}
}

func (s *userServiceSuite) TestResolveMediaURLs() {
	avatarKey := "avatar.jpg"
	coverKey := "cover.jpg"
	publicAvatar := "http://cdn/avatar.jpg"
	publicCover := "http://cdn/cover.jpg"

	type args struct {
		res *domain.UserResponse
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(mediaSvc *mocks.MediaService, a args)
		want      *domain.UserResponse
	}{
		{
			name: "nil_response",
			args: args{res: nil},
			want: nil,
		},
		{
			name: "empty_images",
			args: args{res: &domain.UserResponse{}},
			want: &domain.UserResponse{},
		},
		{
			name: "resolve_images",
			args: args{
				res: &domain.UserResponse{
					Profile: domain.Profile{
						Avatar:     avatarKey,
						CoverImage: coverKey,
					},
				},
			},
			setupMock: func(mediaSvc *mocks.MediaService, a args) {
				mediaSvc.EXPECT().GetPublicURL(avatarKey).Return(publicAvatar).Once()
				mediaSvc.EXPECT().GetPublicURL(coverKey).Return(publicCover).Once()
			},
			want: &domain.UserResponse{
				Profile: domain.Profile{
					Avatar:     publicAvatar,
					CoverImage: publicCover,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mediaSvc := mocks.NewMediaService(s.T())
			userSvc := NewUserService(nil, mediaSvc)

			if tc.setupMock != nil {
				tc.setupMock(mediaSvc, tc.args)
			}

			userSvc.ResolveMediaURLs(tc.args.res)

			if tc.want != nil {
				s.Equal(tc.want, tc.args.res)
			}
		})
	}
}

func (s *userServiceSuite) TestUpdateProfile() {
	var (
		userID   int64 = 1
		fullName       = "New Name"
		bio            = "New Bio"
	)

	baseInput := domain.UpdateProfileParams{
		UserID:   userID,
		FullName: &fullName,
		Bio:      &bio,
	}

	existingUser := &domain.User{
		ID: userID,
		Profile: domain.Profile{
			FullName: "Old Name",
			Bio:      "Old Bio",
		},
	}

	type args struct {
		input domain.UpdateProfileParams
	}

	type want struct {
		response domain.UserResponse
		err      error
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args)
		want      want
	}{
		{
			name: "get_user_error",
			args: args{input: baseInput},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(nil, pkg.ErrNotFound).Once()
			},
			want: want{
				err: pkg.ErrNotFound,
			},
		},
		{
			name: "update_error",
			args: args{input: baseInput},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(existingUser, nil).Once()

				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.FullName == *a.input.FullName && u.Bio == *a.input.Bio
				})).Return(assert.AnError).Once()
			},
			want: want{
				err: pkg.ErrInternal,
			},
		},
		{
			name: "success",
			args: args{input: baseInput},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(existingUser, nil).Once()

				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.FullName == *a.input.FullName && u.Bio == *a.input.Bio
				})).Return(nil).Once()
			},
			want: want{
				response: domain.UserResponse{
					ID: userID,
					Profile: domain.Profile{
						FullName: fullName,
						Bio:      bio,
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			userRepo := mocks.NewUserRepository(s.T())
			mediaSvc := mocks.NewMediaService(s.T())
			userSvc := NewUserService(userRepo, mediaSvc)

			if tc.setupMock != nil {
				tc.setupMock(userRepo, mediaSvc, tc.args)
			}

			got, err := userSvc.UpdateProfile(context.Background(), tc.args.input)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
			} else {
				s.NoError(err)
				s.Equal(tc.want.response.FullName, got.FullName)
				s.Equal(tc.want.response.Bio, got.Bio)
			}
		})
	}
}

func (s *userServiceSuite) TestChangePassword() {
	password := "password123"
	hashedPassword, _ := hashPassword(password)
	userID := int64(1)

	type args struct {
		input domain.ChangePasswordParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args)
		wantErr   error
	}{
		{
			name: "user_not_found",
			args: args{
				input: domain.ChangePasswordParams{UserID: userID},
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "same_password",
			args: args{
				input: domain.ChangePasswordParams{
					UserID:          userID,
					CurrentPassword: password,
					NewPassword:     password,
				},
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(&domain.User{PasswordHash: hashedPassword}, nil).Once()
			},
			wantErr: pkg.ErrSamePassword,
		},
		{
			name: "invalid_current_password",
			args: args{
				input: domain.ChangePasswordParams{
					UserID:          userID,
					CurrentPassword: "wrongpassword",
					NewPassword:     "newpassword",
				},
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(&domain.User{PasswordHash: hashedPassword}, nil).Once()
			},
			wantErr: pkg.ErrInvalidCredentials,
		},
		{
			name: "success",
			args: args{
				input: domain.ChangePasswordParams{
					UserID:          userID,
					CurrentPassword: password,
					NewPassword:     "newpassword",
				},
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(&domain.User{PasswordHash: hashedPassword}, nil).Once()
				
				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return verifyPassword(a.input.NewPassword, u.PasswordHash)
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			userRepo := mocks.NewUserRepository(s.T())
			mediaSvc := mocks.NewMediaService(s.T())
			userSvc := NewUserService(userRepo, mediaSvc)

			if tc.setupMock != nil {
				tc.setupMock(userRepo, mediaSvc, tc.args)
			}

			err := userSvc.ChangePassword(context.Background(), tc.args.input)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *userServiceSuite) TestUpdatePassword() {
	email := "test@example.com"
	newHash := "newhash"

	type args struct {
		email          string
		passwordHashed string
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args)
		wantErr   error
	}{
		{
			name: "user_not_found",
			args: args{email: email, passwordHashed: newHash},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "success",
			args: args{email: email, passwordHashed: newHash},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(&domain.User{Email: email}, nil).Once()
				
				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.PasswordHash == a.passwordHashed
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			userRepo := mocks.NewUserRepository(s.T())
			mediaSvc := mocks.NewMediaService(s.T())
			userSvc := NewUserService(userRepo, mediaSvc)

			if tc.setupMock != nil {
				tc.setupMock(userRepo, mediaSvc, tc.args)
			}

			err := userSvc.UpdatePassword(context.Background(), tc.args.email, tc.args.passwordHashed)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *userServiceSuite) TestVerifyEmail() {
	email := "test@example.com"

	type args struct {
		email string
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args)
		wantErr   error
	}{
		{
			name: "user_not_found",
			args: args{email: email},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "success",
			args: args{email: email},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(&domain.User{Email: email}, nil).Once()

				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.Verified == true && u.VerifiedAt != nil
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			userRepo := mocks.NewUserRepository(s.T())
			mediaSvc := mocks.NewMediaService(s.T())
			userSvc := NewUserService(userRepo, mediaSvc)

			if tc.setupMock != nil {
				tc.setupMock(userRepo, mediaSvc, tc.args)
			}

			err := userSvc.VerifyEmail(context.Background(), tc.args.email)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *userServiceSuite) TestConfirmImageUpload() {
	userID := int64(1)
	objectKey := "user/1/avatar/image.jpg"
	publicURL := "http://localhost/air-social-public/" + objectKey

	type args struct {
		input domain.ConfirmFileParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args)
		wantURL   string
		wantErr   error
	}{
		{
			name: "invalid_feature",
			args: args{
				input: domain.ConfirmFileParams{Feature: domain.FeatureFeedImage},
			},
			setupMock: nil,
			wantURL:   "",
			wantErr:   pkg.ErrInvalidData,
		},
		{
			name: "confirm_upload_error",
			args: args{
				input: domain.ConfirmFileParams{Feature: domain.FeatureAvatar, UserID: userID},
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				mediaSvc.EXPECT().ConfirmUpload(mock.Anything, a.input).Return("", pkg.ErrNotFound).Once()
			},
			wantURL: "",
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "update_repo_error",
			args: args{
				input: domain.ConfirmFileParams{Feature: domain.FeatureAvatar, UserID: userID},
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				mediaSvc.EXPECT().ConfirmUpload(mock.Anything, a.input).Return(objectKey, nil).Once()
				userRepo.EXPECT().UpdateProfileImages(mock.Anything, a.input.UserID, objectKey, a.input.Feature).Return(assert.AnError).Once()
			},
			wantURL: "",
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			args: args{
				input: domain.ConfirmFileParams{Feature: domain.FeatureAvatar, UserID: userID},
			},
			setupMock: func(userRepo *mocks.UserRepository, mediaSvc *mocks.MediaService, a args) {
				mediaSvc.EXPECT().ConfirmUpload(mock.Anything, a.input).Return(objectKey, nil).Once()
				userRepo.EXPECT().UpdateProfileImages(mock.Anything, a.input.UserID, objectKey, a.input.Feature).Return(nil).Once()
				mediaSvc.EXPECT().GetPublicURL(objectKey).Return(publicURL).Once()
			},
			wantURL: publicURL,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			userRepo := mocks.NewUserRepository(s.T())
			mediaSvc := mocks.NewMediaService(s.T())
			userSvc := NewUserService(userRepo, mediaSvc)

			if tc.setupMock != nil {
				tc.setupMock(userRepo, mediaSvc, tc.args)
			}

			got, err := userSvc.ConfirmImageUpload(context.Background(), tc.args.input)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(got)
			} else {
				s.NoError(err)
				s.Equal(tc.wantURL, got)
			}
		})
	}
}
