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
	"github.com/google/uuid"
)

func (h *Handlers) CreateTask(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.CreateTask"

		userID, err := GetUserID(r)
		if err != nil {
			h.log.Error("error getting user id from request header", err, op)
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("error reading request body", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)

			return
		}
		defer r.Body.Close()

		var task entity.Task
		err = json.Unmarshal(body, &task)
		if err != nil {
			h.log.Error("error unmarshalling request body", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		task.UserID = userID

		createdTask, err := h.useCase.CreateTask(ctx, task)
		if err != nil {
			h.log.Error("error creating task", "error", err, "operation", op)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, createdTask)
	}
}

func (h *Handlers) UpdateTask(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.UpdateTask"

		userID, err := GetUserID(r)
		if err != nil {
			h.log.Error("error getting user id from request header", err, op)
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("error reading request body", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)

			return
		}
		defer r.Body.Close()

		var task entity.Task

		err = json.Unmarshal(body, &task)
		if err != nil {
			h.log.Error("error unmarshalling request body", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		task.UserID = userID

		updatedTask, err := h.useCase.UpdateTask(ctx, userID, task)
		if err != nil {
			h.log.Error("error updating task", "error", err, "operation", op)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, updatedTask)
	}
}

func (h *Handlers) DeleteTask(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.DeleteGroup"

		userID, err := GetUserID(r)
		if err != nil {
			h.log.Error("error getting user id", "error", err, "operation", op)
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		strTaskID := chi.URLParam(r, "id")
		taskID, err := strconv.ParseInt(strTaskID, 10, 64)
		if err != nil {
			h.log.Error("error parsing task id", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		deletedTask, err := h.useCase.DeleteTask(ctx, userID, taskID)
		if err != nil {
			h.log.Error("error deleting task", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, deletedTask)
	}
}

func (h *Handlers) GetTask(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.GetGroup"

		taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			h.log.Error("error parsing task id", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		task, err := h.useCase.GetTask(ctx, taskID)
		if err != nil {
			h.log.Error("error getting task", "error", err, "operation", op)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, task)
	}
}

func (h *Handlers) GetListTask(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.GetGroup"

		userID, err := GetUserID(r)
		if err != nil {
			h.log.Error("error getting user id from request header", err, op)
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		// Make transition to route 4 in 1 GET
	}
}

func GetUserID(r *http.Request) (uuid.UUID, error) {
	const op = "GetUserID"

	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		return uuid.UUID{}, err
	}

	return userID, nil
}
