package usecase

import (
	"bernard/internal/config"
	"log/slog"
)

type Storage interface {
}

type UseCase struct {
	config *config.Config
	log    *slog.Logger
	db     Storage
}

func New(config *config.Config, db Storage, log *slog.Logger) *UseCase {
	return &UseCase{
		config: config,
		log:    log,
		db:     db,
	}
}
