package domain

import "context"

type Mailer interface {
	Send(ctx context.Context, email *Email) error
}

type Email struct {
	To           string
	LayoutFile   string
	TemplateFile string
	Data         any
}

type EmailRegisterData struct {
	Email string
	Name  string
}

type EmailVerifyData struct {
	Name   string
	Link   string
	Expiry string
}
