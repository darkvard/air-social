package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"air-social/internal/transport/ws"
	"air-social/internal/worker"
	"air-social/pkg"
)

type App struct {
	httpSrv *http.Server
	worker  *worker.Manager
	ws      *ws.Hub
}

func NewApp(httpSrv *http.Server, worker *worker.Manager, ws *ws.Hub) *App {
	return &App{
		httpSrv: httpSrv,
		worker:  worker,
		ws:      ws,
	}
}

func (a *App) Run() {
	go a.worker.Start(context.Background())
	go a.ws.Run()
	go func() {
		if err := a.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pkg.Log().Error("HTTP server listen: %s\n", err)
		}
	}()
	a.awaitSignal()
	a.shutdown()
}

func (a *App) awaitSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func (a *App) shutdown() {
	pkg.Log().Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.worker.Stop(ctx); err != nil {
		pkg.Log().Error("Worker forced to shutdown: ", err)

	}
	if err := a.httpSrv.Shutdown(ctx); err != nil {
		pkg.Log().Error("Server forced to shutdown: ", err)
	}
	pkg.Log().Info("Server exiting")
}
