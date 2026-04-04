package health

import (
	"context"
	"strconv"
	"sync"
	"time"

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

type Deps struct {
	Checkers map[string]Checker
	Link     common.SystemProvider
}

type usecase struct {
	deps Deps
}

func NewUseCase(
	checkers map[string]Checker,
	link common.SystemProvider,
) UseCase {
	return &usecase{deps: Deps{Checkers: checkers, Link: link}}
}

func (u *usecase) CheckStatus(ctx context.Context) (bool, map[string]string) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var mu sync.Mutex
	var wg sync.WaitGroup

	details := make(map[string]string)
	isHealthy := true

	for name, checker := range u.deps.Checkers {
		wg.Add(1)

		go func(name string, checker Checker) {
			defer wg.Done()

			err := checker.Ping(ctx)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				details[name] = err.Error()
				isHealthy = false
			} else {
				details[name] = "ok"
			}

		}(name, checker)

	}
	wg.Wait()

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
		DocsURL:  u.deps.Link.SwaggerURL(),
		Status:   status,
		HTTPCode: httpCode,
	}
}
