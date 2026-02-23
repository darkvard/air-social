package common

type SystemProvider interface {
	SwaggerURL() string
	MinioConsoleURL() string
	RabbitMQDashboardURL() string
	Print()
}

type RouteProvider interface {
	BaseURL() string
	ApiPath() string
	ApiVersion() string
	Print()
}

type LinkProvider interface {
	VerifyEmail(token string) string
	ResetPassword(token string) string
	ResetPasswordEndpoint() string
	PublicFile(key string) string
	Print()
}

type AppLinkManager struct {
	SystemProvider
	RouteProvider
	LinkProvider
}
