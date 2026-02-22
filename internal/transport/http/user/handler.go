package user

import (
	"air-social/internal/domain/shared"
	"air-social/internal/domain/user"
)

type Handler struct {
	provider shared.LinkProvider
	usecase  user.UseCase
}

func NewHandler(provider shared.LinkProvider, usecase user.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
	}
}
