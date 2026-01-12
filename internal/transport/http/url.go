package http

import "fmt"

type URLFactoryImpl struct {
	baseURL string
	version string
}

func NewURLFactory(baseURL, version string) *URLFactoryImpl {
	return &URLFactoryImpl{
		baseURL: baseURL,
		version: version,
	}
}

func (r *URLFactoryImpl) Prefix() string {
	return fmt.Sprintf("%s/%s", r.baseURL, r.version)
}

func (r *URLFactoryImpl) VerifyEmailURL(token string) string {
	return fmt.Sprintf("%s%s%s?token=%s", r.Prefix(), AuthGroup, VerifyEmail, token)
}

func (r *URLFactoryImpl) ResetPasswordURL(token string) string {
	return fmt.Sprintf("%s%s%s?token=%s", r.Prefix(), AuthGroup, ResetPassword, token)
}

func (r *URLFactoryImpl) SwaggerURL() string {
	return fmt.Sprintf("%s/swagger/index.html", r.Prefix())
}
