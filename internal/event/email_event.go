package event

import (
	"context"
	"encoding/json"

	"air-social/internal/domain"
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
		return d.handleEmailRegister(evt)
	default:
		return nil
	}

}

func (d *EmailHandleImpl) handleEmailRegister(evt domain.EventPayload) error {
	var data domain.RegisterEmailData
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return err
	}
	// todo: impl template register email
	env := &domain.EmailEnvelope{
		To:           data.Email,
		TemplateFile: "welcome.html",
		Data: map[string]any{
			"Name":        data.Name,
			"LuckyNumber": 6868,
		},
	}
	return d.sender.Send(env)
}
