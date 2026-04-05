package provider

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	"air-social/internal/config"
	minioinfra "air-social/internal/infrastructure/minio"
	mongoinfra "air-social/internal/infrastructure/mongo"
	"air-social/internal/infrastructure/postgres"
	"air-social/internal/infrastructure/rabbitmq"
	redisinfra "air-social/internal/infrastructure/redis"
	"air-social/pkg"
)

type Infrastructure struct {
	Postgres *sqlx.DB
	Redis    *redis.Client
	Mongo    *mongo.Client
	MongoDB  *mongo.Database
	Rabbit   *amqp.Connection
	Minio    *minio.Client
	Logger   *zap.SugaredLogger
}

func NewInfrastructure(cfg config.Config) (*Infrastructure, func(), error) {
	var (
		psql        *sqlx.DB
		mongoClient *mongo.Client
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
		if psql != nil {
			psql.Close()
		}
		if mongoClient != nil {
			// Use a background context for cleanup, as the original request context may have been cancelled.
			_ = mongoClient.Disconnect(context.Background())
		}
	}

	psql, err = postgres.NewConnection(cfg.Postgres)
	if err != nil {
		return nil, func() {}, err
	}

	queue, err = rabbitmq.NewConnection(cfg.RabbitMQ)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}

	cache, err = redisinfra.NewConnection(cfg.Redis)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}

	minioClient, err = minioinfra.NewConnection(cfg.MinIO)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}

	mongoClient, err = mongoinfra.NewConnection(cfg.Mongo)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}
	mongoDB := mongoClient.Database(cfg.Mongo.Database)
	{
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Mongo.ConnectTimeout)
		defer cancel()
		if err = mongoinfra.SetupIndexes(ctx, mongoDB); err != nil {
			cleanup()
			return nil, func() {}, err
		}
	}

	infra := &Infrastructure{
		Postgres: psql,
		Mongo:    mongoClient,
		MongoDB:  mongoDB,
		Redis:    cache,
		Rabbit:   queue,
		Minio:    minioClient,
		Logger:   pkg.Log(),
	}
	return infra, cleanup, nil
}
