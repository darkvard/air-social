package provider

import "air-social/internal/transport/http/handler"

type HandlerProvider struct {
	Auth *handler.AuthHandler
}

func NewHandlerProvider(service *ServiceProvider) *HandlerProvider {
	return &HandlerProvider{
		Auth: handler.NewAuthHandler(service.Auth),
	}
}
