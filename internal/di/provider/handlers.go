package provider

import (
	"air-social/internal/transport/http/auth"
	"air-social/internal/transport/http/comment"
	"air-social/internal/transport/http/follow"
	"air-social/internal/transport/http/health"
	"air-social/internal/transport/http/media"
	"air-social/internal/transport/http/post"
	"air-social/internal/transport/http/user"
)

type Handler struct {
	Health  health.Handler
	Auth    auth.Handler
	User    user.Handler
	Media   media.Handler
	Follow  follow.Handler
	Post    post.Handler
	Comment comment.Handler
}

func NewHandlers(prov Provider, usecase UseCase) *Handler {
	return &Handler{
		Health:  health.NewHandler(),
		Auth:    auth.NewHandler(prov.Link, usecase.Auth),
		User:    user.NewHandler(prov.Link, usecase.User),
		Media:   media.NewHandler(prov.Link, usecase.Media),
		Follow:  follow.NewHandler(prov.Link, usecase.Follow),
		Post:    post.NewHandler(prov.Link, usecase.Post),
		Comment: comment.NewHandler(prov.Link, usecase.Comment),
	}
}
