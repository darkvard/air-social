package shared

import "context"

type Mailer interface {
	Send(ctx context.Context, email Email) error
}

type Email struct {
	To           string
	LayoutFile   string
	TemplateFile string
	Data         any
}
 