package main

import (
	"air-social/internal/app"
	"air-social/internal/config"
	"air-social/pkg"
)

func main() {
	app, err := app.NewApplication()
	if err != nil {
		panic(err)
	}
	defer app.Cleanup()

	pkg.NewLogger(app.Config.Server.Env )
	welcome(app.Config.Server)

	app.Run()
}

func welcome(server config.ServerConfig) {
	pkg.Log().Infow("server started",
		"container_port", server.Port,
		"host_port", server.HostPort,
	)
}

 