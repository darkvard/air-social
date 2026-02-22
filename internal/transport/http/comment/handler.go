package comment

import (
	"air-social/internal/domain/comment"
	"air-social/internal/domain/common"
)

type Handler struct {
	provider common.LinkProvider
	usecase  comment.UseCase
}

func NewHandler(provider common.LinkProvider, usecase comment.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
	}
}
