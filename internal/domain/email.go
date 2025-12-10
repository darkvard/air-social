package domain

type EmailSender interface {
	Send(data *EmailEnvelope) error
}

type EmailEnvelope struct {
	To           string
	TemplateFile string
	Data         any
}
