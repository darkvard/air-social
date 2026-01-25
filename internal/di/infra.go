package di

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"air-social/internal/config"
	minioInfra "air-social/internal/infrastructure/minio"
	"air-social/internal/infrastructure/postgres"
	"air-social/internal/infrastructure/rabbitmq"
	redisInfra "air-social/internal/infrastructure/redis"
	"air-social/pkg"
)

type InfraContainer struct {
	DB     *sqlx.DB
	Redis  *redis.Client
	Rabbit *amqp.Connection
	Minio  *minio.Client
	Logger *zap.SugaredLogger
}

func NewInfra(cfg config.Config) (*InfraContainer, error) {
	db, err := postgres.NewConnection(cfg.Postgres)
	if err != nil {
		return nil, err
	}

	queue, err := rabbitmq.NewConnection(cfg.RabbitMQ)
	if err != nil {
		return nil, err
	}

	cache, err := redisInfra.NewConnection(cfg.Redis)
	if err != nil {
		return nil, err
	}

	minioClient, err := minioInfra.NewConnection(cfg.MinIO)
	if err != nil {
		return nil, err
	}

	return &InfraContainer{
		DB:     db,
		Redis:  cache,
		Rabbit: queue,
		Minio:  minioClient,
		Logger: pkg.Log(),
	}, nil
}

func (i *InfraContainer) Close() {
	if i.Rabbit != nil {
		i.Rabbit.Close()
	}
	if i.Redis != nil {
		i.Redis.Close()
	}
	if i.DB != nil {
		i.DB.Close()
	}
}
