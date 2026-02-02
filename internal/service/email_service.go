package service

import (
	"context"
	"encoding/json"
	"fmt"

	"air-social/internal/domain"
	"air-social/pkg"
	"air-social/templates"
)

type EmailService interface {
	Dispatch(ctx context.Context, evt domain.Event) error
}

type emailHandler func(evt domain.Event) error

type EmailServiceImpl struct {
	mailer   domain.Mailer
	handlers map[domain.EventType]emailHandler
}

func NewEmailService(mailer domain.Mailer) *EmailServiceImpl {
	svc := &EmailServiceImpl{
		mailer:   mailer,
		handlers: make(map[domain.EventType]emailHandler),
	}
	svc.registerHandlers()
	return svc
}

func (e *EmailServiceImpl) registerHandlers() {
	e.handlers[domain.EmailVerify] = e.verifyEmail
	e.handlers[domain.EmailResetPassword] = e.resetPassword
}

func (e *EmailServiceImpl) Dispatch(ctx context.Context, evt domain.Event) error {
	handler, ok := e.handlers[evt.EventType]
	if !ok {
		return nil
	}
	return handler(evt)
}

func (e *EmailServiceImpl) verifyEmail(evt domain.Event) error {
	return e.handleStandardEmail(evt, templates.VerifyEmailPath)
}

func (e *EmailServiceImpl) resetPassword(evt domain.Event) error {
	return e.handleStandardEmail(evt, templates.ResetPasswordPath)
}

func (e *EmailServiceImpl) handleStandardEmail(evt domain.Event, templateFile string) error {
	var payload domain.EmailEvent
	if err := parsePayloadData(evt, &payload); err != nil {
		return err
	}

	email := &domain.Email{
		To:           payload.Email,
		LayoutFile:   templates.LayoutPath,
		TemplateFile: templateFile,
		Data: domain.EmailVerifyData{
			Name:   payload.Name,
			Link:   payload.Link,
			Expiry: payload.Expiry,
		},
	}

	return e.sendEmail(email, payload.Email, evt.EventType)
}

func parsePayloadData(evt domain.Event, target any) error {
	dataBytes, err := json.Marshal(evt.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, target); err != nil {
		return err
	}
	return nil
}

func (e *EmailServiceImpl) sendEmail(email *domain.Email, recipient string, eventType domain.EventType) error {
	if err := e.mailer.Send(email); err != nil {
		pkg.Log().Errorw("failed to send email", "event_type", eventType, "error", err, "to", recipient)
		return err
	}
	return nil
}
