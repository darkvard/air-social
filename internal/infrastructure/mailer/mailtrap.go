package mailer

import (
	"bytes"
	"fmt"
	"html/template"

	"gopkg.in/gomail.v2"

	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/templates"
)

type mailtrap struct {
	dialer *gomail.Dialer
	from   string
}

func NewMailtrap(cfg config.MailConfig) *mailtrap {
	return &mailtrap{
		dialer: gomail.NewDialer(
			cfg.Host, cfg.Port, cfg.Username, cfg.Password,
		),
		from: fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromAddress),
	}
}

func (m *mailtrap) Send(env *domain.EmailEnvelope) error {
	// path
	layoutPath := env.LayoutFile
	contentPath := env.TemplateFile

	// parsing (merge layout + content)
	t, err := template.ParseFS(templates.TemplatesFS, layoutPath, contentPath)
	if err != nil {
		return fmt.Errorf("failed to parse templates (%s + %s): %w", layoutPath, contentPath, err)
	}

	// rendering + binding
	var subjectBuffer bytes.Buffer
	if err := t.ExecuteTemplate(&subjectBuffer, "subject", env.Data); err != nil {
		return fmt.Errorf("failed to execute 'subject' block: %w", err)
	}
	var bodyBuffer bytes.Buffer
	if err := t.ExecuteTemplate(&bodyBuffer, "layout", env.Data); err != nil {
		return fmt.Errorf("failed to execute 'layout' block: %w", err)
	}

	// send email
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", env.To)
	msg.SetHeader("Subject", subjectBuffer.String())
	msg.SetBody("text/html", bodyBuffer.String())
	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("mailtrap send error: %w", err)
	}
	return nil
}
