package main

import (
	"encoding/json"
	"fmt"

	_ "air-social/docs"
	"air-social/internal/app"
	"air-social/internal/config"
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
	app, err := app.NewApplication()
	if err != nil {
		panic(err)
	}
	defer app.Cleanup()

	pkg.NewLogger(app.Config.Server.Env)
	welcome(app.Config.Server, app.Registry.SwaggerURL())

	app.Run()
}

func welcome(server config.ServerConfig, swaggerURL string) {
	info := struct {
		ContainerPort string `json:"container_port"`
		HostPort      string `json:"host_port"`
		SwaggerURL    string `json:"swagger_url"`
	}{
		ContainerPort: server.Port,
		HostPort:      server.HostPort,
		SwaggerURL:    swaggerURL,
	}

	data, _ := json.MarshalIndent(info, "", "  ")
	pkg.Log().Info("server started")
	fmt.Println(string(data))
}
