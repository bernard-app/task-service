package router

import (
	"bernard/internal/http/handlers"
	"bernard/internal/http/middleware/httpLogger"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Router(ctx context.Context, router *chi.Mux, http *handlers.Handlers, log *slog.Logger) {
	router.Use(middleware.RequestID)
	router.Use(httpLogger.New(log))
	router.Use(middleware.Recoverer)
}
