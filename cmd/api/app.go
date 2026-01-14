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
			pkg.Log().Errorw("http server listen failed", "error", err)
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
	pkg.Log().Infow("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpSrv.Shutdown(ctx); err != nil {
		pkg.Log().Errorw("server forced to shutdown", "error", err)
	}

	if err := a.worker.Stop(ctx); err != nil {
		pkg.Log().Errorw("worker forced to shutdown", "error", err)
	}
	pkg.Log().Infow("server exiting")
}
