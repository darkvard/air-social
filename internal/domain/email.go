package domain

type Mailer interface {
	Send(email *Email) error
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
