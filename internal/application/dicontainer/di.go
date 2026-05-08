package dicontainer

import (
	"bernard/internal/config"
	"bernard/internal/storage/cache"
	"bernard/internal/storage/postgres"
	"bernard/internal/usecase"
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Container struct {
	Cfg     *config.Config
	Log     *slog.Logger
	DB      *postgres.Storage
	Tx      *postgres.PgxTxManager
	Redis *cache.Cache
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

	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Cfg.Redis.Addr,
		Password: c.Cfg.Redis.Password,
		DB:       c.Cfg.Redis.DB,
	})
	
	c.Redis = cache.New(rdb)
	
	c.UseCase = usecase.New(c.DB, c.Log, c.Tx, c.Redis)

	return nil
}
