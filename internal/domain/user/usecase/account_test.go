package usecase_test

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"

	"air-social/internal/domain/user"
	usermocks "air-social/internal/domain/user/mocks"
	"air-social/internal/domain/user/usecase"
	"air-social/pkg"
)

type accountUseCaseSuite struct {
	suite.Suite
}

func TestAccountUseCaseSuite(t *testing.T) {
	suite.Run(t, new(accountUseCaseSuite))
}

// mustHash generates a valid hash compatible with account.go logic for test fixtures
func mustHash(pwd string) string {
	sha := sha256.Sum256([]byte(pwd))
	hash, _ := bcrypt.GenerateFromPassword(sha[:], bcrypt.DefaultCost)
	return string(hash)
}

func (s *accountUseCaseSuite) TestVerifyEmail() {
	var (
		email  = "test@example.com"
		userID = int64(1)
	)

	existingUser := &user.User{
		ID:    userID,
		Email: email,
	}

	type testDeps struct {
		repo  *usermocks.MockRepository
		cache *usermocks.MockCache
	}

	type args struct {
		ctx   context.Context
		email string
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "user_not_found",
			args: args{ctx: context.Background(), email: email},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "update_failed",
			args: args{ctx: context.Background(), email: email},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(existingUser, nil).Once()

				deps.repo.EXPECT().
					UpdateVerified(mock.Anything, userID, true).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			args: args{ctx: context.Background(), email: email},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(existingUser, nil).Once()

				deps.repo.EXPECT().
					UpdateVerified(mock.Anything, userID, true).
					Return(nil).Once()

				deps.cache.EXPECT().
					Invalidate(mock.Anything, userID).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := usermocks.NewMockRepository(s.T())
			mockCache := usermocks.NewMockCache(s.T())

			deps := testDeps{repo: mockRepo, cache: mockCache}
			uc := usecase.NewAccountUseCase(usecase.AccountDeps{Repo: mockRepo, Cache: mockCache})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.VerifyEmail(tc.args.ctx, tc.args.email)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *accountUseCaseSuite) TestCreateUser() {
	input := user.CreateParams{
		Email:       "new@example.com",
		Username:    "newuser",
		NewPassword: "password123",
	}

	type testDeps struct {
		repo  *usermocks.MockRepository
		cache *usermocks.MockCache
	}

	type args struct {
		ctx    context.Context
		params user.CreateParams
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
			name: "repo_error",
			args: args{ctx: context.Background(), params: input},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(u *user.User) bool {
						return u.Email == input.Email && u.Username == input.Username && u.PasswordHash != ""
					})).
					Return(pkg.ErrAlreadyExists).Once()
			},
			want: want{
				user: nil,
				err:  pkg.ErrAlreadyExists, // Assuming OrInternalError maps generic error to this or Internal
			},
		},
		{
			name: "success",
			args: args{ctx: context.Background(), params: input},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(u *user.User) bool {
						return u.Email == input.Email && u.Username == input.Username && u.PasswordHash != ""
					})).
					Return(nil).Once()
			},
			want: want{
				user: &user.User{Email: input.Email, Username: input.Username},
				err:  nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := usermocks.NewMockRepository(s.T())
			mockCache := usermocks.NewMockCache(s.T())

			deps := testDeps{repo: mockRepo, cache: mockCache}
			uc := usecase.NewAccountUseCase(usecase.AccountDeps{Repo: mockRepo, Cache: mockCache})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.CreateUser(tc.args.ctx, tc.args.params)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.user.Email, got.Email)
				s.Equal(tc.want.user.Username, got.Username)
				s.NotEmpty(got.PasswordHash)
			}
		})
	}
}

func (s *accountUseCaseSuite) TestChangePassword() {
	var (
		userID          = int64(1)
		currentPassword = "oldPassword"
		newPassword     = "newPassword"
		hashedCurrent   = mustHash(currentPassword)
	)

	existingUser := &user.User{
		ID:           userID,
		PasswordHash: hashedCurrent,
	}

	type testDeps struct {
		repo  *usermocks.MockRepository
		cache *usermocks.MockCache
	}

	type args struct {
		ctx    context.Context
		params user.ChangePasswordParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "user_not_found",
			args: args{
				ctx: context.Background(),
				params: user.ChangePasswordParams{
					UserID: userID,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrInvalidCredentials,
		},
		{
			name: "same_password",
			args: args{
				ctx: context.Background(),
				params: user.ChangePasswordParams{
					UserID:          userID,
					CurrentPassword: currentPassword,
					NewPassword:     currentPassword,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(existingUser, nil).Once()
			},
			wantErr: pkg.ErrSamePassword,
		},
		{
			name: "invalid_current_password",
			args: args{
				ctx: context.Background(),
				params: user.ChangePasswordParams{
					UserID:          userID,
					CurrentPassword: "wrongPassword",
					NewPassword:     newPassword,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(existingUser, nil).Once()
			},
			wantErr: pkg.ErrInvalidCredentials,
		},
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				params: user.ChangePasswordParams{
					UserID:          userID,
					CurrentPassword: currentPassword,
					NewPassword:     newPassword,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, userID).
					Return(existingUser, nil).Once()

				deps.repo.EXPECT().
					UpdatePassword(mock.Anything, userID, mock.AnythingOfType("string")).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := usermocks.NewMockRepository(s.T())
			mockCache := usermocks.NewMockCache(s.T())

			deps := testDeps{repo: mockRepo, cache: mockCache}
			uc := usecase.NewAccountUseCase(usecase.AccountDeps{Repo: mockRepo, Cache: mockCache})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.ChangePassword(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *accountUseCaseSuite) TestAuthenticate() {
	var (
		email    = "test@example.com"
		password = "password123"
		hashed   = mustHash(password)
	)

	existingUser := &user.User{
		ID:           1,
		Email:        email,
		PasswordHash: hashed,
	}

	type testDeps struct {
		repo  *usermocks.MockRepository
		cache *usermocks.MockCache
	}

	type args struct {
		ctx    context.Context
		params user.AuthenticateParams
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
			name: "user_not_found",
			args: args{
				ctx:    context.Background(),
				params: user.AuthenticateParams{Email: email, Password: password},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(nil, pkg.ErrNotFound).Once()
			},
			want: want{
				user: nil,
				err:  pkg.ErrInvalidCredentials,
			},
		},
		{
			name: "invalid_password",
			args: args{
				ctx:    context.Background(),
				params: user.AuthenticateParams{Email: email, Password: "wrong"},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(existingUser, nil).Once()
			},
			want: want{
				user: nil,
				err:  pkg.ErrInvalidCredentials,
			},
		},
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				params: user.AuthenticateParams{Email: email, Password: password},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(existingUser, nil).Once()
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
			mockCache := usermocks.NewMockCache(s.T())

			deps := testDeps{repo: mockRepo, cache: mockCache}
			uc := usecase.NewAccountUseCase(usecase.AccountDeps{Repo: mockRepo, Cache: mockCache})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.Authenticate(tc.args.ctx, tc.args.params)

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
