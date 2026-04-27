package handlers

import (
	"bernard/internal/domain/entity"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type UseCase interface {
	CreateTask(ctx context.Context, task entity.Task) (*entity.Task, error)
	UpdateTask(ctx context.Context, userID uuid.UUID, task entity.Task) (*entity.Task, error)
	DeleteTask(ctx context.Context, userID uuid.UUID, taskID int64) (*entity.Task, error)
	GetTask(ctx context.Context, taskID int64) (*entity.Task, error)
	GetListTask(ctx context.Context, tasksFilter entity.TasksFilter) ([]*entity.UserTasksTab, error)
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

func GetUserID(r *http.Request) (uuid.UUID, error) {
	const op = "GetUserID"

	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		return uuid.Nil, fmt.Errorf("op: %s, error parsing user id: %w", op, err)
	}

	return userID, nil
}
