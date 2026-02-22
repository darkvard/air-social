package shared

type SystemProvider interface {
	SwaggerURL() string
	MinioConsoleURL() string
	RabbitMQDashboardURL() string
}

type RouteProvider interface {
	BaseURL() string
	ApiPath() string
}

type LinkProvider interface {
	VerifyEmail(token string) string
	ResetPassword(token string) string
	ResetPasswordEndpoint() string
	PublicFile(key string) string
}

type AppLinkManager struct {
	SystemProvider
	RouteProvider
	LinkProvider
}
