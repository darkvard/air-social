package mailer

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"gopkg.in/gomail.v2"

	"air-social/internal/config"
	"air-social/internal/domain/common"
	"air-social/templates"
)

type mailtrap struct {
	dialer *gomail.Dialer
	from   string
}

func NewMailtrap(cfg config.MailConfig, dialer *gomail.Dialer) *mailtrap {
	return &mailtrap{
		dialer: dialer,
		from:   fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromAddress),
	}
}

func (m *mailtrap) Send(ctx context.Context, email common.Email) error {
	// path
	layoutPath := email.LayoutFile
	contentPath := email.TemplateFile

	// parsing (merge layout + content)
	t, err := template.ParseFS(templates.TemplatesFS, layoutPath, contentPath)
	if err != nil {
		return fmt.Errorf("failed to parse templates (%s + %s): %w", layoutPath, contentPath, err)
	}

	// rendering + binding
	var subjectBuffer bytes.Buffer
	if err := t.ExecuteTemplate(&subjectBuffer, "subject", email.Data); err != nil {
		return fmt.Errorf("failed to execute 'subject' block: %w", err)
	}
	var bodyBuffer bytes.Buffer
	if err := t.ExecuteTemplate(&bodyBuffer, "layout", email.Data); err != nil {
		return fmt.Errorf("failed to execute 'layout' block: %w", err)
	}

	// send email
	errChan := make(chan error, 1)

	go func() {
		msg := gomail.NewMessage()
		msg.SetHeader("From", m.from)
		msg.SetHeader("To", email.To)
		msg.SetHeader("Subject", subjectBuffer.String())
		msg.SetBody("text/html", bodyBuffer.String())
		errChan <- m.dialer.DialAndSend(msg)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
