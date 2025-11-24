package app

import (
	"context"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"air-social/internal/config"
	"air-social/internal/transport/ws"
	"air-social/pkg"
)

type Application struct {
	Config *config.Config
	Logger *zap.SugaredLogger
	DB     *sqlx.DB
	Redis  *redis.Client
	Hub    *ws.Hub
	JWT    pkg.JWTAuth
}

func NewApplication() (*Application, error) {
	cfg := config.Load()

	// logger
	logger, err := newLogger(cfg.AppEnv)
	if err != nil {
		return nil, err
	}
	defer logger.Sync()

	// database
	db, err := newDatabase(cfg.Postgres)
	if err != nil {
		return nil, err
	}
	logger.Info("database connection pool established")

	// redis
	redis := newRedis(cfg.Redis)
	logger.Info("redis cache connection pool established")

	return &Application{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Redis:  redis,
		JWT:    pkg.NewJWTAuth(cfg.JWT),
		Hub:    ws.NewHub(),
	}, nil
}

func newLogger(env string) (*zap.SugaredLogger, error) {
	var baseLogger *zap.Logger
	var err error
	if env == "development" {
		baseLogger, err = zap.NewDevelopment()
	} else {
		baseLogger, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}
	sugar := baseLogger.Sugar()
	return sugar, nil
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

func newRedis(rc config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     rc.Addr,
		Password: rc.Password,
		DB:       rc.DB,
	})
}
