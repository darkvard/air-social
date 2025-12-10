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

type mailtrapSender struct {
	dialer *gomail.Dialer
	from   string
}

func NewMailtrapSender(cfg *config.MailConfig) *mailtrapSender {
	return &mailtrapSender{
		dialer: gomail.NewDialer(
			cfg.Host, cfg.Port, cfg.Username, cfg.Password,
		),
		from: fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromAddress),
	}
}

func (m *mailtrapSender) Send(env *domain.EmailEnvelope) error {
	tmpPath := fmt.Sprintf("email/%s", env.TemplateFile)

	t, err := template.ParseFS(templates.EmailFS, tmpPath)
	if err != nil {
		return err
	}

	var subject bytes.Buffer
	if err := t.ExecuteTemplate(&subject, "subject", env.Data); err != nil {
		return err
	}
	var htmlBody bytes.Buffer
	if err := t.ExecuteTemplate(&htmlBody, "htmlBody", env.Data); err != nil {
		return err
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", env.To)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/html", htmlBody.String())

	return m.dialer.DialAndSend(msg)
}
