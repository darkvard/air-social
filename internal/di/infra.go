package di

import (
	"context"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"air-social/internal/config"
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
	db, err := newDatabase(cfg.Postgres)
	if err != nil {
		return nil, err
	}

	minioClient, err := newMinioClient(cfg.MinIO)
	if err != nil {
		return nil, err
	}

	return &InfraContainer{
		DB:     db,
		Redis:  newRedisClient(cfg.Redis),
		Rabbit: newRabbitMQ(cfg.RabbitMQ),
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

func newDatabase(ps config.PostgresConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", ps.DSN)
	if err != nil {
		return nil, fmt.Errorf("cannot open DB: %w", err)
	}

	db.SetMaxOpenConns(ps.MaxOpenConn)
	db.SetMaxIdleConns(ps.MaxIdleConn)
	db.SetConnMaxLifetime(ps.MaxLifeTime)
	db.SetConnMaxIdleTime(ps.MaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("cannot ping DB: %w", err)
	}

	return db, nil
}

func newRedisClient(rc config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     rc.Addr,
		Password: rc.Password,
		DB:       rc.DB,
	})
}

func newRabbitMQ(cfg config.RabbitMQConfig) *amqp.Connection {
	conn, err := amqp.Dial(cfg.ConnURL)
	if err != nil {
		panic(err)
	}
	return conn
}

func newMinioClient(cfg config.MinioStorageConfig) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UserSSl,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	buckets := []string{cfg.BucketPublic, cfg.BucketPrivate}

	for _, bucketName := range buckets {
		err := ensureBucketExists(ctx, minioClient, bucketName, cfg.BucketPublic)
		if err != nil {
			return nil, err
		}
	}

	return minioClient, nil
}

func ensureBucketExists(ctx context.Context, client *minio.Client, bucketName, publicName string) error {
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("check bucket %s error: %w", bucketName, err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket %s error: %w", bucketName, err)
		}

		if bucketName == publicName {
			policy := fmt.Sprintf(`{
                "Version": "2012-10-17",
                "Statement": [
                    {"Effect": "Allow", "Principal": {"AWS": ["*"]}, "Action": ["s3:GetObject"], "Resource": ["arn:aws:s3:::%s/*"]}
                ]
            }`, bucketName)

			if err := client.SetBucketPolicy(ctx, bucketName, policy); err != nil {
				return fmt.Errorf("set public policy error: %w", err)
			}
			fmt.Printf("Set public policy for: %s\n", bucketName)
		}
	}

	return nil
}
