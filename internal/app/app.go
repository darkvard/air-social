package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"air-social/internal/app/bootstrap"
	"air-social/internal/app/provider"
	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/infrastructure/messaging"
	"air-social/internal/transport/ws"
	"air-social/internal/worker"
	"air-social/pkg"
)

type Application struct {
	Config *config.Config
	DB     *sqlx.DB
	Logger *zap.SugaredLogger
	Redis  *redis.Client

	RabbitConn *amqp.Connection
	Event      domain.EventPublisher
	Worker     worker.Worker

	Http *provider.HttpProvider
	Hub  *ws.Hub
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
	rabbitConn := bootstrap.NewRabbitMQ(cfg.RabbitMQ)
	publisher, err := messaging.NewPublisher(
		rabbitConn,
		messaging.ExchangeConfig{
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
		publisher,
	)

	return &Application{
		Config:     cfg,
		DB:         db,
		Redis:      redis,
		RabbitConn: rabbitConn,
		Event:      publisher,
		Worker:     nil,
		Http:       httpServer,
		Hub:        ws.NewHub(),
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
	port := fmt.Sprintf(":%s", a.Config.Server.Port)
	a.NewRouter().Run(port)
}
