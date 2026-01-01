package routes

import "fmt"

const Health = "/health"

const (
	AuthGroup      = "/auth"
	Register       = "/register"
	Login          = "/login"
	Refresh        = "/refresh"
	ResetPassword  = "/reset-password"
	ForgotPassword = "/forgot-password"
	VerifyEmail    = "/verify-email"
	Logout         = "/logout"
)


type Registry interface {
	Prefix() string
	VerifyEmailURL(token string) string
	ResetPasswordURL(token string) string
}

type RegistryImpl struct {
	baseURL string
	version string
}

func NewRegistry(baseURL, version string) *RegistryImpl {
	return &RegistryImpl{
		baseURL: baseURL,
		version: version,
	}
}

func (r *RegistryImpl) Prefix() string {
	return fmt.Sprintf("%s/%s", r.baseURL, r.version)
}

func (r *RegistryImpl) VerifyEmailURL(token string) string {
	return fmt.Sprintf("%s%s%s?token=%s", r.Prefix(), AuthGroup, VerifyEmail, token)
}

func (r *RegistryImpl) ResetPasswordURL(token string) string {
	return fmt.Sprintf("%s%s%s?token=%s", r.Prefix(), AuthGroup, ResetPassword, token)
}