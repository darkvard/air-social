package http

import (
	"encoding/json"
	"fmt"

	"air-social/internal/config"
	"air-social/pkg"
)

type URLFactoryImpl struct {
	protocol     string
	domain       string
	appName      string
	version      string
	bucketPublic string
}

func NewURLFactory(cfg config.Config) *URLFactoryImpl {
	serverCfg := cfg.Server
	return &URLFactoryImpl{
		protocol:     serverCfg.Protocol,
		domain:       serverCfg.Domain,
		appName:      serverCfg.AppName,
		version:      serverCfg.Version,
		bucketPublic: cfg.MinIO.BucketPublic,
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

func (r *URLFactoryImpl) FileStorageBaseURL() string {
	return r.baseURL()
}

func (r *URLFactoryImpl) PrintInfraConsole() {
	info := map[string]string{
		"swagger_docs":      r.SwaggerUI(),
		"rabbit_mq_console": r.RabbitMQDashboardUI(),
		"minio_console":     r.MinioConsoleUI(),
	}
	data, _ := json.MarshalIndent(info, "", "  ")
	pkg.Log().Info("server started")
	fmt.Println(string(data))
}

func (r *URLFactoryImpl) ResetPasswordEndpoint() string {
	return fmt.Sprintf("%s%s%s", r.apiBaseURL(), AuthGroup, ResetPassword)
}
