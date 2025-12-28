package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	boot "air-social/internal/app/bootstrap"
	"air-social/internal/app/provider"
	"air-social/internal/config"
	mess "air-social/internal/infrastructure/messaging"
	rp "air-social/internal/repository/redis"
	"air-social/internal/transport/ws"
	"air-social/internal/worker"
	"air-social/pkg"
)

type Application struct {
	Config        *config.Config
	DB            *sqlx.DB
	Logger        *zap.SugaredLogger
	Redis         *redis.Client
	RabbitConn    *amqp.Connection
	Event         *mess.Publisher
	WorkerManager *worker.Manager
	Http          *provider.HttpProvider
	Hub           *ws.Hub
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
	cache := rp.NewRedisCache(rc)

	// rabbit
	rabbitConn := boot.NewRabbitMQ(cfg.RabbitMQ)
	rabbitPublisher := boot.NewPublisher(rabbitConn)
	workerManager := boot.NewWorkerManager(rabbitConn, cache, cfg.Mailer)

	// http
	httpServer := provider.NewHttpProvider(
		db,
		cfg.Token,
		pkg.NewBcrypt(),
		cache,
		rabbitPublisher,
	)

	return &Application{
		Config:        cfg,
		DB:            db,
		Redis:         rc,
		RabbitConn:    rabbitConn,
		Event:         rabbitPublisher,
		WorkerManager: workerManager,
		Http:          httpServer,
		Hub:           ws.NewHub(),
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