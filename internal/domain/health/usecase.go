package health

import (
	"context"

	"air-social/internal/domain/shared"
)

type UseCase interface {
	CheckStatus(ctx context.Context) (bool, map[string]string)
	AppInfo() map[string]any
}

type Checker interface {
	Ping(ctx context.Context) error
}

type usecase struct {
	checkers map[string]Checker
	link     shared.SystemProvider
}

func NewUseCase(
	checkers map[string]Checker,
	link shared.SystemProvider,
) UseCase {
	return &usecase{
		checkers: checkers,
		link:     link,
	}
}

func (u *usecase) CheckStatus(ctx context.Context) (bool, map[string]string) {
	details := make(map[string]string)
	isHealthy := true

	for name, checker := range u.checkers {
		if err := checker.Ping(ctx); err != nil {
			details[name] = err.Error()
			isHealthy = false
		} else {
			details[name] = "ok"
		}
	}
	return isHealthy, details
}

func (u *usecase) AppInfo() map[string]any {
	return map[string]any{
		"Title":   "Air Social API",
		"DocsURL": u.link.SwaggerURL(),
	}

}
