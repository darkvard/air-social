package app

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	boot "air-social/internal/app/bootstrap"
	"air-social/internal/app/provider"
	"air-social/internal/config"
	mess "air-social/internal/infrastructure/messaging"
	"air-social/internal/repository/redis"
	"air-social/internal/routes"
	"air-social/internal/transport/ws"
	"air-social/internal/worker"
	"air-social/pkg"
)

type Application struct {
	Config     *config.Config
	DB         *sqlx.DB
	Logger     *zap.SugaredLogger
	Redis      *goredis.Client
	RabbitConn *amqp.Connection
	Event      *mess.Publisher
	Worker     *worker.Manager
	Http       *provider.HttpProvider
	Hub        *ws.Hub
	Registry   routes.Registry
}

func NewApplication() (*Application, error) {
	cfg := config.Load()

	// db
	db, err := boot.NewDatabase(cfg.Postgres)
	if err != nil {
		return nil, err
	}

	// redis
	rc := boot.NewRedisClient(cfg.Redis)
	cache := redis.NewRedisCache(rc)

	// rabbit
	rabbitConn := boot.NewRabbitMQ(cfg.RabbitMQ)
	rabbitPublisher := boot.NewPublisher(rabbitConn)
	workerManager := boot.NewWorkerManager(rabbitConn, cache, cfg.Mailer)

	rr := routes.NewRegistry(cfg.Server.BaseURL, cfg.Server.Version)

	// http
	deps := &provider.Container{
		DB:       db,
		Redis:    rc,
		Rabbit:   rabbitConn,
		Config:   cfg,
		Cache:    cache,
		Pub:      rabbitPublisher,
		Registry: rr,
	}
	httpServer := provider.NewHttpProvider(deps)

	return &Application{
		Config:     cfg,
		DB:         db,
		Redis:      rc,
		RabbitConn: rabbitConn,
		Event:      rabbitPublisher,
		Worker:     workerManager,
		Http:       httpServer,
		Hub:        ws.NewHub(),
		Registry:   rr,
	}, nil
}

func (a *Application) Cleanup() {
	if a.Event != nil {
		a.Event.Close()
	}
	if a.RabbitConn != nil {
		a.RabbitConn.Close()
	}
	if a.Redis != nil {
		a.Redis.Close()
	}
	if a.DB != nil {
		a.DB.Close()
	}
}

func (a *Application) Run() {
	gin.SetMode(gin.DebugMode)

	if err := a.Worker.Start(context.Background()); err != nil {
		pkg.Log().Errorw("failed to start worker", "error", err)
	}

	port := fmt.Sprintf(":%s", a.Config.Server.Port)
	a.NewRouter().Run(port)
}
