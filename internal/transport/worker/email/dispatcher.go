package email

import (
	"context"
	"fmt"

	"air-social/internal/domain/common"
	"air-social/pkg"
	"air-social/templates"
)

type Dispatcher struct {
	mailer   common.Mailer
	handlers map[common.EventType]common.EventHandler
}

func NewDispatcher(mailer common.Mailer) *Dispatcher {
	disp := &Dispatcher{
		mailer:   mailer,
		handlers: make(map[common.EventType]common.EventHandler),
	}
	disp.registerHandlers()
	return disp
}

func (d *Dispatcher) Dispatch(ctx context.Context, event common.Event) error {
	handler, ok := d.handlers[event.Typ]
	if !ok {
		return fmt.Errorf("no handler for event type %s: %w", event.Typ, pkg.ErrNoEventHandler)
	}
	return handler(ctx, event)
}

func (d *Dispatcher) registerHandlers() {
	d.handlers[common.EventVerify] = d.makeEmailHandler(templates.VerifyEmailPath)
	d.handlers[common.EventResetPassword] = d.makeEmailHandler(templates.ResetPasswordPath)
}

func (d *Dispatcher) makeEmailHandler(templateFile string) common.EventHandler {
	return func(ctx context.Context, event common.Event) error {
		payload, err := common.UnmarshalEvent[common.EmailEventPayload](event.Data)
		if err != nil {
			return err
		}

		email := common.Email{
			To:           payload.Email,
			LayoutFile:   templates.LayoutPath,
			TemplateFile: templateFile,
			Data:         payload,
		}

		return d.mailer.Send(ctx, email)
	}
}
