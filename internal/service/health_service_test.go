package service

import (
	"context"
	"errors"
	"testing"

	"github.com/minio/minio-go/v7"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/mocks"
)

type healthServiceSuite struct {
	suite.Suite
}

func TestHealthServiceSuite(t *testing.T) {
	suite.Run(t, new(healthServiceSuite))
}

func (s *healthServiceSuite) TestCheck() {
	type want struct {
		healthy bool
		status  string
	}

	tests := []struct {
		name      string
		setupMock func(db *mocks.HealthDB, r *mocks.HealthRedis, rb *mocks.HealthRabbit, m *mocks.HealthMinio)
		want      want
	}{
		{
			name: "all_healthy",
			setupMock: func(db *mocks.HealthDB, r *mocks.HealthRedis, rb *mocks.HealthRabbit, m *mocks.HealthMinio) {
				db.EXPECT().Ping().Return(nil).Once()
				r.EXPECT().Ping(mock.Anything).Return(goredis.NewStatusResult("PONG", nil)).Once()
				rb.EXPECT().Ping().Return(nil).Once()
				m.EXPECT().ListBuckets(mock.Anything).Return([]minio.BucketInfo{}, nil).Once()
			},
			want: want{
				healthy: true,
				status:  "ok",
			},
		},
		{
			name: "db_error",
			setupMock: func(db *mocks.HealthDB, r *mocks.HealthRedis, rb *mocks.HealthRabbit, m *mocks.HealthMinio) {
				db.EXPECT().Ping().Return(errors.New("db down")).Once()
				r.EXPECT().Ping(mock.Anything).Return(goredis.NewStatusResult("PONG", nil)).Once()
				rb.EXPECT().Ping().Return(nil).Once()
				m.EXPECT().ListBuckets(mock.Anything).Return([]minio.BucketInfo{}, nil).Once()
			},
			want: want{
				healthy: false,
				status:  "error",
			},
		},
		{
			name: "redis_error",
			setupMock: func(db *mocks.HealthDB, r *mocks.HealthRedis, rb *mocks.HealthRabbit, m *mocks.HealthMinio) {
				db.EXPECT().Ping().Return(nil).Once()
				r.EXPECT().Ping(mock.Anything).Return(goredis.NewStatusResult("", errors.New("redis down"))).Once()
				rb.EXPECT().Ping().Return(nil).Once()
				m.EXPECT().ListBuckets(mock.Anything).Return([]minio.BucketInfo{}, nil).Once()
			},
			want: want{
				healthy: false,
				status:  "error",
			},
		},
		{
			name: "rabbit_error",
			setupMock: func(db *mocks.HealthDB, r *mocks.HealthRedis, rb *mocks.HealthRabbit, m *mocks.HealthMinio) {
				db.EXPECT().Ping().Return(nil).Once()
				r.EXPECT().Ping(mock.Anything).Return(goredis.NewStatusResult("PONG", nil)).Once()
				rb.EXPECT().Ping().Return(errors.New("connection closed")).Once()
				m.EXPECT().ListBuckets(mock.Anything).Return([]minio.BucketInfo{}, nil).Once()
			},
			want: want{
				healthy: false,
				status:  "error",
			},
		},
		{
			name: "minio_error",
			setupMock: func(db *mocks.HealthDB, r *mocks.HealthRedis, rb *mocks.HealthRabbit, m *mocks.HealthMinio) {
				db.EXPECT().Ping().Return(nil).Once()
				r.EXPECT().Ping(mock.Anything).Return(goredis.NewStatusResult("PONG", nil)).Once()
				rb.EXPECT().Ping().Return(nil).Once()
				m.EXPECT().ListBuckets(mock.Anything).Return(nil, errors.New("minio down")).Once()
			},
			want: want{
				healthy: false,
				status:  "error",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockDB := mocks.NewHealthDB(s.T())
			mockRedis := mocks.NewHealthRedis(s.T())
			mockRabbit := mocks.NewHealthRabbit(s.T())
			mockMinio := mocks.NewHealthMinio(s.T())
			mockURL := mocks.NewURLFactory(s.T())

			if tc.setupMock != nil {
				tc.setupMock(mockDB, mockRedis, mockRabbit, mockMinio)
			}

			svc := NewHealthService(mockDB, mockRedis, mockRabbit, mockMinio, mockURL)
			healthy, details := svc.Check(context.Background())

			s.Equal(tc.want.healthy, healthy)
			s.Equal(tc.want.status, details["status"])
			s.NotEmpty(details["timestamp"])
		})
	}
}

func (s *healthServiceSuite) TestGetAppInfo() {
	mockURL := mocks.NewURLFactory(s.T())

	svc := NewHealthService(nil, nil, nil, nil, mockURL)

	expectedDocs := "http://localhost:8080/swagger/index.html"
	mockURL.EXPECT().SwaggerUI().Return(expectedDocs).Once()

	info := svc.GetAppInfo()

	s.NotNil(info)
	s.Equal("Air Social API", info["Title"])
	s.Equal(expectedDocs, info["DocsURL"])
}
