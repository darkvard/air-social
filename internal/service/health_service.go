package service

import (
	"context"
	"time"

	"github.com/minio/minio-go/v7"
	goredis "github.com/redis/go-redis/v9"

	"air-social/internal/domain"
	"air-social/pkg"
)

type HealthService interface {
	Check(ctx context.Context) (bool, map[string]string)
	GetAppInfo() map[string]any
}

type HealthDB interface {
	Ping() error
}

type HealthRedis interface {
	Ping(ctx context.Context) *goredis.StatusCmd
}

type HealthRabbit interface {
	Ping() error
}

type HealthMinio interface {
	ListBuckets(ctx context.Context) ([]minio.BucketInfo, error)
}

type HealthServiceImpl struct {
	DB     HealthDB
	Redis  HealthRedis
	Rabbit HealthRabbit
	Minio  HealthMinio
	url    domain.URLFactory
}

func NewHealthService(db HealthDB, redis HealthRedis, rabbit HealthRabbit, minio HealthMinio, url domain.URLFactory) *HealthServiceImpl {
	return &HealthServiceImpl{
		DB:     db,
		Redis:  redis,
		Rabbit: rabbit,
		Minio:  minio,
		url:    url,
	}
}

func (s *HealthServiceImpl) Check(ctx context.Context) (bool, map[string]string) {
	status := "ok"
	details := map[string]string{
		"db":       "ok",
		"redis":    "ok",
		"rabbitmq": "ok",
		"minio":    "ok",
	}
	isHealthy := true

	if err := s.DB.Ping(); err != nil {
		details["db"] = err.Error()
		status = "error"
		isHealthy = false
	}

	if err := s.Redis.Ping(ctx).Err(); err != nil {
		details["redis"] = err.Error()
		status = "error"
		isHealthy = false
	}

	if err := s.Rabbit.Ping(); err != nil {
		details["rabbitmq"] = err.Error()
		status = "error"
		isHealthy = false
	}

	if _, err := s.Minio.ListBuckets(ctx); err != nil {
		details["minio"] = err.Error()
		status = "error"
		isHealthy = false
	}

	details["status"] = status
	details["timestamp"] = pkg.TimeNowUTC().Format(time.RFC3339)

	return isHealthy, details
}

func (s *HealthServiceImpl) GetAppInfo() map[string]any {
	return map[string]any{
		"Title":   "Air Social API",
		"DocsURL": s.url.SwaggerUI(),
	}
}
