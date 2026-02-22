package post

import (
	"air-social/internal/domain/post"
	"air-social/internal/domain/shared"
)

type Handler struct {
	provider shared.LinkProvider
	usecase  post.UseCase
}

func NewHandler(provider shared.LinkProvider, usecase post.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
	}
}
