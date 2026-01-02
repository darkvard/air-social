package provider

import (
	"air-social/internal/transport/http/handler"
)

type HandlerProvider struct {
	Health *handler.HealthHandler
	Auth   *handler.AuthHandler
	User   *handler.UserHandler
}

func NewHandlerProvider(
	deps *Container,
	service *ServiceProvider,
) *HandlerProvider {
	return &HandlerProvider{
		Health: handler.NewHealthHandler(deps.DB, deps.Redis, deps.Rabbit),
		Auth:   handler.NewAuthHandler(service.Auth),
		User:   handler.NewUserHandler(service.User),
	}
}
