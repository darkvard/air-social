package comment

import (
	"air-social/internal/domain/comment"
	"air-social/internal/domain/shared"
)

type Handler struct {
	provider shared.LinkProvider
	usecase  comment.UseCase
}

func NewHandler(provider shared.LinkProvider, usecase comment.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
	}
}
