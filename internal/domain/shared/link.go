package shared

type SystemProvider interface {
	SwaggerURL() string
	MinioConsoleURL() string
	RabbitMQDashboardURL() string
}

type RouteProvider interface {
	BaseURL() string
	ApiPath() string // e.g: /api/v1
}

type AppLinkProvider interface {
	VerifyEmail(token string) string
	ResetPassword(token string) string
	ResetPasswordApi() string
	PublicFile(key string) string
}


// todo: update impl