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

func (h *Handlers) CreateProject(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.CreateProject"

	ctx := r.Context()

	project, err := ProjectCU(r)
	if err != nil {
		h.log.Error("cannot get project", "error", err, "op", op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	createdProject, err := h.useCase.CreateProject(ctx, project)
	if err != nil {
		h.log.Error("error creating project", "op", op, "error", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	render.JSON(w, r, createdProject)
}

func (h *Handlers) UpdateProject(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.UpdateProject"

	ctx := r.Context()

	project, err := ProjectCU(r)
	if err != nil {
		h.log.Error("cannot get project", "error", err, "op", op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.log.Error("error converting groupID to int", err, op)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	project.ID = projectID

	updatedProject, err := h.useCase.UpdateProject(ctx, project)
	if err != nil {
		h.log.Error("error updating project", "op", op, "error", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	render.JSON(w, r, updatedProject)
}

func (h *Handlers) DeleteProject(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.DeleteProject"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("cannot get user ID", "error", err, "op", op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.log.Error("error parsing project id", "op", op, "error", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	deletedProject, err := h.useCase.DeleteProject(ctx, projectID)
	if err != nil {
		h.log.Error("error deleting project", "op", op, "error", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	deletedProject.UserID = userID

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	render.JSON(w, r, deletedProject)
}

func (h *Handlers) GetProject(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetProject"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("cannot get user ID", "error", err, "op", op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	limit, err := strconv.ParseUint(chi.URLParam(r, "limit"), 10, 64)
	if err != nil {
		h.log.Error("error parsing limit", "error", err, "op", op, "limit", limit)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	offset, err := strconv.ParseUint(chi.URLParam(r, "offset"), 10, 64)
	if err != nil {
		h.log.Error("error parsing offset", "error", err, "op", op, "offset", offset)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	projects, err := h.useCase.GetProjects(ctx, userID, limit, offset)
	if err != nil {
		h.log.Error("error getting projects", "op", op, "error", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	render.JSON(w, r, projects)
}

func (h *Handlers) GetProjectTree(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetProjectTree"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("cannot get user ID", "error", err, "op", op)
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.log.Error("error parsing project id", "op", op, "error", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	project, err := h.useCase.GetProjectTree(ctx, projectID, userID)
	if err != nil {
		h.log.Error("error getting project", "op", op, "error", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	render.JSON(w, r, project)
}

func ProjectCU(r *http.Request) (entity.Project, error) {
	userID, err := GetUserID(r)
	if err != nil {
		return entity.Project{}, err
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return entity.Project{}, err
	}
	defer r.Body.Close()

	var project entity.Project

	if err = json.Unmarshal(body, &project); err != nil {
		return entity.Project{}, err
	}

	project.UserID = userID

	return project, nil
}
