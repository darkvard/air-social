package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	commonmocks "air-social/internal/domain/common/mocks"
	"air-social/internal/domain/health"
	healthmocks "air-social/internal/domain/health/mocks"
)

type healthUseCaseSuite struct {
	suite.Suite
}

func TestHealthUseCaseSuite(t *testing.T) {
	suite.Run(t, new(healthUseCaseSuite))
}

func (s *healthUseCaseSuite) TestCheckStatus() {
	type testDeps struct {
		dbChecker    *healthmocks.MockChecker
		redisChecker *healthmocks.MockChecker
	}

	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      bool
		wantKeys  []string
	}{
		{
			name: "all_healthy",
			args: args{ctx: context.Background()},
			setupMock: func(deps testDeps) {
				deps.dbChecker.EXPECT().
					Ping(mock.Anything).
					Return(nil).
					Once()

				deps.redisChecker.EXPECT().
					Ping(mock.Anything).
					Return(nil).
					Once()
			},
			want:     true,
			wantKeys: []string{"db", "redis", "status", "timestamp"},
		},
		{
			name: "one_service_down",
			args: args{ctx: context.Background()},
			setupMock: func(deps testDeps) {
				deps.dbChecker.EXPECT().
					Ping(mock.Anything).
					Return(errors.New("connection refused")).
					Once()

				deps.redisChecker.EXPECT().
					Ping(mock.Anything).
					Return(nil).
					Once()
			},
			want:     false,
			wantKeys: []string{"db", "redis", "status", "timestamp"},
		},
		{
			name: "all_services_down",
			args: args{ctx: context.Background()},
			setupMock: func(deps testDeps) {
				deps.dbChecker.EXPECT().
					Ping(mock.Anything).
					Return(errors.New("db down")).
					Once()

				deps.redisChecker.EXPECT().
					Ping(mock.Anything).
					Return(errors.New("redis down")).
					Once()
			},
			want:     false,
			wantKeys: []string{"db", "redis", "status", "timestamp"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockDB := healthmocks.NewMockChecker(s.T())
			mockRedis := healthmocks.NewMockChecker(s.T())
			mockSystem := commonmocks.NewMockSystemProvider(s.T())

			deps := testDeps{
				dbChecker:    mockDB,
				redisChecker: mockRedis,
			}

			checkers := map[string]health.Checker{
				"db":    mockDB,
				"redis": mockRedis,
			}

			uc := health.NewUseCase(checkers, mockSystem)

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			gotHealthy, gotDetails := uc.CheckStatus(tc.args.ctx)

			s.Equal(tc.want, gotHealthy)
			for _, key := range tc.wantKeys {
				s.Contains(gotDetails, key)
			}
			s.NotEmpty(gotDetails["timestamp"])
		})
	}
}

func (s *healthUseCaseSuite) TestOverview() {
	var (
		docsURL = "http://localhost/docs"
	)

	type testDeps struct {
		checker *healthmocks.MockChecker
		system  *commonmocks.MockSystemProvider
	}

	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      health.OverviewResponse
	}{
		{
			name: "system_healthy",
			args: args{ctx: context.Background()},
			setupMock: func(deps testDeps) {
				deps.checker.EXPECT().
					Ping(mock.Anything).
					Return(nil).
					Once()

				deps.system.EXPECT().
					SwaggerURL().
					Return(docsURL).
					Once()
			},
			want: health.OverviewResponse{
				Title:    "Air Social API",
				DocsURL:  docsURL,
				Status:   "Active",
				HTTPCode: 200,
			},
		},
		{
			name: "system_unhealthy",
			args: args{ctx: context.Background()},
			setupMock: func(deps testDeps) {
				deps.checker.EXPECT().
					Ping(mock.Anything).
					Return(errors.New("down")).
					Once()

				deps.system.EXPECT().
					SwaggerURL().
					Return(docsURL).
					Once()
			},
			want: health.OverviewResponse{
				Title:    "Air Social API",
				DocsURL:  docsURL,
				Status:   "Maintenance",
				HTTPCode: 503,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockChecker := healthmocks.NewMockChecker(s.T())
			mockSystem := commonmocks.NewMockSystemProvider(s.T())

			deps := testDeps{
				checker: mockChecker,
				system:  mockSystem,
			}

			checkers := map[string]health.Checker{
				"main": mockChecker,
			}

			uc := health.NewUseCase(checkers, mockSystem)

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got := uc.Overview(tc.args.ctx)

			s.Equal(tc.want.Title, got.Title)
			s.Equal(tc.want.DocsURL, got.DocsURL)
			s.Equal(tc.want.Status, got.Status)
			s.Equal(tc.want.HTTPCode, got.HTTPCode)
		})
	}
}
