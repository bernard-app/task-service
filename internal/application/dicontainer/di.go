package dicontainer

import (
	"bernard/internal/config"
	"bernard/internal/storage/postgres"
	"bernard/internal/usecase"
	"context"
	"fmt"
	"log/slog"
)

type Container struct {
	Cfg     *config.Config
	Log     *slog.Logger
	DB      *postgres.Storage
	Tx      *postgres.PgxTxManager
	UseCase *usecase.UseCase
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

	c.Tx = postgres.NewTxManager(c.DB.DB)
	c.UseCase = usecase.New(c.DB, c.Log, c.Tx)

	return nil
}
