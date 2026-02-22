package health

import (
	"context"
	"strconv"

	"air-social/internal/domain/common"
	"air-social/pkg"
)

type UseCase interface {
	CheckStatus(ctx context.Context) (bool, map[string]string)
	Overview(ctx context.Context) OverviewResponse
}

type Checker interface {
	Ping(ctx context.Context) error
}

type usecase struct {
	checkers map[string]Checker
	link     common.SystemProvider
}

func NewUseCase(
	checkers map[string]Checker,
	link common.SystemProvider,
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

	details["status"] = strconv.FormatBool(isHealthy)
	details["timestamp"] = pkg.TimeNowUTC().String()

	return isHealthy, details
}

func (u *usecase) Overview(ctx context.Context) OverviewResponse {
	isHealthy, _ := u.CheckStatus(ctx)

	status := "Active"
	httpCode := 200

	if !isHealthy {
		status = "Maintenance"
		httpCode = 503
	}

	return OverviewResponse{
		Title:    "Air Social API",
		DocsURL:  u.link.SwaggerURL(),
		Status:   status,
		HTTPCode: httpCode,
	}
}
