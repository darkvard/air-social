package main

import (
	_ "air-social/docs/swagger"
	"air-social/internal/config"
	"air-social/internal/di"
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

	container, cleanup, err := di.Initialize(cfg)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	NewApp(container.Server, container.Worker, container.Hub).Run()
}
