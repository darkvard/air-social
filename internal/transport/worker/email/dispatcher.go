package email

import (
	"context"
	"fmt"

	"air-social/internal/domain/shared"
	"air-social/pkg"
	"air-social/templates"
)

type Dispatcher struct {
	mailer   shared.Mailer
	handlers map[shared.EventType]shared.EventHandler
}

func NewDispatcher(mailer shared.Mailer) *Dispatcher {
	disp := &Dispatcher{
		mailer:   mailer,
		handlers: make(map[shared.EventType]shared.EventHandler),
	}
	disp.registerHandlers()
	return disp
}

func (d *Dispatcher) Dispatch(ctx context.Context, event shared.Event) error {
	handler, ok := d.handlers[event.Typ]
	if !ok {
		return fmt.Errorf("no handler for event type %s: %w", event.Typ, pkg.ErrNoEventHandler)
	}
	return handler(ctx, event)
}

func (d *Dispatcher) registerHandlers() {
	d.handlers[shared.EventVerify] = d.makeEmailHandler(templates.VerifyEmailPath)
	d.handlers[shared.EventResetPassword] = d.makeEmailHandler(templates.ResetPasswordPath)
}

func (d *Dispatcher) makeEmailHandler(templateFile string) shared.EventHandler {
	return func(ctx context.Context, event shared.Event) error {
		payload, err := shared.UnmarshalEvent[shared.EmailEventPayload](event.Data)
		if err != nil {
			return err
		}

		email := shared.Email{
			To:           payload.Email,
			LayoutFile:   templates.LayoutPath,
			TemplateFile: templateFile,
			Data:         event,
		}

		return d.mailer.Send(ctx, email)
	}
}
