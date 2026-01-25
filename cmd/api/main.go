package main

import (
	"encoding/json"
	"fmt"

	_ "air-social/docs/swagger"
	"air-social/internal/config"
	"air-social/internal/di"
	"air-social/internal/domain"
	"air-social/internal/infrastructure/rabbitmq"
	transport "air-social/internal/transport/http"
	"air-social/internal/transport/worker"
	"air-social/internal/transport/worker/email"
	"air-social/internal/transport/ws"
	"air-social/pkg"
)

//	@title			Air Social API
//	@version		1.0
//	@description	Backend service for Air Social application.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost
//	@BasePath	/air-social/api/v1

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	cfg := config.Load()
	url := transport.NewURLFactory(cfg.Server)

	infra, err := di.NewInfra(cfg)
	if err != nil {
		panic(err)
	}
	defer infra.Close()

	services, err := di.NewServices(cfg, infra, url)
	if err != nil {
		panic(err)
	}

	app, err := newApp(cfg, url, infra, services)
	if err != nil {
		panic(err)
	}

	printWelcome(cfg, url)

	app.Run()
}

func newApp(cfg config.Config, url domain.URLFactory, infra *di.InfraContainer, services *di.ServiceContainer) (*App, error) {
	server := transport.NewServer(cfg, services, infra, url)
	if server == nil {
		return nil, fmt.Errorf("new server failed")
	}

	workers := newWorker(infra, services)
	if workers == nil {
		return nil, fmt.Errorf("new workers failed")
	}

	return NewApp(server, workers, ws.NewHub()), nil
}

func newWorker(infra *di.InfraContainer, services *di.ServiceContainer) *worker.Manager {
	conn := infra.Rabbit
	cache := services.Cache
	dispatcher := services.Email

	queueExchangeCfg := rabbitmq.EventsExchange
	verifyWorker := email.NewEmailWorker(conn, cache, dispatcher, queueExchangeCfg, rabbitmq.EmailVerifyQueueConfig)
	resetPasswordWorker := email.NewEmailWorker(conn, cache, dispatcher, queueExchangeCfg, rabbitmq.EmailResetPasswordQueueConfig)

	return worker.NewManager(verifyWorker, resetPasswordWorker)
}

func printWelcome(cfg config.Config, urls domain.URLFactory) {
	info := map[string]string{
		"swagger_docs":      urls.SwaggerUI(),
		"rabbit_mq_console": urls.RabbitMQDashboardUI(),
		"minio_console":     urls.MinioConsoleUI(),
	}
	data, _ := json.MarshalIndent(info, "", "  ")
	pkg.Log().Info("server started")
	fmt.Println(string(data))
}
