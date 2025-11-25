package main

import (
	"air-social/internal/app"
)

func main() {
	app, err := app.NewApplication()
	if err != nil {
		panic(err)
	}
	defer app.Cleanup()
	app.Run()
}
