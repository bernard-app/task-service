package router

import (
	"bernard/internal/http/handlers"
	"bernard/internal/http/middleware/httpLogger"
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Router(router *chi.Mux, http *handlers.Handlers, log *slog.Logger) {
	router.Use(middleware.RequestID)
	router.Use(httpLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(5 * time.Second))

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/tasks", func(r chi.Router) {
			r.Post("/", http.CreateTask)
			r.Get("/", http.GetListTask)

			r.Route("/{id}", func(r chi.Router) {
				r.Put("/", http.UpdateTask)
				r.Delete("/", http.DeleteTask)
				r.Get("/", http.GetTask)
			})
		})

		r.Route("/projects", func(r chi.Router) {
			r.Post("/", http.CreateProject)
			r.Get("/", http.GetProject)

			r.Route("/{id}", func(r chi.Router) {
				r.Put("/", http.UpdateProject)
				r.Delete("/", http.DeleteProject)
				r.Get("/", http.GetProjectTree)
			})
		})

		r.Route("/groups", func(r chi.Router) {
			r.Post("/", http.CreateGroup)

			r.Route("/{id}", func(r chi.Router) {
				r.Put("/", http.UpdateGroup)
				r.Delete("/", http.DeleteGroup)
			})
		})
	})
}
