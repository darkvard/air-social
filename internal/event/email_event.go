package event

import (
	"context"
	"encoding/json"

	"air-social/internal/domain"
)

type EmailDispatcher interface {
	Dispatch(ctx context.Context, evt domain.EventPayload) error
}

type EmailDispImpl struct {
	engine domain.EmailSender
}

func NewEmailDispatcher(engine domain.EmailSender) *EmailDispImpl {
	return &EmailDispImpl{engine: engine}
}

func (d *EmailDispImpl) Dispatch(ctx context.Context, evt domain.EventPayload) error {
	switch evt.EventType {
	case "email.register":
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
		return d.engine.Send(env)
	default:
		return nil
	}

}
