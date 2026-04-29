package dicontainer

import (
	"bernard/internal/config"
	"bernard/internal/http/handlers"
	"bernard/internal/http/router"
	"bernard/internal/storage/postgres"
	"bernard/internal/usecase"
	"context"
	"fmt"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

type Container struct {
	Cfg          *config.Config
	Log          *slog.Logger
	DB           *postgres.Storage
	HTTPRouter   *chi.Mux
	UseCase      *usecase.UseCase
	HTTPHandlers *handlers.Handlers
}

func NewContainer(log *slog.Logger, cfg *config.Config) *Container {
	return &Container{
		Cfg: cfg,
		Log: log,
	}
}

func (c *Container) Init(ctx context.Context) error {
	var err error

	c.DB, err = postgres.NewStorage(ctx, c.Cfg.Postgres.Addr)
	if err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}

	c.Log.Info("database initialized")

	c.HTTPRouter = chi.NewRouter()

	c.UseCase = usecase.New(c.Cfg, c.DB, c.Log)

	c.HTTPHandlers = handlers.New(c.UseCase, c.Log)

	router.Router(c.HTTPRouter, c.HTTPHandlers, c.Log)

	return nil
}
