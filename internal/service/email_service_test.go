package service

// import (
// 	"context"
// 	"testing"
//
//

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/suite"

// 	"air-social/internal/domain"
// 	"air-social/internal/mocks"
// 	"air-social/templates"
// )

// type emailServiceSuite struct {
// 	suite.Suite
// }

// func TestEmailServiceSuite(t *testing.T) {
// 	suite.Run(t, new(emailServiceSuite))
// }

// func (s *emailServiceSuite) TestDispatch() {
// 	baseData := domain.EmailEvent{
// 		Email:  "test@example.com",
// 		Name:   "Test User",
// 		Link:   "http://link.com",
// 		Expiry: "30m",
// 	}

// 	type args struct {
// 		evt domain.Event
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		setupMock func(mailer *mocks.Mailer, a args)
// 		wantErr   error
// 	}{
// 		{
// 			name: "verify_email_success",
// 			args: args{
// 				evt: domain.Event{
// 					EventType: domain.EmailVerify,
// 					Data:      baseData,
// 				},
// 			},
// 			setupMock: func(mailer *mocks.Mailer, a args) {
// 				mailer.EXPECT().Send(mock.Anything, mock.MatchedBy(func(email *domain.Email) bool {
// 					data, ok := email.Data.(domain.EmailVerifyData)
// 					return ok &&
// 						email.To == baseData.Email &&
// 						email.TemplateFile == templates.VerifyEmailPath &&
// 						data.Link == baseData.Link &&
// 						data.Name == baseData.Name &&
// 						data.Expiry == baseData.Expiry
// 				})).Return(nil).Once()
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name: "verify_email_send_error",
// 			args: args{
// 				evt: domain.Event{
// 					EventType: domain.EmailVerify,
// 					Data:      baseData,
// 				},
// 			},
// 			setupMock: func(mailer *mocks.Mailer, a args) {
// 				mailer.EXPECT().Send(mock.Anything, mock.Anything).Return(assert.AnError).Once()
// 			},
// 			wantErr: assert.AnError,
// 		},
// 		{
// 			name: "reset_password_success",
// 			args: args{
// 				evt: domain.Event{
// 					EventType: domain.EmailResetPassword,
// 					Data:      baseData,
// 				},
// 			},
// 			setupMock: func(mailer *mocks.Mailer, a args) {
// 				mailer.EXPECT().Send(mock.Anything, mock.MatchedBy(func(email *domain.Email) bool {
// 					data, ok := email.Data.(domain.EmailVerifyData)
// 					return ok &&
// 						email.To == baseData.Email &&
// 						email.TemplateFile == templates.ResetPasswordPath &&
// 						data.Link == baseData.Link &&
// 						data.Name == baseData.Name
// 				})).Return(nil).Once()
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name: "unknown_event",
// 			args: args{
// 				evt: domain.Event{
// 					EventType: domain.EventType("unknown_event"),
// 					Data:      baseData,
// 				},
// 			},
// 			setupMock: func(mailer *mocks.Mailer, a args) {
// 				// No calls expected
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name: "parse_error",
// 			args: args{
// 				evt: domain.Event{
// 					EventType: domain.EmailVerify,
// 					Data:      make(chan int), // Invalid for JSON marshal
// 				},
// 			},
// 			setupMock: func(mailer *mocks.Mailer, a args) {
// 				// No calls expected
// 			},
// 			wantErr: assert.AnError,
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			mockMailer := mocks.NewMailer(s.T())
// 			svc := NewEmailService(mockMailer)

// 			if tc.setupMock != nil {
// 				tc.setupMock(mockMailer, tc.args)
// 			}

// 			err := svc.Dispatch(context.Background(), tc.args.evt)

// 			if tc.wantErr != nil {
// 				s.Error(err)
// 				if tc.wantErr != assert.AnError {
// 					s.ErrorIs(err, tc.wantErr)
// 				}
// 			} else {
// 				s.NoError(err)
// 			}
// 		})
// 	}
// }
