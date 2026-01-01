package templates

import "embed"

const (
	LayoutPath        = "email/layout_boxed.gohtml"
	VerifyEmailPath   = "email/verify_email.gohtml"
	ResetPasswordPath = "email/reset_password.gohtml"
)

//go:embed email pages
var TemplatesFS embed.FS
