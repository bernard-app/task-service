package application

import (
	"bernard/internal/application/dicontainer"
	"bernard/internal/config"
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Application struct {
	ctx       context.Context
	cfg       *config.Config
	log       *slog.Logger
	container *dicontainer.Container
	server    *http.Server
	wg        *sync.WaitGroup
}

func NewApplication(ctx context.Context, cfg *config.Config, log *slog.Logger) *Application {
	return &Application{
		ctx:       ctx,
		cfg:       cfg,
		log:       log,
		container: dicontainer.NewContainer(log, cfg),
		wg:        &sync.WaitGroup{},
	}
}

func (a *Application) MustRun() {
	err := a.Run()
	if err != nil {
		panic(err)
	}
}

func (a *Application) Run() error {
	err := a.container.Init(a.ctx)
	if err != nil {
		a.log.Error("failed to init dependencies dicontainer", "error", err)
		return err
	}

	a.server = &http.Server{
		Addr:         a.cfg.HTTPServer.Address,
		Handler:      a.container.HTTPRouter,
		ReadTimeout:  a.cfg.HTTPServer.Timeout,
		WriteTimeout: a.cfg.HTTPServer.Timeout,
		IdleTimeout:  a.cfg.HTTPServer.IdleTimeout,
	}

	a.wg.Add(1)

	go func() {
		defer a.wg.Done()
		a.log.Info("Run: server started")

		err = a.server.ListenAndServe()
		if err != nil {
			a.log.Error("ListenAndServe", "error", err)
		}
	}()

	return nil
}

func (a *Application) Shutdown() {
	a.log.Info("Shutdown")

	srvCtx, srvCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer srvCancel()

	err := a.server.Shutdown(srvCtx)
	if err != nil {
		a.log.Error("Shutdown: failed to shutdown server", "error", err)
	}

	a.wg.Wait()

	a.log.Info("Shutdown completed gracefully")
}
