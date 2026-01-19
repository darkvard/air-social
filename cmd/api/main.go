package main

import (
	"encoding/json"
	"fmt"

	_ "air-social/docs/swagger"
	"air-social/internal/config"
	"air-social/internal/di"
	"air-social/internal/domain"
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

//	@host		localhost
//	@BasePath	/air-social/api/v1

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	cfg := config.Load()
	urls := transport.NewURLFactory(cfg.Server)

	infra, err := di.NewInfra(cfg)
	if err != nil {
		panic(err)
	}
	defer infra.Close()

	services, err := di.NewServices(cfg, infra, urls)
	if err != nil {
		panic(err)
	}

	printWelcome(cfg, urls)

	NewApp(
		transport.NewServer(cfg, services, infra, urls),
		worker.NewWorker(infra.Rabbit, services.Cache, cfg.Mailer),
		ws.NewHub(),
	).Run()
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
