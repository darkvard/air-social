package route

// import (
// 	"encoding/json"
// 	"fmt"
// 	"strings"
//
//

// 	"air-social/internal/config"
// 	"air-social/pkg"
// )

// type URLFactoryImpl struct {
// 	protocol     string
// 	domain       string
// 	appName      string
// 	version      string
// 	bucketPublic string
// }

// func NewURLFactory(cfg config.Config) *URLFactoryImpl {
// 	serverCfg := cfg.Server
// 	return &URLFactoryImpl{
// 		protocol:     serverCfg.Protocol,
// 		domain:       serverCfg.Domain,
// 		appName:      serverCfg.AppName,
// 		version:      serverCfg.Version,
// 		bucketPublic: cfg.MinIO.BucketPublic,
// 	}
// }

// func (r *URLFactoryImpl) BaseURL() string {
// 	return fmt.Sprintf("%s://%s", r.protocol, r.domain)
// }

// func (r *URLFactoryImpl) ApiURL() string {
// 	return fmt.Sprintf("%s/%s/api/%s", r.BaseURL(), r.appName, r.version)
// }

// func (r *URLFactoryImpl) ApiPath() string {
// 	return fmt.Sprintf("api/%s", r.version)
// }

// func (r *URLFactoryImpl) VerifyEmailURL(token string) string {
// 	return fmt.Sprintf("%s%s%s?token=%s", r.ApiURL(), AuthGroup, VerifyEmail, token)
// }

// func (r *URLFactoryImpl) ResetPasswordURL(token string) string {
// 	return fmt.Sprintf("%s%s%s?token=%s", r.ApiURL(), AuthGroup, ResetPassword, token)
// }

// func (r *URLFactoryImpl) SwaggerURL() string {
// 	return fmt.Sprintf("%s/swagger/index.html", r.ApiURL())
// }

// func (r *URLFactoryImpl) MinioConsoleURL() string {
// 	return fmt.Sprintf("%s/storage-admin/", r.BaseURL())
// }

// func (r *URLFactoryImpl) RabbitMQDashboardURL() string {
// 	return fmt.Sprintf("%s/rabbitmq/", r.BaseURL())
// }

// func (r *URLFactoryImpl) PublicFileURL(key string) string {
// 	if key == "" {
// 		return ""
// 	}
// 	domain := strings.TrimSuffix(r.BaseURL(), "/")
// 	return fmt.Sprintf("%s/%s/%s", domain, r.bucketPublic, key)
// }

// func (r *URLFactoryImpl) PrintInfraConsole() {
// 	info := map[string]string{
// 		"swagger_docs":      r.SwaggerURL(),
// 		"rabbit_mq_console": r.RabbitMQDashboardURL(),
// 		"minio_console":     r.MinioConsoleURL(),
// 	}
// 	data, _ := json.MarshalIndent(info, "", "  ")
// 	pkg.Log().Info("server started")
// 	fmt.Println(string(data))
// }

// func (r *URLFactoryImpl) ResetPasswordApiURL() string {
// 	return fmt.Sprintf("%s%s%s", r.ApiURL(), AuthGroup, ResetPassword)
// }
