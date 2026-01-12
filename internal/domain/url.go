package domain

type URLFactory interface {
	Prefix() string
	VerifyEmailURL(token string) string
	ResetPasswordURL(token string) string
	SwaggerURL() string
}
