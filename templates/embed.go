package templates

import "embed"

const (
	LayoutPath      = "email/layout_boxed.gohtml"
	VerifyEmailPath = "email/verify_email.gohtml"
)

//go:embed email/*.gohtml
var EmailFS embed.FS
