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

type Infrastructures struct {
	DB     *sqlx.DB
	Redis  *redis.Client
	Rabbit *amqp.Connection
	Minio  *minio.Client
	Logger *zap.SugaredLogger
}

func initInfrastructures(cfg config.Config) (*Infrastructures, func(), error) {
	var (
		db          *sqlx.DB
		queue       *amqp.Connection
		cache       *redis.Client
		minioClient *minio.Client
		err         error
	)

	cleanup := func() {
		if queue != nil {
			queue.Close()
		}
		if cache != nil {
			cache.Close()
		}
		if db != nil {
			db.Close()
		}
	}

	db, err = postgres.NewConnection(cfg.Postgres)
	if err != nil {
		return nil, func() {}, err
	}

	queue, err = rabbitmq.NewConnection(cfg.RabbitMQ)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}

	cache, err = redisInfra.NewConnection(cfg.Redis)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}

	minioClient, err = minioInfra.NewConnection(cfg.MinIO)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}

	infra := &Infrastructures{
		DB:     db,
		Redis:  cache,
		Rabbit: queue,
		Minio:  minioClient,
		Logger: pkg.Log(),
	}
	return infra, cleanup, nil
}
