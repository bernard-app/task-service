package handlers

import "log/slog"

type UseCase interface {
}

type Handlers struct {
	useCase UseCase
	log     *slog.Logger
}

func New(useCase UseCase, log *slog.Logger) *Handlers {
	return &Handlers{useCase: useCase, log: log}
}
