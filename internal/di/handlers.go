package di

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
}

func initHandlers(services *Services, url domain.URLFactory) *Handlers {
	return &Handlers{
		Auth:   handler.NewAuthHandler(services.Auth, url),
		User:   handler.NewUserHandler(services.User),
		Media:  handler.NewMediaHandler(services.Media),
		Health: handler.NewHealthHandler(services.Health),
		Follow: handler.NewFollowHandler(services.Follow),
	}
}
