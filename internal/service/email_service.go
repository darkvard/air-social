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
	Handle(ctx context.Context, evt domain.EventPayload) error
}

type emailHandler func(evt domain.EventPayload) error

type EmailServiceImpl struct {
	sender   domain.EmailSender
	handlers map[domain.EventType]emailHandler
}

func NewEmailService(sender domain.EmailSender) *EmailServiceImpl {
	svc := &EmailServiceImpl{
		sender:   sender,
		handlers: make(map[domain.EventType]emailHandler),
	}
	svc.registerHandlers()
	return svc
}

func (e *EmailServiceImpl) registerHandlers() {
	e.handlers[domain.EmailVerify] = e.verifyEmail
	e.handlers[domain.EmailResetPassword] = e.resetPassword
}

func (e *EmailServiceImpl) Handle(ctx context.Context, evt domain.EventPayload) error {
	handler, ok := e.handlers[evt.EventType]
	if !ok {
		return nil
	}
	return handler(evt)
}

func (e *EmailServiceImpl) verifyEmail(evt domain.EventPayload) error {
	return e.handleStandardEmail(evt, templates.VerifyEmailPath)
}

func (e *EmailServiceImpl) resetPassword(evt domain.EventPayload) error {
	return e.handleStandardEmail(evt, templates.ResetPasswordPath)
}

func (e *EmailServiceImpl) handleStandardEmail(evt domain.EventPayload, templateFile string) error {
	var payload domain.EventEmailData
	if err := parsePayloadData(evt, &payload); err != nil {
		return err
	}

	env := &domain.EmailEnvelope{
		To:           payload.Email,
		LayoutFile:   templates.LayoutPath,
		TemplateFile: templateFile,
		Data: domain.VerifyEmailData{
			Name:   payload.Name,
			Link:   payload.Link,
			Expiry: payload.Expiry,
		},
	}

	return e.sendEmail(env, payload.Email, evt.EventType)
}

func parsePayloadData(evt domain.EventPayload, target any) error {
	dataBytes, err := json.Marshal(evt.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, target); err != nil {
		return err
	}
	return nil
}

func (e *EmailServiceImpl) sendEmail(env *domain.EmailEnvelope, email string, eventType domain.EventType) error {
	if err := e.sender.Send(env); err != nil {
		pkg.Log().Errorw("failed to send email", "event_type", eventType, "error", err, "to", email)
		return err
	}
	return nil
}
