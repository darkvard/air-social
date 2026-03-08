package verify_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/auth/verify"
	"air-social/internal/domain/common"
	commonmocks "air-social/internal/domain/common/mocks"
	"air-social/pkg"
)

type verifyProviderSuite struct {
	suite.Suite
}

func TestVerifyProviderSuite(t *testing.T) {
	suite.Run(t, new(verifyProviderSuite))
}

func (s *verifyProviderSuite) TestSendVerification() {
	var (
		email    = "test@example.com"
		username = "testuser"
		link     = "http://verify.link"
	)

	type testDeps struct {
		cache *commonmocks.MockCache
		event *commonmocks.MockEventPublisher
		link  *commonmocks.MockLinkProvider
	}

	type args struct {
		ctx      context.Context
		email    string
		username string
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "cache_error",
			args: args{ctx: context.Background(), email: email, username: username},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					Set(mock.Anything, mock.Anything, email, 30*time.Minute).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			args: args{ctx: context.Background(), email: email, username: username},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					Set(mock.Anything, mock.Anything, email, 30*time.Minute).
					Return(nil).Once()

				deps.link.EXPECT().
					VerifyEmail(mock.Anything).
					Return(link).Once()

				deps.event.EXPECT().
					Publish(mock.Anything, mock.MatchedBy(func(e common.Event) bool {
						payload, ok := e.Data.(common.EmailEventPayload)
						return ok && payload.Email == email && payload.Link == link && e.Typ == common.EventEmailVerify
					})).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := commonmocks.NewMockCache(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())

			deps := testDeps{
				cache: mockCache,
				event: mockEvent,
				link:  mockLink,
			}
			p := verify.NewVerifyProvider(verify.Deps{
				Cache: mockCache,
				Event: mockEvent,
				Link:  mockLink,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := p.SendVerification(tc.args.ctx, tc.args.email, tc.args.username)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *verifyProviderSuite) TestVerifyVerification() {
	var (
		token = "valid-token"
		email = "test@example.com"
	)

	type testDeps struct {
		cache *commonmocks.MockCache
		event *commonmocks.MockEventPublisher
		link  *commonmocks.MockLinkProvider
	}

	type args struct {
		ctx   context.Context
		token string
	}

	type want struct {
		email string
		err   error
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      want
	}{
		{
			name: "cache_miss",
			args: args{ctx: context.Background(), token: token},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(pkg.ErrNotFound).Once()
			},
			want: want{
				email: "",
				err:   pkg.ErrNotFound,
			},
		},
		{
			name: "success",
			args: args{ctx: context.Background(), token: token},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Run(func(_ context.Context, _ string, dest interface{}) {
						*dest.(*string) = email
					}).
					Return(nil).Once()
			},
			want: want{
				email: email,
				err:   nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := commonmocks.NewMockCache(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())

			deps := testDeps{
				cache: mockCache,
				event: mockEvent,
				link:  mockLink,
			}
			p := verify.NewVerifyProvider(verify.Deps{
				Cache: mockCache,
				Event: mockEvent,
				Link:  mockLink,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := p.VerifyVerification(tc.args.ctx, tc.args.token)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Empty(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.email, got)
			}
		})
	}
}

func (s *verifyProviderSuite) TestInvalidatePasswordReset() {
	mockCache := commonmocks.NewMockCache(s.T())
	p := verify.NewVerifyProvider(verify.Deps{Cache: mockCache})

	token := "token"
	mockCache.EXPECT().
		Delete(mock.Anything, mock.Anything).
		Return(nil).Once()

	err := p.InvalidatePasswordReset(context.Background(), token)
	s.NoError(err)
}

func (s *verifyProviderSuite) TestSendPasswordReset() {
	var (
		email    = "test@example.com"
		username = "testuser"
		link     = "http://reset.link"
	)

	type testDeps struct {
		cache *commonmocks.MockCache
		event *commonmocks.MockEventPublisher
		link  *commonmocks.MockLinkProvider
	}

	type args struct {
		ctx      context.Context
		email    string
		username string
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "cache_error",
			args: args{ctx: context.Background(), email: email, username: username},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					Set(mock.Anything, mock.Anything, email, 15*time.Minute).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			args: args{ctx: context.Background(), email: email, username: username},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					Set(mock.Anything, mock.Anything, email, 15*time.Minute).
					Return(nil).Once()

				deps.link.EXPECT().
					ResetPassword(mock.Anything).
					Return(link).Once()

				deps.event.EXPECT().
					Publish(mock.Anything, mock.MatchedBy(func(e common.Event) bool {
						payload, ok := e.Data.(common.EmailEventPayload)
						return ok && payload.Email == email && payload.Link == link && e.Typ == common.EventEmailResetPassword
					})).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := commonmocks.NewMockCache(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())

			deps := testDeps{
				cache: mockCache,
				event: mockEvent,
				link:  mockLink,
			}
			p := verify.NewVerifyProvider(verify.Deps{
				Cache: mockCache,
				Event: mockEvent,
				Link:  mockLink,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := p.SendPasswordReset(tc.args.ctx, tc.args.email, tc.args.username)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *verifyProviderSuite) TestVerifyPasswordReset() {
	var (
		token = "valid-token"
		email = "test@example.com"
	)

	type testDeps struct {
		cache *commonmocks.MockCache
	}

	type args struct {
		ctx   context.Context
		token string
	}

	type want struct {
		email string
		err   error
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      want
	}{
		{
			name: "cache_miss",
			args: args{ctx: context.Background(), token: token},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(pkg.ErrNotFound).Once()
			},
			want: want{
				email: "",
				err:   pkg.ErrNotFound,
			},
		},
		{
			name: "success",
			args: args{ctx: context.Background(), token: token},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Run(func(_ context.Context, _ string, dest interface{}) {
						*dest.(*string) = email
					}).
					Return(nil).Once()
			},
			want: want{
				email: email,
				err:   nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := commonmocks.NewMockCache(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())

			deps := testDeps{
				cache: mockCache,
			}
			p := verify.NewVerifyProvider(verify.Deps{
				Cache: mockCache,
				Event: mockEvent,
				Link:  mockLink,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := p.VerifyPasswordReset(tc.args.ctx, tc.args.token)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Empty(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.email, got)
			}
		})
	}
}

func (s *verifyProviderSuite) TestValidateResetPasswordToken() {
	var token = "valid-token"

	type testDeps struct {
		cache *commonmocks.MockCache
	}

	type args struct {
		ctx   context.Context
		token string
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      bool
	}{
		{
			name: "token_exists",
			args: args{ctx: context.Background(), token: token},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					IsExist(mock.Anything, mock.Anything).
					Return(true, nil).Once()
			},
			want: true,
		},
		{
			name: "token_not_found",
			args: args{ctx: context.Background(), token: token},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					IsExist(mock.Anything, mock.Anything).
					Return(false, nil).Once()
			},
			want: false,
		},
		{
			name: "cache_error",
			args: args{ctx: context.Background(), token: token},
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					IsExist(mock.Anything, mock.Anything).
					Return(false, assert.AnError).Once()
			},
			want: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := commonmocks.NewMockCache(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockLink := commonmocks.NewMockLinkProvider(s.T())

			deps := testDeps{
				cache: mockCache,
			}
			p := verify.NewVerifyProvider(verify.Deps{
				Cache: mockCache,
				Event: mockEvent,
				Link:  mockLink,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got := p.ValidateResetPasswordToken(tc.args.ctx, tc.args.token)
			s.Equal(tc.want, got)
		})
	}
}
