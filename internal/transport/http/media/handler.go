package media

import (
	"air-social/internal/domain/media"
	"air-social/internal/domain/shared"
)

type Handler struct {
	provider shared.LinkProvider
	usecase  media.UseCase
}

func NewHandler(provider shared.LinkProvider, usecase media.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
	}
}
