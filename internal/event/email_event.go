package event

import (
	"context"
	"encoding/json"

	"air-social/internal/domain"
	"air-social/templates"
)

type EmailHandleImpl struct {
	sender domain.EmailSender
}

func NewEmailHandler(sender domain.EmailSender) *EmailHandleImpl {
	return &EmailHandleImpl{sender: sender}
}

func (d *EmailHandleImpl) Handle(ctx context.Context, evt domain.EventPayload) error {
	switch evt.EventType {
	case domain.EventEmailRegister:
		return d.verifyEmail(evt)
	default:
		return nil
	}
}

func (d *EmailHandleImpl) verifyEmail(evt domain.EventPayload) error {
	var payload domain.RegisterEventPayload
	if err := json.Unmarshal(evt.Data, &payload); err != nil {
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

	return d.sender.Send(env)
}
