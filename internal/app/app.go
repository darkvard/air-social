package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"air-social/internal/app/bootstrap"
	"air-social/internal/app/provider"
	"air-social/internal/config"
	"air-social/internal/infrastructure/queue"
	"air-social/internal/transport/ws"
	"air-social/pkg"
)

type Application struct {
	Config   *config.Config
	Logger   *zap.SugaredLogger
	DB       *sqlx.DB
	Redis    *redis.Client
	RabbitMQ *queue.RabbitMQPublisher
	Http     *provider.HttpProvider
	Hub      *ws.Hub
}

func NewApplication() (*Application, error) {
	cfg := config.Load()

	// db
	db, err := bootstrap.NewDatabase(cfg.Postgres)
	if err != nil {
		return nil, err
	}

	// redis
	redis := bootstrap.NewRedis(cfg.Redis)

	// rabbit
	rabbitPublisher, err := queue.NewRabbitMQPublisher(
		bootstrap.NewRabbitMQ(cfg.RabbitMQ),
		queue.ExchangeConfig{
			Name: "events",
			Type: "topic",
		}, 10)
	if err != nil {
		return nil, err
	}

	// http
	httpServer := provider.NewHttpProvider(
		db,
		cfg.Token,
		pkg.NewBcrypt(),
		redis,
		rabbitPublisher,
	)

	return &Application{
		Config:   cfg,
		DB:       db,
		Redis:    redis,
		RabbitMQ: rabbitPublisher,
		Http:     httpServer,
		Hub:      ws.NewHub(),
	}, nil
}

func (a *Application) Cleanup() {
	a.DB.Close()
	a.Redis.Close()
	a.RabbitMQ.Close()
	a.RabbitMQ.Conn.Close()
}

func (a *Application) Run() {
	gin.SetMode(gin.DebugMode)
	port := fmt.Sprintf(":%s", a.Config.Server.Port)
	a.NewRouter().Run(port)
}
