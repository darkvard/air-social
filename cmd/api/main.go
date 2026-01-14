package main

import (
	"encoding/json"
	"fmt"

	_ "air-social/docs"
	"air-social/internal/config"
	"air-social/internal/di"
	transport "air-social/internal/transport/http"
	"air-social/internal/transport/ws"
	"air-social/internal/worker"
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

//	@host		localhost:3000
//	@BasePath	/v1

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	cfg := config.Load()
	urls := transport.NewURLFactory(cfg.Server.BaseURL, cfg.Server.Version)

	infra, err := di.NewInfra(*cfg)
	if err != nil {
		panic(err)
	}
	defer infra.Close()

	services, err := di.NewServices(*cfg, infra, urls)
	if err != nil {
		panic(err)
	}

	printWelcome(cfg, urls.SwaggerURL())

	NewApp(
		transport.NewServer(*cfg, services, infra),
		worker.NewWorker(infra.Rabbit, services.Cache, cfg.Mailer),
		ws.NewHub(),
	).Run()
}

func printWelcome(cfg *config.Config, swaggerURL string) {
	info := map[string]string{
		"container_port":    cfg.Server.Port,
		"host_port":         cfg.Server.HostPort,
		"rabbit_mq_manager": cfg.RabbitMQ.ManagementURL,
		"minio_manager":     cfg.MinIO.PublicURL,
		"swagger_url":       swaggerURL,
	}
	data, _ := json.MarshalIndent(info, "", "  ")
	pkg.Log().Info("server started")
	fmt.Println(string(data))
}
