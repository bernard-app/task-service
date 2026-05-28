package application

import (
	"bernard/internal/config"
	"bernard/internal/storage/cache"
	"bernard/internal/storage/postgres"
	"bernard/internal/usecase"
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Container struct {
	cfg     *config.Config
	log     *slog.Logger
	db      *postgres.Storage
	tx      *postgres.PgxTxManager
	cache   *cache.Cache
	usecase *usecase.UseCase
}

func NewContainer(log *slog.Logger, cfg *config.Config) *Container {
	return &Container{
		cfg: cfg,
		log: log,
	}
}

func (di *Container) DB(ctx context.Context) *postgres.Storage {
	if di.db == nil {
		db, err := postgres.NewStorage(ctx, di.cfg.Postgres.Addr)
		if err != nil {
			di.log.Error("failed to init DB", "error", err)
			panic("failed to init DB")
		}

		di.db = db
	}

	return di.db
}

func (di *Container) Tx(ctx context.Context) *postgres.PgxTxManager {
	if di.tx == nil {
		di.tx = postgres.NewTxManager(di.DB(ctx).DB)
	}

	return di.tx
}

func (di *Container) Cache() *cache.Cache {
	if di.cache == nil {
		rdb := redis.NewClient(&redis.Options{
			Addr:     di.cfg.Redis.Addr,
			Password: di.cfg.Redis.Password,
			DB:       di.cfg.Redis.DB,
		})

		di.cache = cache.New(rdb)
	}

	return di.cache
}

func (di *Container) UseCase(ctx context.Context) *usecase.UseCase {
	if di.usecase == nil {
		di.usecase = usecase.New(di.DB(ctx), di.log, di.Tx(ctx), di.Cache())
	}

	return di.usecase
}


