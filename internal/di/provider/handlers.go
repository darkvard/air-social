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

func NewHandler(prov Provider, usecase UseCase) Handler {
	link := prov.Link.LinkProvider
	return Handler{
		Health:  health.NewHandler(usecase.Health),
		Auth:    auth.NewHandler(link, usecase.Auth),
		User:    user.NewHandler(link, usecase.User),
		Media:   media.NewHandler(link, usecase.Media),
		Follow:  follow.NewHandler(link, usecase.Follow),
		Post:    post.NewHandler(link, usecase.Post),
		Comment: comment.NewHandler(link, usecase.Comment),
	}
}
