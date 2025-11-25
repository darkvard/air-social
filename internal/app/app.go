package app

import (
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
	JWT    pkg.JWTAuth
	Http   *provider.HttpProvider
	Hub    *ws.Hub
}

func NewApplication() (*Application, error) {
	cfg := config.Load()

	logger, err := bootstrap.NewLogger(cfg.AppEnv)
	if err != nil {
		return nil, err
	}

	db, err := bootstrap.NewDatabase(cfg.Postgres)
	if err != nil {
		return nil, err
	}

	redis := bootstrap.NewRedis(cfg.Redis)

	jwt := pkg.NewJWTAuth(cfg.JWT)
	hash := pkg.NewBcrypt()
	httpServer := provider.NewHttpProvider(db, jwt, hash)

	return &Application{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Redis:  redis,
		JWT:    jwt,
		Http:   httpServer,
		Hub:    ws.NewHub(),
	}, nil
}

func (a *Application) Cleanup() {
	a.Logger.Sync()
	a.DB.Close()
	a.Redis.Close()
}

func (a *Application) Run() {
	a.NewRouter().Run(a.Config.Server.Port)
}
