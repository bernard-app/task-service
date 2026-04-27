package handlers

import (
	"bernard/internal/domain/entity"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (h *Handlers) CreateGroup(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.CreateGroup"

	ctx := r.Context()

	_, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from header", err, op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	group, err := GroupCU(r)
	if err != nil {
		h.log.Error("error calling Group", "operation", op, "error", err)
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

func (h *Handlers) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.UpdateGroup"

	ctx := r.Context()

	_, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from header", err, op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	groupID, err := strconv.ParseInt(chi.URLParam(r, "groupID"), 10, 64)
	if err != nil {
		h.log.Error("error converting groupID to int", err, op)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	group, err := GroupCU(r)
	if err != nil {
		h.log.Error("error calling Group", "operation", op, "error", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	group.ID = groupID

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

func (h *Handlers) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.DeleteGroup"

	ctx := r.Context()

	_, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from header", err, op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	groupID, err := strconv.ParseInt(chi.URLParam(r, "groupID"), 10, 64)
	if err != nil {
		h.log.Error("error converting groupID to int", err, op)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	deletedGroup, err := h.useCase.DeleteGroup(ctx, groupID)
	if err != nil {
		h.log.Error("error calling DeleteGroup", "operation", op, "error", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	render.JSON(w, r, deletedGroup)
}

func GroupCU(r *http.Request) (entity.Group, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return entity.Group{}, err
	}
	defer r.Body.Close()

	var group entity.Group

	if err = json.Unmarshal(body, &group); err != nil {
		return entity.Group{}, err
	}

	return group, nil
}
