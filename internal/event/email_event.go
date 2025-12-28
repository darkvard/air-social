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
	default:
		return nil
	}
}

func (e *EmailHandleImpl) verifyEmail(evt domain.EventPayload) error {
	dataBytes, err := json.Marshal(evt.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	var payload domain.EventEmailVerify
	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	data := domain.VerifyEmailData{
		Name:   payload.Name,
		Link:   payload.Link,
		Expiry: payload.Expiry,
	}

	env := &domain.EmailEnvelope{
		To:           payload.Email,
		LayoutFile:   templates.LayoutPath,
		TemplateFile: templates.VerifyEmailPath,
		Data:         data,
	}

	if err := e.sender.Send(env); err != nil {
		pkg.Log().Errorw("failed to send email", "error", err, "to", payload.Email)
		return err
	}
	return nil
}
