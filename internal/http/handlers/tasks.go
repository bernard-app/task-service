package handlers

import (
	"bernard/internal/domain/entity"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.CreateTask"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from request header", err, op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	task, err := TaskCU(r)
	if err != nil {
		h.log.Error("error getting task from request header", err, op)
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

func (h *Handlers) UpdateTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.UpdateTask"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from request header", err, op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.log.Error("error converting task id to int", err, op)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	task, err := TaskCU(r)
	if err != nil {
		h.log.Error("error getting task from request", err, op)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	task.ID = taskID

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

func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.DeleteGroup"

	ctx := r.Context()

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

func (h *Handlers) GetTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetGroup"

	ctx := r.Context()

	_, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from request header", err, op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

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

func (h *Handlers) GetListTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetGroup"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from request header", err, op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	var filters entity.TasksFilter
	filters.UserID = userID

	priority := r.URL.Query().Get("priority")
	if priority != "" {
		filters.Priority, err = strconv.Atoi(priority)
		if err != nil {
			h.log.Error("error parsing priority", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		filters.FilterType = entity.Priority
	}

	tag := r.URL.Query().Get("tag")
	if tag != "" {
		filters.Tag = tag
		filters.FilterType = entity.Tag
	}

	from := r.URL.Query().Get("from")
	if from != "" {
		filters.From, err = time.Parse(time.RFC3339, from)
		if err != nil {
			h.log.Error("error parsing from", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		filters.FilterType = entity.Date
	}

	to := r.URL.Query().Get("to")
	if to != "" {
		filters.To, err = time.Parse(time.RFC3339, to)
		if err != nil {
			h.log.Error("error parsing to", "error", err, "operation", op)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		filters.FilterType = entity.Date
	}

	tasks, err := h.useCase.GetListTask(ctx, filters)
	if err != nil {
		h.log.Error("error getting tasks", "error", err, "operation", op)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	render.JSON(w, r, tasks)
}

func TaskCU(r *http.Request) (entity.Task, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return entity.Task{}, err
	}
	defer r.Body.Close()

	var task entity.Task

	err = json.Unmarshal(body, &task)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}
