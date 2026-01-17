package domain

type URLFactory interface {
	SwaggerUI() string
	MinioConsoleUI() string
	RabbitMQDashboardUI() string

	APIRouterPath() string
	FileStorageBaseURL() string

	VerifyEmailLink(token string) string
	ResetPasswordLink(token string) string
}
