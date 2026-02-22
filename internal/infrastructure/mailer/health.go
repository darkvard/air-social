package mailer

import (
	"context"

	"gopkg.in/gomail.v2"
)

type Health struct {
	dialer *gomail.Dialer
}

func NewHealth(dialer *gomail.Dialer) *Health {
	return &Health{
		dialer: dialer,
	}
}

func (h *Health) Ping(ctx context.Context) error {
	s, err := h.dialer.Dial()
	if err != nil {
		return err
	}
	defer s.Close()
	return nil
}
