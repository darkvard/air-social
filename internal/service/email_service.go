package service

import (
	"context"

	"air-social/internal/domain"
	"air-social/pkg"
	"air-social/templates"
)

var _ domain.EventDispatcher = (*EmailServiceImpl)(nil)

type EmailService interface {
	Dispatch(ctx context.Context, evt domain.Event) error
}

type EmailServiceImpl struct {
	mailer   domain.Mailer
	handlers map[domain.EventType]domain.EventHandler
}

func NewEmailService(mailer domain.Mailer) *EmailServiceImpl {
	svc := &EmailServiceImpl{
		mailer:   mailer,
		handlers: make(map[domain.EventType]domain.EventHandler),
	}
	svc.registerHandlers()
	return svc
}

func (e *EmailServiceImpl) registerHandlers() {
	e.handlers[domain.EmailVerify] = e.handleVerifyEmail
	e.handlers[domain.EmailResetPassword] = e.handleResetPassword
}

func (e *EmailServiceImpl) Dispatch(ctx context.Context, evt domain.Event) error {
	handler, ok := e.handlers[evt.EventType]
	if !ok {
		return nil
	}
	return handler(ctx, evt)
}

func (e *EmailServiceImpl) handleVerifyEmail(ctx context.Context, evt domain.Event) error {
	p, err := pkg.UnmarshalEvent[domain.EmailEvent](evt.Data)
	if err != nil {
		return err
	}

	data := domain.EmailVerifyData{
		Name:   p.Name,
		Link:   p.Link,
		Expiry: p.Expiry,
	}

	return e.send(ctx, p.Email, templates.VerifyEmailPath, data)
}

func (e *EmailServiceImpl) handleResetPassword(ctx context.Context, evt domain.Event) error {
	p, err := pkg.UnmarshalEvent[domain.EmailEvent](evt.Data)
	if err != nil {
		return err
	}

	data := domain.EmailVerifyData{
		Name: p.Name,
		Link: p.Link,
	}

	return e.send(ctx, p.Email, templates.ResetPasswordPath, data)
}

func (e *EmailServiceImpl) send(ctx context.Context, to, templateFile string, data any) error {
	email := &domain.Email{
		To:           to,
		LayoutFile:   templates.LayoutPath,
		TemplateFile: templateFile,
		Data:         data,
	}
	return e.mailer.Send(ctx, email)
}
