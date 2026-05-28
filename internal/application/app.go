package application

import (
	"bernard/internal/config"
	th "bernard/internal/http/grpc"
	"context"
	"log/slog"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Application struct {
	cfg        *config.Config
	log        *slog.Logger
	container  *Container
	grpcServer *grpc.Server
	wg         *sync.WaitGroup
}

func NewApplication(cfg *config.Config, log *slog.Logger) *Application {
	return &Application{
		cfg:       cfg,
		log:       log,
		container: NewContainer(log, cfg),
		wg:        &sync.WaitGroup{},
	}
}

func (a *Application) MustRun(ctx context.Context) {
	err := a.Run(ctx)
	if err != nil {
		panic(err)
	}
}

func (a *Application) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", a.cfg.HTTPServer.Address)
	if err != nil {
		a.log.Error("failed to listen", "error", err)
	
		return err
	}

	a.grpcServer = grpc.NewServer()

	th.Register(a.grpcServer, a.container.UseCase(ctx), a.log)

	reflection.Register(a.grpcServer)
	
	a.wg.Go(
		func() {
			a.container.UseCase(ctx).StartArchiveWorker(ctx)
		},
	)

	a.wg.Go(
		func() {
			a.log.Info("Run: server started", "address", a.cfg.HTTPServer.Address)

			err := a.grpcServer.Serve(listener)
			if err != nil {
				a.log.Error("failed to serve", "error", err)
			}
		},
	)

	return nil
}

func (a *Application) Shutdown() {
	a.log.Info("Shutdown")

	a.grpcServer.GracefulStop()
	err := a.container.db.Close()
	if err != nil {
		a.log.Error("error closing DB connection")
	}
	
	a.wg.Wait()

	a.log.Info("Shutdown completed gracefully")
}
