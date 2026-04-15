package usecase

import (
	"bernard/internal/config"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UseCase struct {
	config config.Config
	log    *slog.Logger
	db     *pgxpool.Pool
}
