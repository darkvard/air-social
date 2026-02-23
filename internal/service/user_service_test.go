package service

// import (
// 	"context"
// 	"testing"
// 	"time"
//
//
//

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/suite"

// 	"air-social/internal/domain"
// 	"air-social/internal/mocks"
// 	"air-social/pkg"
// )

// type userServiceSuite struct {
// 	suite.Suite
// }

// func TestUserServiceSuite(t *testing.T) {
// 	suite.Run(t, new(userServiceSuite))
// }

// func (s *userServiceSuite) TestCreateUser() {
// 	baseInput := domain.CreateUserParams{
// 		Email:          "email@example.com",
// 		Username:       "test",
// 		PasswordHashed: "hash",
// 	}

// 	type args struct {
// 		input domain.CreateUserParams
// 	}

// 	type want struct {
// 		response *domain.User
// 		err      error
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args)
// 		want      want
// 	}{
// 		{
// 			name: "repo_create_error",
// 			args: args{
// 				input: baseInput,
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().
// 					Create(mock.Anything, mock.Anything).
// 					Return(assert.AnError).
// 					Once()
// 			},
// 			want: want{
// 				err:      pkg.ErrInternal,
// 				response: nil,
// 			},
// 		},
// 		{
// 			name: "success",
// 			args: args{
// 				input: baseInput,
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().
// 					Create(
// 						mock.Anything,
// 						mock.MatchedBy(func(u *domain.User) bool {
// 							return u.Email == a.input.Email &&
// 								u.Username == a.input.Username &&
// 								u.PasswordHash == a.input.PasswordHashed
// 						}),
// 					).
// 					Return(nil).
// 					Once()
// 			},
// 			want: want{
// 				err: nil,
// 				response: &domain.User{
// 					Email:        baseInput.Email,
// 					Username:     baseInput.Username,
// 					PasswordHash: baseInput.PasswordHashed,
// 				},
// 			},
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			mockRepo := mocks.NewUserRepository(s.T())
// 			mockMediaMgr := mocks.NewUserMediaManager(s.T())
// 			mockCache := mocks.NewCacheStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			userSvc := NewUserService(mockRepo, mockMediaMgr, mockCache, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(mockRepo, mockMediaMgr, mockCache, mockURL, tc.args)
// 			}

// 			got, err := userSvc.CreateUser(context.Background(), tc.args.input)

// 			if tc.want.err != nil {
// 				s.ErrorIs(err, tc.want.err)
// 				s.Empty(got)
// 			} else {
// 				s.NoError(err)
// 				s.Equal(tc.want.response, got)
// 			}
// 		})
// 	}
// }

// func (s *userServiceSuite) TestGetByID() {
// 	expectedUser := &domain.User{
// 		ID:     1,
// 		Email:  "email@example.com",
// 		Status: domain.UserStatus{Verified: false},
// 	}

// 	type args struct {
// 		ctx context.Context
// 		id  int64
// 	}

// 	type want struct {
// 		user *domain.User
// 		err  error
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args)
// 		want      want
// 	}{
// 		{
// 			name: "error_internal",
// 			args: args{
// 				id: 3,
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByID(mock.Anything, a.id).Return(nil, pkg.ErrInternal).Once()
// 			},
// 			want: want{
// 				user: nil,
// 				err:  pkg.ErrInternal,
// 			},
// 		},
// 		{
// 			name: "error_notfound",
// 			args: args{
// 				id: 2,
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByID(mock.Anything, a.id).Return(nil, pkg.ErrNotFound).Once()
// 			},
// 			want: want{
// 				user: nil,
// 				err:  pkg.ErrNotFound,
// 			},
// 		},
// 		{
// 			name: "success",
// 			args: args{
// 				id: expectedUser.ID,
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByID(mock.Anything, a.id).Return(expectedUser, nil).Once()
// 			},
// 			want: want{
// 				user: expectedUser,
// 				err:  nil,
// 			},
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			userRepo := mocks.NewUserRepository(s.T())
// 			mockMediaMgr := mocks.NewUserMediaManager(s.T())
// 			mockCache := mocks.NewCacheStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			userSvc := NewUserService(userRepo, mockMediaMgr, mockCache, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(userRepo, mockMediaMgr, mockCache, mockURL, tc.args)
// 			}
// 			got, err := userSvc.GetByID(context.Background(), tc.args.id)

// 			if tc.want.err != nil {
// 				s.ErrorIs(err, tc.want.err)
// 				s.Nil(got)
// 			} else {
// 				s.NoError(err)
// 				s.Equal(tc.want.user, got)
// 			}
// 		})
// 	}
// }

// func (s *userServiceSuite) TestGetByEmail() {
// 	expectedUser := &domain.User{
// 		ID:     1,
// 		Email:  "email@example.com",
// 		Status: domain.UserStatus{Verified: false},
// 	}

// 	type args struct {
// 		email string
// 	}

// 	type want struct {
// 		user *domain.User
// 		err  error
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args)
// 		want      want
// 	}{
// 		{
// 			name: "error_internal",
// 			args: args{
// 				email: "error@example.com",
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(nil, pkg.ErrInternal).Once()
// 			},
// 			want: want{
// 				user: nil,
// 				err:  pkg.ErrInternal,
// 			},
// 		},
// 		{
// 			name: "error_notfound",
// 			args: args{
// 				email: "notfound@example.com",
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(nil, pkg.ErrNotFound).Once()
// 			},
// 			want: want{
// 				user: nil,
// 				err:  pkg.ErrNotFound,
// 			},
// 		},
// 		{
// 			name: "success",
// 			args: args{
// 				email: expectedUser.Email,
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(expectedUser, nil).Once()
// 			},
// 			want: want{
// 				user: expectedUser,
// 				err:  nil,
// 			},
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			userRepo := mocks.NewUserRepository(s.T())
// 			mockMediaMgr := mocks.NewUserMediaManager(s.T())
// 			mockCache := mocks.NewCacheStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			userSvc := NewUserService(userRepo, mockMediaMgr, mockCache, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(userRepo, mockMediaMgr, mockCache, mockURL, tc.args)
// 			}
// 			got, err := userSvc.GetByEmail(context.Background(), tc.args.email)

// 			if tc.want.err != nil {
// 				s.ErrorIs(err, tc.want.err)
// 				s.Nil(got)
// 			} else {
// 				s.NoError(err)
// 				s.Equal(tc.want.user, got)
// 			}
// 		})
// 	}
// }

// func (s *userServiceSuite) TestGetUserSummary() {
// 	user := &domain.User{
// 		ID:     1,
// 		Email:  "email@example.com",
// 		Status: domain.UserStatus{Verified: false},
// 		Profile: domain.Profile{
// 			Avatar:     "user/1/avatar/ab12dgh31.jpg",
// 			CoverImage: "user/1/cover/oik98anc.png",
// 		},
// 	}
// 	publicInfo := &domain.UserSummary{
// 		ID:         user.ID,
// 		Avatar:     "http://cdn/user/1/avatar/ab12dgh31.jpg",
// 		CoverImage: "http://cdn/user/1/cover/oik98anc.png",
// 	}
// 	cacheKey := domain.GetUserSummaryKey(user.ID)

// 	type args struct {
// 		id int64
// 	}

// 	type want struct {
// 		info *domain.UserSummary
// 		err  error
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args)
// 		want      want
// 	}{
// 		{
// 			name: "cache_miss_repo_error",
// 			args: args{
// 				id: 3,
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				cache.EXPECT().Get(mock.Anything, domain.GetUserSummaryKey(a.id), mock.Anything).Return(pkg.ErrNotFound).Once()
// 				userRepo.EXPECT().GetByID(mock.Anything, a.id).Return(nil, assert.AnError).Once()
// 			},
// 			want: want{
// 				info: nil,
// 				err:  assert.AnError,
// 			},
// 		},
// 		{
// 			name: "cache_miss_success",
// 			args: args{
// 				id: user.ID,
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				cache.EXPECT().Get(mock.Anything, cacheKey, mock.Anything).Return(pkg.ErrNotFound).Once()
// 				userRepo.EXPECT().GetByID(mock.Anything, a.id).Return(user, nil).Once()
// 				url.EXPECT().PublicFileURL(user.Profile.Avatar).Return(publicInfo.Avatar).Once()
// 				url.EXPECT().PublicFileURL(user.Profile.CoverImage).Return(publicInfo.CoverImage).Once()
// 				cache.EXPECT().Set(mock.Anything, cacheKey, mock.Anything, 12*time.Hour).Return(nil).Once()
// 			},
// 			want: want{
// 				info: publicInfo,
// 				err:  nil,
// 			},
// 		},
// 		{
// 			name: "cache_hit_success",
// 			args: args{id: user.ID},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				cache.EXPECT().Get(mock.Anything, cacheKey, mock.Anything).
// 					Run(func(_ context.Context, _ string, dest any) { *dest.(*domain.UserSummary) = *publicInfo }).Return(nil).Once()
// 			},
// 			want: want{
// 				info: publicInfo,
// 				err:  nil,
// 			},
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			userRepo := mocks.NewUserRepository(s.T())
// 			mockMediaMgr := mocks.NewUserMediaManager(s.T())
// 			mockCache := mocks.NewCacheStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			userSvc := NewUserService(userRepo, mockMediaMgr, mockCache, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(userRepo, mockMediaMgr, mockCache, mockURL, tc.args)
// 			}
// 			got, err := userSvc.GetSummary(context.Background(), tc.args.id)

// 			if tc.want.err != nil {
// 				s.Error(err)
// 			} else {
// 				s.NoError(err)
// 				s.Equal(tc.want.info.ID, got.ID)
// 				s.Equal(tc.want.info.Avatar, got.Avatar)
// 			}
// 		})
// 	}
// }

// func (s *userServiceSuite) TestUpdateProfile() {
// 	var (
// 		userID   int64 = 1
// 		fullName       = "New Name"
// 		bio            = "New Bio"
// 	)

// 	baseInput := domain.UpdateProfileParams{
// 		UserID:   userID,
// 		FullName: &fullName,
// 		Bio:      &bio,
// 	}

// 	existingUser := &domain.User{
// 		ID: userID,
// 		Profile: domain.Profile{
// 			FullName: "Old Name",
// 			Bio:      "Old Bio",
// 		},
// 	}

// 	type args struct {
// 		input domain.UpdateProfileParams
// 	}

// 	type want struct {
// 		response *domain.User
// 		err      error
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args)
// 		want      want
// 	}{
// 		{
// 			name: "get_user_error",
// 			args: args{input: baseInput},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(nil, pkg.ErrNotFound).Once()
// 			},
// 			want: want{
// 				err: pkg.ErrNotFound,
// 			},
// 		},
// 		{
// 			name: "update_error",
// 			args: args{input: baseInput},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(existingUser, nil).Once()

// 				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
// 					return u.Profile.FullName == *a.input.FullName && u.Profile.Bio == *a.input.Bio
// 				})).Return(assert.AnError).Once()
// 			},
// 			want: want{
// 				err: pkg.ErrInternal,
// 			},
// 		},
// 		{
// 			name: "success",
// 			args: args{input: baseInput},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(existingUser, nil).Once()

// 				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
// 					return u.Profile.FullName == *a.input.FullName && u.Profile.Bio == *a.input.Bio
// 				})).Return(nil).Once()
// 				cache.EXPECT().Delete(mock.Anything, domain.GetUserSummaryKey(userID)).Return(nil).Once()
// 			},
// 			want: want{
// 				response: &domain.User{
// 					ID: userID,
// 					Profile: domain.Profile{
// 						FullName: fullName,
// 						Bio:      bio,
// 					},
// 				},
// 				err: nil,
// 			},
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			userRepo := mocks.NewUserRepository(s.T())
// 			mockMediaMgr := mocks.NewUserMediaManager(s.T())
// 			mockCache := mocks.NewCacheStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			userSvc := NewUserService(userRepo, mockMediaMgr, mockCache, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(userRepo, mockMediaMgr, mockCache, mockURL, tc.args)
// 			}

// 			got, err := userSvc.UpdateProfile(context.Background(), tc.args.input)

// 			if tc.want.err != nil {
// 				s.ErrorIs(err, tc.want.err)
// 			} else {
// 				s.NoError(err)
// 				s.Equal(tc.want.response.Profile.FullName, got.Profile.FullName)
// 				s.Equal(tc.want.response.Profile.Bio, got.Profile.Bio)
// 			}
// 		})
// 	}
// }

// func (s *userServiceSuite) TestChangePassword() {
// 	password := "password123"
// 	hashedPassword, _ := pkg.HashPassword(password)
// 	userID := int64(1)

// 	type args struct {
// 		input domain.ChangePasswordParams
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args)
// 		wantErr   error
// 	}{
// 		{
// 			name: "user_not_found",
// 			args: args{
// 				input: domain.ChangePasswordParams{UserID: userID},
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(nil, pkg.ErrNotFound).Once()
// 			},
// 			wantErr: pkg.ErrNotFound,
// 		},
// 		{
// 			name: "same_password",
// 			args: args{
// 				input: domain.ChangePasswordParams{
// 					UserID:          userID,
// 					CurrentPassword: password,
// 					NewPassword:     password,
// 				},
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(&domain.User{PasswordHash: hashedPassword}, nil).Once()
// 			},
// 			wantErr: pkg.ErrSamePassword,
// 		},
// 		{
// 			name: "invalid_current_password",
// 			args: args{
// 				input: domain.ChangePasswordParams{
// 					UserID:          userID,
// 					CurrentPassword: "wrongpassword",
// 					NewPassword:     "newpassword",
// 				},
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(&domain.User{PasswordHash: hashedPassword}, nil).Once()
// 			},
// 			wantErr: pkg.ErrInvalidCredentials,
// 		},
// 		{
// 			name: "success",
// 			args: args{
// 				input: domain.ChangePasswordParams{
// 					UserID:          userID,
// 					CurrentPassword: password,
// 					NewPassword:     "newpassword",
// 				},
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByID(mock.Anything, a.input.UserID).Return(&domain.User{PasswordHash: hashedPassword}, nil).Once()

// 				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
// 					return pkg.VerifyPassword(a.input.NewPassword, u.PasswordHash)
// 				})).Return(nil).Once()
// 			},
// 			wantErr: nil,
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			userRepo := mocks.NewUserRepository(s.T())
// 			mockMediaMgr := mocks.NewUserMediaManager(s.T())
// 			mockCache := mocks.NewCacheStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			userSvc := NewUserService(userRepo, mockMediaMgr, mockCache, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(userRepo, mockMediaMgr, mockCache, mockURL, tc.args)
// 			}

// 			err := userSvc.ChangePassword(context.Background(), tc.args.input)

// 			if tc.wantErr != nil {
// 				s.ErrorIs(err, tc.wantErr)
// 			} else {
// 				s.NoError(err)
// 			}
// 		})
// 	}
// }

// func (s *userServiceSuite) TestUpdatePassword() {
// 	email := "test@example.com"
// 	newHash := "newhash"

// 	type args struct {
// 		email          string
// 		passwordHashed string
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args)
// 		wantErr   error
// 	}{
// 		{
// 			name: "user_not_found",
// 			args: args{email: email, passwordHashed: newHash},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(nil, pkg.ErrNotFound).Once()
// 			},
// 			wantErr: pkg.ErrNotFound,
// 		},
// 		{
// 			name: "success",
// 			args: args{email: email, passwordHashed: newHash},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(&domain.User{Email: email}, nil).Once()

// 				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
// 					return u.PasswordHash == a.passwordHashed
// 				})).Return(nil).Once()
// 			},
// 			wantErr: nil,
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			userRepo := mocks.NewUserRepository(s.T())
// 			mockMediaMgr := mocks.NewUserMediaManager(s.T())
// 			mockCache := mocks.NewCacheStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			userSvc := NewUserService(userRepo, mockMediaMgr, mockCache, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(userRepo, mockMediaMgr, mockCache, mockURL, tc.args)
// 			}

// 			err := userSvc.UpdatePassword(context.Background(), tc.args.email, tc.args.passwordHashed)

// 			if tc.wantErr != nil {
// 				s.ErrorIs(err, tc.wantErr)
// 			} else {
// 				s.NoError(err)
// 			}
// 		})
// 	}
// }

// func (s *userServiceSuite) TestVerifyEmail() {
// 	email := "test@example.com"

// 	type args struct {
// 		email string
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args)
// 		wantErr   error
// 	}{
// 		{
// 			name: "user_not_found",
// 			args: args{email: email},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(nil, pkg.ErrNotFound).Once()
// 			},
// 			wantErr: pkg.ErrNotFound,
// 		},
// 		{
// 			name: "success",
// 			args: args{email: email},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				userRepo.EXPECT().GetByEmail(mock.Anything, a.email).Return(&domain.User{ID: 1, Email: email}, nil).Once()

// 				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
// 					return u.Status.Verified == true && u.Status.VerifiedAt != nil
// 				})).Return(nil).Once()
// 				cache.EXPECT().Delete(mock.Anything, domain.GetUserSummaryKey(1)).Return(nil).Once()
// 			},
// 			wantErr: nil,
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			userRepo := mocks.NewUserRepository(s.T())
// 			mockMediaMgr := mocks.NewUserMediaManager(s.T())
// 			mockCache := mocks.NewCacheStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			userSvc := NewUserService(userRepo, mockMediaMgr, mockCache, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(userRepo, mockMediaMgr, mockCache, mockURL, tc.args)
// 			}

// 			err := userSvc.VerifyEmail(context.Background(), tc.args.email)

// 			if tc.wantErr != nil {
// 				s.ErrorIs(err, tc.wantErr)
// 			} else {
// 				s.NoError(err)
// 			}
// 		})
// 	}
// }

// func (s *userServiceSuite) TestConfirmImageUpload() {
// 	userID := int64(1)
// 	objectKey := "users/1/avatar/image.jpg"
// 	publicURL := "http://localhost/air-social-public/" + objectKey

// 	type args struct {
// 		input []domain.ConfirmFileParams
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args)
// 		want      []domain.ConfirmFileResult
// 		wantErr   error
// 	}{
// 		{
// 			name: "invalid_feature",
// 			args: args{
// 				input: []domain.ConfirmFileParams{{Feature: domain.FeatureFeedImage}},
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 			},
// 			want:    nil,
// 			wantErr: pkg.ErrInvalidData,
// 		},
// 		{
// 			name: "confirm_upload_error",
// 			args: args{
// 				input: []domain.ConfirmFileParams{{Feature: domain.FeatureAvatar, EntityID: userID, Domain: domain.DomainUser}},
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				mediaMgr.EXPECT().ConfirmUpload(mock.Anything, a.input).Return(nil, pkg.ErrNotFound).Once()
// 			},
// 			want:    nil,
// 			wantErr: pkg.ErrNotFound,
// 		},
// 		{
// 			name: "update_repo_error",
// 			args: args{
// 				input: []domain.ConfirmFileParams{{Feature: domain.FeatureAvatar, EntityID: userID, Domain: domain.DomainUser}},
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				mediaMgr.EXPECT().ConfirmUpload(mock.Anything, a.input).Return([]string{objectKey}, nil).Once()
// 				userRepo.EXPECT().UpdateProfileImages(mock.Anything, a.input[0].EntityID, objectKey, a.input[0].Feature).Return(assert.AnError).Once()
// 			},
// 			want:    nil,
// 			wantErr: pkg.ErrInternal,
// 		},
// 		{
// 			name: "success",
// 			args: args{
// 				input: []domain.ConfirmFileParams{{Feature: domain.FeatureAvatar, EntityID: userID, Domain: domain.DomainUser}},
// 			},
// 			setupMock: func(userRepo *mocks.UserRepository, mediaMgr *mocks.UserMediaManager, cache *mocks.CacheStorage, url *mocks.URLFactory, a args) {
// 				mediaMgr.EXPECT().ConfirmUpload(mock.Anything, a.input).Return([]string{objectKey}, nil).Once()
// 				userRepo.EXPECT().UpdateProfileImages(mock.Anything, a.input[0].EntityID, objectKey, a.input[0].Feature).Return(nil).Once()
// 				url.EXPECT().PublicFileURL(objectKey).Return(publicURL).Once()
// 				cache.EXPECT().Delete(mock.Anything, domain.GetUserSummaryKey(userID)).Return(nil).Once()
// 			},
// 			want: []domain.ConfirmFileResult{{
// 				Domain:  domain.DomainUser,
// 				Feature: domain.FeatureAvatar,
// 				URL:     publicURL,
// 			}},
// 			wantErr: nil,
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			userRepo := mocks.NewUserRepository(s.T())
// 			mockMediaMgr := mocks.NewUserMediaManager(s.T())
// 			mockCache := mocks.NewCacheStorage(s.T())
// 			mockURL := mocks.NewURLFactory(s.T())
// 			userSvc := NewUserService(userRepo, mockMediaMgr, mockCache, mockURL)

// 			if tc.setupMock != nil {
// 				tc.setupMock(userRepo, mockMediaMgr, mockCache, mockURL, tc.args)
// 			}

// 			got, err := userSvc.ConfirmImageUpload(context.Background(), tc.args.input)

// 			if tc.wantErr != nil {
// 				s.ErrorIs(err, tc.wantErr)
// 				s.Empty(got)
// 			} else {
// 				s.NoError(err)
// 				s.Equal(tc.want, got)
// 			}
// 		})
// 	}
// }
