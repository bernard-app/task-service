package handlers

import (
	"bernard/internal/domain/entity"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (h *Handlers) CreateGroup(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.CreateGroup"

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("error reading body", "operation", op, "error", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}
		defer r.Body.Close()

		var group entity.Group

		if err = json.Unmarshal(body, &group); err != nil {
			h.log.Error("error unmarshalling body", "operation", op, "error", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		createdGroup, err := h.useCase.CreateGroup(ctx, group)
		if err != nil {
			h.log.Error("error calling CreateGroup", "operation", op, "error", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, createdGroup)
	}
}

func (h *Handlers) UpdateGroup(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.UpdateGroup"

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("error reading body", "operation", op, "error", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}
		defer r.Body.Close()

		var group entity.Group

		if err = json.Unmarshal(body, &group); err != nil {
			h.log.Error("error unmarshalling body", "operation", op, "error", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		updatedGroup, err := h.useCase.UpdateGroup(ctx, group)
		if err != nil {
			h.log.Error("error calling UpdateGroup", "operation", op, "error", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, updatedGroup)
	}
}

func (h *Handlers) DeleteGroup(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.DeleteGroup"

		groupID := chi.URLParam(r, "id")

		intGroupID, err := strconv.ParseInt(groupID, 10, 64)
		if err != nil {
			h.log.Error("error parsing group id", "operation", op, "error", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		deletedGroup, err := h.useCase.DeleteGroup(ctx, intGroupID)
		if err != nil {
			h.log.Error("error calling DeleteGroup", "operation", op, "error", err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, deletedGroup)
	}
}