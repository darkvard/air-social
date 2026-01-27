package di

import "air-social/internal/transport/http/handler"

type Handlers struct {
	Auth   *handler.AuthHandler
	User   *handler.UserHandler
	Media  *handler.MediaHandler
	Health *handler.HealthHandler
}

func initHandlers(services *Services) *Handlers {
	return &Handlers{
		Auth:   handler.NewAuthHandler(services.Auth),
		User:   handler.NewUserHandler(services.User),
		Media:  handler.NewMediaHandler(services.Media),
		Health: handler.NewHealthHandler(services.Health),
	}
}
