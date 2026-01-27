package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain"
	"air-social/internal/mocks"
	"air-social/templates"
)

type emailServiceSuite struct {
	suite.Suite
}

func TestEmailServiceSuite(t *testing.T) {
	suite.Run(t, new(emailServiceSuite))
}

func (s *emailServiceSuite) TestHandle() {
	baseData := domain.EventEmailData{
		Email:  "test@example.com",
		Name:   "Test User",
		Link:   "http://link.com",
		Expiry: "30m",
	}

	type args struct {
		evt domain.EventPayload
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(sender *mocks.EmailSender, a args)
		wantErr   error
	}{
		{
			name: "verify_email_success",
			args: args{
				evt: domain.EventPayload{
					EventType: domain.EmailVerify,
					Data:      baseData,
				},
			},
			setupMock: func(sender *mocks.EmailSender, a args) {
				sender.EXPECT().Send(mock.MatchedBy(func(env *domain.EmailEnvelope) bool {
					data, ok := env.Data.(domain.VerifyEmailData)
					return ok &&
						env.To == baseData.Email &&
						env.TemplateFile == templates.VerifyEmailPath &&
						data.Link == baseData.Link
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "verify_email_send_error",
			args: args{
				evt: domain.EventPayload{
					EventType: domain.EmailVerify,
					Data:      baseData,
				},
			},
			setupMock: func(sender *mocks.EmailSender, a args) {
				sender.EXPECT().Send(mock.Anything).Return(assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "reset_password_success",
			args: args{
				evt: domain.EventPayload{
					EventType: domain.EmailResetPassword,
					Data:      baseData,
				},
			},
			setupMock: func(sender *mocks.EmailSender, a args) {
				sender.EXPECT().Send(mock.MatchedBy(func(env *domain.EmailEnvelope) bool {
					return env.To == baseData.Email &&
						env.TemplateFile == templates.ResetPasswordPath
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "unknown_event",
			args: args{
				evt: domain.EventPayload{
					EventType: domain.EventType("unknown_event"),
					Data:      baseData,
				},
			},
			setupMock: func(sender *mocks.EmailSender, a args) {
				// No calls expected
			},
			wantErr: nil,
		},
		{
			name: "parse_error",
			args: args{
				evt: domain.EventPayload{
					EventType: domain.EmailVerify,
					Data:      make(chan int), // Invalid for JSON marshal
				},
			},
			setupMock: func(sender *mocks.EmailSender, a args) {
				// No calls expected
			},
			wantErr: assert.AnError,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockSender := mocks.NewEmailSender(s.T())
			svc := NewEmailService(mockSender)

			if tc.setupMock != nil {
				tc.setupMock(mockSender, tc.args)
			}

			err := svc.Handle(context.Background(), tc.args.evt)

			if tc.wantErr != nil {
				s.Error(err)
				if tc.wantErr != assert.AnError {
					s.ErrorIs(err, tc.wantErr)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}
