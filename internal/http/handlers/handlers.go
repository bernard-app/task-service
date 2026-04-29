package handlers

import (
	"bernard/internal/domain/entity"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UseCase interface {
	CreateTask(ctx context.Context, task entity.Task) (*entity.Task, error)
	UpdateTask(ctx context.Context, task entity.Task) (*entity.Task, error)
	DeleteTask(ctx context.Context, userID uuid.UUID, taskID int64) (*entity.Task, error)
	GetTask(ctx context.Context, taskID int64) (*entity.Task, error)
	GetListTask(ctx context.Context, userID uuid.UUID, priority, tag, from, to string) ([]*entity.UserTasksTab, error)
	CreateGroup(ctx context.Context, group entity.Group) (*entity.Group, error)
	UpdateGroup(ctx context.Context, group entity.Group) (*entity.Group, error)
	DeleteGroup(ctx context.Context, groupID int64) (*entity.Group, error)
	CreateProject(ctx context.Context, project entity.Project) (*entity.Project, error)
	UpdateProject(ctx context.Context, project entity.Project) (*entity.Project, error)
	DeleteProject(ctx context.Context, projectID int64) (*entity.Project, error)
	GetProjectTree(ctx context.Context, projectID int64, userID uuid.UUID) (*entity.Project, error)
	GetProjects(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Project, error)
}

type Handlers struct {
	useCase UseCase
	log     *slog.Logger
}

func New(useCase UseCase, log *slog.Logger) *Handlers {
	return &Handlers{useCase: useCase, log: log}
}

// CreateTask godoc
// @Summary Create new task
// @Description creates new user task
// @Tags tasks
// @Param X-User-ID header string true "Authorized user ID"
// @Param input body entity.Task true "Task data"
// @Success 200 {object} entity.Task
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/tasks/ [post]
func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.CreateTask"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from request header", err, op)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID", nil)

		return
	}

	task, err := TaskCU(r)
	if err != nil {
		h.log.Error("error getting task from request header", err, op)
		h.errorResponse(w, http.StatusBadRequest, "error getting task from request", nil)

		return
	}

	task.UserID = userID

	createdTask, err := h.useCase.CreateTask(ctx, task)
	if err != nil {
		h.log.Error("error creating task", "error", err, "operation", op)
		h.serverError(w, r, err)

		return
	}

	h.writeJSON(w, http.StatusOK, createdTask)
}

// UpdateTask godoc
// @Summary Update task
// @Description updates user's task
// @Tags tasks
// @Param id path int64 true "TaskID"
// @Param X-User-ID header string true "Authorized user ID"
// @Param input body entity.Task true "Task data"
// @Success 200 {object} entity.Task
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/tasks/{id} [put]
func (h *Handlers) UpdateTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.UpdateTask"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from request header", err, op)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID", nil)

		return
	}

	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.log.Error("error converting task id to int", err, op)
		h.errorResponse(w, http.StatusBadRequest, "error converting task id to int", nil)

		return
	}

	task, err := TaskCU(r)
	if err != nil {
		h.log.Error("error getting task from request", err, op)
		h.errorResponse(w, http.StatusBadRequest, "error getting task from request", nil)

		return
	}

	task.ID = taskID
	task.UserID = userID

	updatedTask, err := h.useCase.UpdateTask(ctx, task)
	if err != nil {
		h.log.Error("error updating task", "error", err, "operation", op)
		h.serverError(w, r, err)

		return
	}

	h.writeJSON(w, http.StatusOK, updatedTask)
}

// DeleteTask godoc
// @Summary Delete task
// @Description deletes user's task
// @Tags tasks
// @Param id path int64 true "TaskID"
// @Param X-User-ID header string true "Authorized user ID"
// @Success 200 {object} entity.Task
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/tasks/{id} [delete]
func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.DeleteGroup"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id", "error", err, "operation", op)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID", nil)

		return
	}

	strTaskID := chi.URLParam(r, "id")
	taskID, err := strconv.ParseInt(strTaskID, 10, 64)
	if err != nil {
		h.log.Error("error parsing task id", "error", err, "operation", op)
		h.errorResponse(w, http.StatusBadRequest, "error parsing task id", nil)

		return
	}

	deletedTask, err := h.useCase.DeleteTask(ctx, userID, taskID)
	if err != nil {
		h.log.Error("error deleting task", "error", err, "operation", op)
		h.serverError(w, r, err)

		return
	}

	h.writeJSON(w, http.StatusOK, deletedTask)
}

// GetTask godoc
// @Summary Get task
// @Description gets user's task
// @Tags tasks
// @Param id path int64 true "TaskID"
// @Param X-User-ID header string true "Authorized user ID"
// @Success 200 {object} entity.Task
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/tasks/{id} [get]
func (h *Handlers) GetTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetGroup"

	ctx := r.Context()

	_, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from request header", err, op)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID", nil)

		return
	}

	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.log.Error("error parsing task id", "error", err, "operation", op)
		h.errorResponse(w, http.StatusBadRequest, "error parsing task id", nil)

		return
	}

	task, err := h.useCase.GetTask(ctx, taskID)
	if err != nil {
		h.log.Error("error getting task", "error", err, "operation", op)
		h.serverError(w, r, err)

		return
	}

	h.writeJSON(w, http.StatusOK, task)
}

// GetListTask godoc
// @Summary Get list of tasks
// @Description gets user's tasks with optionals filters
// @Tags tasks
// @Param X-User-ID header string true "Authorized user ID"
// @Param priority query int false "Task priority"
// @Param tag query string false "Task tag"
// @Param from query string false "Task from param"
// @Param to query string false "Task to param"
// @Success 200 {object} entity.Task
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/tasks/ [get]
func (h *Handlers) GetListTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetGroup"

	ctx := r.Context()

	userID, err := GetUserID(r)
	priority := r.URL.Query().Get("priority")
	tag := r.URL.Query().Get("tag")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	tasks, err := h.useCase.GetListTask(ctx, userID, priority, tag, from, to)
	if err != nil {
		h.log.Error("error getting tasks", "error", err, "operation", op)
		h.serverError(w, r, err)

		return
	}

	h.writeJSON(w, http.StatusOK, tasks)
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

// CreateGroup godoc
// @Summary Create new group
// @Description gets userID from header
// @Tags groups
// @Param X-User-ID header string true "Authorized user ID"
// @Param input body entity.Group true "Group data"
// @Success 200 {object} entity.Group
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/groups/ [post]
func (h *Handlers) CreateGroup(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.CreateGroup"

	ctx := r.Context()

	_, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from header", err, op)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID", nil)

		return
	}

	group, err := GroupCU(r)
	if err != nil {
		h.log.Error("error calling Group", "operation", op, "error", err)
		h.errorResponse(w, http.StatusBadRequest, "invalid data format", nil)

		return
	}

	createdGroup, err := h.useCase.CreateGroup(ctx, group)
	if err != nil {
		h.log.Error("error calling CreateGroup", "operation", op, "error", err)
		h.serverError(w, r, err)

		return
	}

	h.writeJSON(w, http.StatusOK, createdGroup)
}

// UpdateGroup godoc
// @Summary Update group
// @Description updates group information
// @Tags groups
// @Param id path int64 true "GroupID"
// @Param X-User-ID header string true "Authorized user ID"
// @Param input body entity.Group true "Group data"
// @Success 200 {object} entity.Group
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/groups/{id} [put]
func (h *Handlers) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.UpdateGroup"

	ctx := r.Context()
	_, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from header", err, op)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID", nil)

		return
	}

	groupID, err := strconv.ParseInt(chi.URLParam(r, "groupID"), 10, 64)
	if err != nil {
		h.log.Error("error converting groupID to int", err, op)
		h.errorResponse(w, http.StatusBadRequest, "invalid groupID", nil)

		return
	}

	group, err := GroupCU(r)
	if err != nil {
		h.log.Error("error calling Group", "operation", op, "error", err)
		h.errorResponse(w, http.StatusBadRequest, "invalid data format", nil)

		return
	}

	group.ID = groupID

	updatedGroup, err := h.useCase.UpdateGroup(ctx, group)
	if err != nil {
		h.log.Error("error calling UpdateGroup", "operation", op, "error", err)
		h.serverError(w, r, err)

		return
	}

	h.writeJSON(w, http.StatusOK, updatedGroup)
}

// DeleteGroup godoc
// @Summary Delete group
// @Description deletes group from project
// @Tags groups
// @Param id path int64 true "GroupID"
// @Param X-User-ID header string true "Authorized user ID"
// @Success 200 {object} entity.Group
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/groups/{id} [delete]
func (h *Handlers) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.DeleteGroup"

	ctx := r.Context()

	_, err := GetUserID(r)
	if err != nil {
		h.log.Error("error getting user id from header", err, op)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID", nil)

		return
	}

	groupID, err := strconv.ParseInt(chi.URLParam(r, "groupID"), 10, 64)
	if err != nil {
		h.log.Error("error converting groupID to int", err, op)
		h.errorResponse(w, http.StatusBadRequest, "invalid groupID", nil)

		return
	}

	deletedGroup, err := h.useCase.DeleteGroup(ctx, groupID)
	if err != nil {
		h.log.Error("error calling DeleteGroup", "operation", op, "error", err)
		h.serverError(w, r, err)

		return
	}

	h.writeJSON(w, http.StatusOK, deletedGroup)
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

// CreateProject godoc
// @Summary Create new project
// @Description creates new user project
// @Tags projects
// @Param X-User-ID header string true "Authorized user ID"
// @Param input body entity.Project true "Project data"
// @Success 200 {object} entity.Project
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/projects/ [post]
func (h *Handlers) CreateProject(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.CreateProject"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("empty userID from request", err)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID from request", nil)

		return
	}

	project, err := ProjectCU(r)
	if err != nil {
		h.log.Error("cannot get project", "error", err, "op", op)
		h.errorResponse(w, http.StatusBadRequest, "invalid data format", nil)

		return
	}

	project.UserID = userID

	createdProject, err := h.useCase.CreateProject(ctx, project)
	if err != nil {
		h.log.Error("error creating project", "op", op, "error", err)
		h.serverError(w, r, err)

		return
	}

	h.writeJSON(w, http.StatusOK, createdProject)
}

// UpdateProject godoc
// @Summary Update project
// @Description updates user's project information
// @Tags projects
// @Param id path int64 true "ProjectID"
// @Param X-User-ID header string true "Authorized user ID"
// @Param input body entity.Project true "Project data"
// @Success 200 {object} entity.Project
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/projects/{id} [put]
func (h *Handlers) UpdateProject(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.UpdateProject"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("empty userID from request", err)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID from request", nil)

		return
	}

	project, err := ProjectCU(r)
	if err != nil {
		h.log.Error("cannot get project", "error", err, "op", op)
		h.errorResponse(w, http.StatusBadRequest, "invalid data format", nil)

		return
	}

	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.log.Error("error converting groupID to int", err, op)
		h.errorResponse(w, http.StatusBadRequest, "invalid groupID", nil)

		return
	}

	project.ID = projectID
	project.UserID = userID

	updatedProject, err := h.useCase.UpdateProject(ctx, project)
	if err != nil {
		h.log.Error("error updating project", "op", op, "error", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	h.writeJSON(w, http.StatusOK, updatedProject)
}

// DeleteProject godoc
// @Summary Delete project
// @Description delete user's project
// @Tags projects
// @Param id path int64 true "ProjectID"
// @Param X-User-ID header string true "Authorized user ID"
// @Success 200 {object} entity.Project
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/projects/{id} [delete]
func (h *Handlers) DeleteProject(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.DeleteProject"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("cannot get user ID", "error", err, "op", op)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID from request", nil)

		return
	}

	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.log.Error("error parsing project id", "op", op, "error", err)
		h.errorResponse(w, http.StatusBadRequest, "invalid project ID", nil)

		return
	}

	deletedProject, err := h.useCase.DeleteProject(ctx, projectID)
	if err != nil {
		h.log.Error("error deleting project", "op", op, "error", err)
		h.serverError(w, r, err)

		return
	}

	deletedProject.UserID = userID

	h.writeJSON(w, http.StatusOK, deletedProject)
}

// GetProject godoc
// @Summary Get user projects
// @Description get user's projects with limit and offset
// @Tags projects
// @Param X-User-ID header string true "Authorized user ID"
// @Success 200 {object} entity.Project
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/projects/ [get]
func (h *Handlers) GetProject(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetProject"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("cannot get user ID", "error", err, "op", op)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID from request", nil)

		return
	}

	limit, err := strconv.ParseUint(chi.URLParam(r, "limit"), 10, 64)
	if err != nil {
		h.log.Error("error parsing limit", "error", err, "op", op, "limit", limit)
		h.errorResponse(w, http.StatusBadRequest, "invalid limit", nil)

		return
	}

	offset, err := strconv.ParseUint(chi.URLParam(r, "offset"), 10, 64)
	if err != nil {
		h.log.Error("error parsing offset", "error", err, "op", op, "offset", offset)
		h.errorResponse(w, http.StatusBadRequest, "invalid offset", nil)

		return
	}

	projects, err := h.useCase.GetProjects(ctx, userID, limit, offset)
	if err != nil {
		h.log.Error("error getting projects", "op", op, "error", err)
		h.serverError(w, r, err)

		return
	}

	h.writeJSON(w, http.StatusOK, projects)
}

// GetProjectTree godoc
// @Summary Get user project-tree
// @Description get user's specific project with groups and tasks
// @Tags projects
// @Param id path int64 true "ProjectID"
// @Param X-User-ID header string true "Authorized user ID"
// @Success 200 {object} entity.Project
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/projects/{id} [get]
func (h *Handlers) GetProjectTree(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetProjectTree"

	ctx := r.Context()

	userID, err := GetUserID(r)
	if err != nil {
		h.log.Error("cannot get user ID", "error", err, "op", op)
		h.errorResponse(w, http.StatusUnauthorized, "empty userID from request", nil)

		return
	}

	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.log.Error("error parsing project id", "op", op, "error", err)
		h.errorResponse(w, http.StatusBadRequest, "invalid project ID", nil)

		return
	}

	project, err := h.useCase.GetProjectTree(ctx, projectID, userID)
	if err != nil {
		h.log.Error("error getting project", "op", op, "error", err)
		h.serverError(w, r, err)

		return
	}

	h.writeJSON(w, http.StatusOK, project)
}

func ProjectCU(r *http.Request) (entity.Project, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return entity.Project{}, err
	}
	defer r.Body.Close()

	var project entity.Project

	if err = json.Unmarshal(body, &project); err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (h *Handlers) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.log.Error("failed to write json response", "error", err)
	}
}

func GetUserID(r *http.Request) (uuid.UUID, error) {
	const op = "GetUserID"

	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		return uuid.Nil, fmt.Errorf("op: %s, error parsing user id: %w", op, err)
	}

	return userID, nil
}
