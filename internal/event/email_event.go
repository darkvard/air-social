package event

import (
	"context"
	"encoding/json"
	"fmt"

	"air-social/internal/domain"
	"air-social/pkg"
	"air-social/templates"
)

type EmailHandleImpl struct {
	sender domain.EmailSender
}

func NewEmailHandler(sender domain.EmailSender) *EmailHandleImpl {
	return &EmailHandleImpl{sender: sender}
}

func (e *EmailHandleImpl) Handle(ctx context.Context, evt domain.EventPayload) error {
	switch evt.EventType {
	case domain.EmailVerify:
		return e.verifyEmail(evt)
	case domain.EmailResetPassword:
		return e.resetPassword(evt)
	default:
		return nil
	}
}

func (e *EmailHandleImpl) verifyEmail(evt domain.EventPayload) error {
	var payload domain.EventEmailData
	if err := parsePayloadData(evt, &payload); err != nil {
		return err
	}

	env := &domain.EmailEnvelope{
		To:           payload.Email,
		LayoutFile:   templates.LayoutPath,
		TemplateFile: templates.VerifyEmailPath,
		Data: domain.VerifyEmailData{
			Name:   payload.Name,
			Link:   payload.Link,
			Expiry: payload.Expiry,
		},
	}

	return e.sendEmail(env, payload.Email, "verify-email")
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

func (e *EmailHandleImpl) sendEmail(env *domain.EmailEnvelope, email, tag string) error {
	if err := e.sender.Send(env); err != nil {
		pkg.Log().Errorw("failed to send email", "tag", tag, "error", err, "to", email)
		return err
	}
	return nil
}

func (e *EmailHandleImpl) resetPassword(evt domain.EventPayload) error {
	var payload domain.EventEmailData
	if err := parsePayloadData(evt, &payload); err != nil {
		return err
	}

	env := &domain.EmailEnvelope{
		To:           payload.Email,
		LayoutFile:   templates.LayoutPath,
		TemplateFile: templates.ResetPasswordPath,
		Data: domain.VerifyEmailData{
			Name:   payload.Name,
			Link:   payload.Link,
			Expiry: payload.Expiry,
		},
	}

	return e.sendEmail(env, payload.Email, "reset-password")
}
