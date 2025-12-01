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
	"air-social/internal/transport/ws"
	"air-social/pkg"
)

type Application struct {
	Config *config.Config
	Logger *zap.SugaredLogger
	DB     *sqlx.DB
	Redis  *redis.Client
	Http   *provider.HttpProvider
	Hub    *ws.Hub
}

func NewApplication() (*Application, error) {
	cfg := config.Load()

	db, err := bootstrap.NewDatabase(cfg.Postgres)
	if err != nil {
		return nil, err
	}

	redis := bootstrap.NewRedis(cfg.Redis)

	hash := pkg.NewBcrypt()
	httpServer := provider.NewHttpProvider(db, cfg.Token, hash)

	return &Application{
		Config: cfg,
		DB:     db,
		Redis:  redis,
		Http:   httpServer,
		Hub:    ws.NewHub(),
	}, nil
}

func (a *Application) Cleanup() {
	a.DB.Close()
	a.Redis.Close()
}

func (a *Application) Run() {
	gin.SetMode(gin.DebugMode)
	port := fmt.Sprintf(":%s", a.Config.Server.Port)
	a.NewRouter().Run(port)
}
