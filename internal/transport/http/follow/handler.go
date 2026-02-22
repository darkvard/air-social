package follow

import (
	"air-social/internal/domain/follow"
	"air-social/internal/domain/shared"
)

type Handler struct {
	provider shared.LinkProvider
	usecase  follow.UseCase
}

func NewHandler(provider shared.LinkProvider, usecase follow.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
	}
}
