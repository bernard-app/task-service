package main

import (
	"bernard/internal/application"
	"bernard/internal/config"
	"bernard/internal/logger"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoadConfig()
	ctx, cancel := context.WithCancel(context.Background())

	log := logger.SetupLogger(cfg.Env)

	app := application.NewApplication(ctx, cfg, log)

	app.MustRun()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	cancel()
	app.Shutdown()
}
