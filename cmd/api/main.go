package main

import (
	"net/http"

	"air-social/internal/app"
)

func main() {
	app, err := app.NewApplication()
	if err != nil {
		panic(err)
	}

	// CLOSE RESOURCES ON APP SHUTDOWN
	defer app.Logger.Sync()
	defer app.DB.Close()
	defer app.Redis.Close()

	http.ListenAndServe(":8080", nil)
}
