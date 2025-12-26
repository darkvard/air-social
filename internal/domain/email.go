package domain

type EmailSender interface {
	Send(data *EmailEnvelope) error
}

type EmailEnvelope struct {
	To           string
	LayoutFile   string
	TemplateFile string
	Data         any
}

type RegisterEmailData struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type VerifyEmailData struct {
	Name   string `json:"name"`
	Link   string `json:"link"`
	Expiry string `json:"expiry"`
}
