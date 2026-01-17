package http

import (
	"fmt"

	"air-social/internal/config"
)

type URLFactoryImpl struct {
	protocol string
	domain   string
	appName  string
	version  string
}

func NewURLFactory(cfg config.ServerConfig) *URLFactoryImpl {
	return &URLFactoryImpl{
		protocol: cfg.Protocol,
		domain:   cfg.Domain,
		appName:  cfg.AppName,
		version:  cfg.Version,
	}
}

func (r *URLFactoryImpl) baseURL() string {
	return fmt.Sprintf("%s://%s", r.protocol, r.domain)
}

func (r *URLFactoryImpl) apiBaseURL() string {
	return fmt.Sprintf("%s/%s/api/%s", r.baseURL(), r.appName, r.version)
}

func (r *URLFactoryImpl) APIRouterPath() string {
	return fmt.Sprintf("api/%s", r.version)
}

func (r *URLFactoryImpl) VerifyEmailLink(token string) string {
	return fmt.Sprintf("%s%s%s?token=%s", r.apiBaseURL(), AuthGroup, VerifyEmail, token)
}

func (r *URLFactoryImpl) ResetPasswordLink(token string) string {
	return fmt.Sprintf("%s%s%s?token=%s", r.apiBaseURL(), AuthGroup, ResetPassword, token)
}

func (r *URLFactoryImpl) SwaggerUI() string {
	return fmt.Sprintf("%s/swagger/index.html", r.apiBaseURL())
}

func (r *URLFactoryImpl) MinioConsoleUI() string {
	return fmt.Sprintf("%s/storage-admin/", r.baseURL())
}

func (r *URLFactoryImpl) RabbitMQDashboardUI() string {
	return fmt.Sprintf("%s/rabbitmq/", r.baseURL())
}

// 4. Resources
func (r *URLFactoryImpl) FileStorageBaseURL() string {
	return fmt.Sprintf("%s/%s-public", r.baseURL(), r.appName)
}
