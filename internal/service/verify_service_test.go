package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain"
	"air-social/internal/mocks"
	"air-social/pkg"
)

type verifyServiceSuite struct {
	suite.Suite
}

func TestVerifyServiceSuite(t *testing.T) {
	suite.Run(t, new(verifyServiceSuite))
}

func (s *verifyServiceSuite) TestSendEmailVerification() {
	email := "test@example.com"
	username := "tester"
	verifyLink := "http://verify.link/token"

	tests := []struct {
		name      string
		setupMock func(c *mocks.CacheStorage, e *mocks.EventPublisher, u *mocks.URLFactory)
		wantErr   error
	}{
		{
			name: "cache_error",
			setupMock: func(c *mocks.CacheStorage, e *mocks.EventPublisher, u *mocks.URLFactory) {
				c.EXPECT().Set(mock.Anything, mock.Anything, email, 30*time.Minute).Return(errors.New("cache error")).Once()
			},
			wantErr: errors.New("cache error"),
		},
		{
			name: "publish_error",
			setupMock: func(c *mocks.CacheStorage, e *mocks.EventPublisher, u *mocks.URLFactory) {
				c.EXPECT().Set(mock.Anything, mock.Anything, email, 30*time.Minute).Return(nil).Once()
				u.EXPECT().VerifyEmailLink(mock.Anything).Return(verifyLink).Once()
				e.EXPECT().Publish(mock.Anything, string(domain.EmailVerify), mock.Anything).Return(errors.New("pub error")).Once()
			},
			wantErr: errors.New("pub error"),
		},
		{
			name: "success",
			setupMock: func(c *mocks.CacheStorage, e *mocks.EventPublisher, u *mocks.URLFactory) {
				c.EXPECT().Set(mock.Anything, mock.MatchedBy(func(key string) bool {
					return len(key) > 0 // check key format if needed
				}), email, 30*time.Minute).Return(nil).Once()

				u.EXPECT().VerifyEmailLink(mock.Anything).Return(verifyLink).Once()

				e.EXPECT().Publish(mock.Anything, string(domain.EmailVerify), mock.MatchedBy(func(evt domain.Event) bool {
					data, ok := evt.Data.(domain.EmailEvent)
					return ok && data.Email == email && data.Link == verifyLink && evt.EventType == domain.EmailVerify
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := mocks.NewCacheStorage(s.T())
			mockEvent := mocks.NewEventPublisher(s.T())
			mockURL := mocks.NewURLFactory(s.T())

			svc := NewVerifyService(mockCache, mockEvent, mockURL)

			if tc.setupMock != nil {
				tc.setupMock(mockCache, mockEvent, mockURL)
			}

			err := svc.SendEmailVerification(context.Background(), email, username)
			if tc.wantErr != nil {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *verifyServiceSuite) TestVerifyEmailToken() {
	token := "valid-token"
	email := "test@example.com"

	tests := []struct {
		name      string
		token     string
		setupMock func(c *mocks.CacheStorage)
		want      string
		wantErr   error
	}{
		{
			name:  "cache_miss",
			token: "invalid",
			setupMock: func(c *mocks.CacheStorage) {
				c.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(pkg.ErrNotFound).Once()
			},
			want:    "",
			wantErr: pkg.ErrNotFound,
		},
		{
			name:  "success",
			token: token,
			setupMock: func(c *mocks.CacheStorage) {
				c.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
					Run(func(ctx context.Context, key string, dest any) {
						*dest.(*string) = email
					}).Return(nil).Once()
			},
			want:    email,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := mocks.NewCacheStorage(s.T())
			svc := NewVerifyService(mockCache, nil, nil)

			if tc.setupMock != nil {
				tc.setupMock(mockCache)
			}

			got, err := svc.VerifyEmailToken(context.Background(), tc.token)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
				s.Equal(tc.want, got)
			}
		})
	}
}

func (s *verifyServiceSuite) TestSendPasswordReset() {
	email := "test@example.com"
	username := "tester"
	resetLink := "http://reset.link/token"

	tests := []struct {
		name      string
		setupMock func(c *mocks.CacheStorage, e *mocks.EventPublisher, u *mocks.URLFactory)
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(c *mocks.CacheStorage, e *mocks.EventPublisher, u *mocks.URLFactory) {
				c.EXPECT().Set(mock.Anything, mock.Anything, email, 15*time.Minute).Return(nil).Once()
				u.EXPECT().ResetPasswordLink(mock.Anything).Return(resetLink).Once()
				e.EXPECT().Publish(mock.Anything, string(domain.EmailResetPassword), mock.MatchedBy(func(evt domain.Event) bool {
					data, ok := evt.Data.(domain.EmailEvent)
					return ok && data.Email == email && data.Link == resetLink && evt.EventType == domain.EmailResetPassword
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := mocks.NewCacheStorage(s.T())
			mockEvent := mocks.NewEventPublisher(s.T())
			mockURL := mocks.NewURLFactory(s.T())

			svc := NewVerifyService(mockCache, mockEvent, mockURL)

			if tc.setupMock != nil {
				tc.setupMock(mockCache, mockEvent, mockURL)
			}

			err := svc.SendPasswordReset(context.Background(), email, username)
			if tc.wantErr != nil {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *verifyServiceSuite) TestVerifyPasswordResetToken() {
	token := "valid-token"
	email := "test@example.com"

	tests := []struct {
		name      string
		setupMock func(c *mocks.CacheStorage)
		want      string
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(c *mocks.CacheStorage) {
				c.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
					Run(func(ctx context.Context, key string, dest any) {
						*dest.(*string) = email
					}).Return(nil).Once()
			},
			want:    email,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := mocks.NewCacheStorage(s.T())
			svc := NewVerifyService(mockCache, nil, nil)
			if tc.setupMock != nil {
				tc.setupMock(mockCache)
			}
			got, err := svc.VerifyPasswordResetToken(context.Background(), token)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
				s.Equal(tc.want, got)
			}
		})
	}
}
