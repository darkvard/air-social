package domain

type URLFactory interface {
	SwaggerURL() string
	MinioConsoleURL() string
	RabbitMQDashboardURL() string

	ApiPath() string
	BaseURL() string

	VerifyEmailURL(token string) string
	ResetPasswordURL(token string) string
	ResetPasswordApiURL() string

	PublicFileURL(key string) string
}
