package app

import (
	"context"
	"time"
)

type HealthResult struct {
	Status    string `json:"status"`
	DB        string `json:"db"`
	Redis     string `json:"redis"`
	Timestamp string `json:"timestamp"`
}

func (app *Application) HealthStatus() HealthResult {
	ok := "OK"

	dbStatus := ok
	if err := app.DB.Ping(); err != nil {
		dbStatus = err.Error()
	}

	redisStatus := ok
	if err := app.Redis.Ping(context.Background()).Err(); err != nil {
		redisStatus = err.Error()
	}

	status := ok
	if dbStatus != ok || redisStatus != ok {
		status = "ERROR"
	}

	return HealthResult{
		Status:    status,
		DB:        dbStatus,
		Redis:     redisStatus,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
