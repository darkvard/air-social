package provider

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"

	"air-social/internal/config"
	mi "air-social/internal/infrastructure/minio"
	"air-social/internal/infrastructure/postgres"
	"air-social/internal/infrastructure/rabbitmq"
	ri "air-social/internal/infrastructure/redis"
	"air-social/pkg"
)

type Infrastructure struct {
	DB       *sqlx.DB
	Redis    *redis.Client
	Rabbit   *amqp.Connection
	Minio    *minio.Client
	Mailtrap *gomail.Dialer
	Logger   *zap.SugaredLogger
}

func NewInfrastructure(cfg config.Config) (*Infrastructure, func(), error) {
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

	cache, err = ri.NewConnection(cfg.Redis)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}

	minioClient, err = mi.NewConnection(cfg.MinIO)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}

	mailer := gomail.NewDialer(
		cfg.Mailer.Host, cfg.Mailer.Port, cfg.Mailer.Username, cfg.Mailer.Password,
	)

	infra := &Infrastructure{
		DB:       db,
		Redis:    cache,
		Rabbit:   queue,
		Minio:    minioClient,
		Mailtrap: mailer,
		Logger:   pkg.Log(),
	}
	return infra, cleanup, nil
}

func (i *Infrastructure) GetRabbit(cfg config.RabbitMQConfig) *rabbitmq.HealthChecker {
	return &rabbitmq.HealthChecker{
		Conn: i.Rabbit,
		URL:  cfg.URL,
	}
}
