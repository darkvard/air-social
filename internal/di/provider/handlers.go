package provider

import (
	"air-social/internal/domain"
	"air-social/internal/transport/http/handler"
)

type Handlers struct {
	Auth   *handler.AuthHandler
	User   *handler.UserHandler
	Media  *handler.MediaHandler
	Health *handler.HealthHandler
	Follow *handler.FollowHandler
	Post   *handler.PostHandler
}

func NewHandlers(services *Services, urlFactory domain.URLFactory) *Handlers {
	return &Handlers{
		Auth:   handler.NewAuthHandler(services.Auth, urlFactory),
		User:   handler.NewUserHandler(services.User, urlFactory),
		Media:  handler.NewMediaHandler(services.Media),
		Health: handler.NewHealthHandler(services.Health),
		Follow: handler.NewFollowHandler(services.Follow, services.User, urlFactory),
		Post:   handler.NewPostHandler(services.Post, urlFactory),
	}
}
