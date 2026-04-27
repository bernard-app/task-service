package usecase

import (
	"bernard/internal/config"
	"bernard/internal/domain/entity"
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	CreateGroup(ctx context.Context, group entity.Group) (*entity.Group, error)
	UpdateGroup(ctx context.Context, group entity.Group) (*entity.Group, error)
	DeleteGroup(ctx context.Context, groupID int64) (*entity.Group, error)
	CreateTask(ctx context.Context, task entity.Task) (*entity.Task, error)
	UpdateTask(ctx context.Context, userID uuid.UUID, task entity.Task) (*entity.Task, error)
	DeleteTask(ctx context.Context, userID uuid.UUID, id int64) (*entity.Task, error)
	GetTask(ctx context.Context, id int64) (*entity.Task, error)
	ListTasks(ctx context.Context, userID uuid.UUID) ([]*entity.UserTasksTab, error)
	GetTasksByPriority(ctx context.Context, userID uuid.UUID, priority int) ([]*entity.UserTasksTab, error)
	GetTasksByDate(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*entity.UserTasksTab, error)
	GetTasksByTag(ctx context.Context, userID uuid.UUID, tag string) ([]*entity.UserTasksTab, error)
	CreateProject(ctx context.Context, project entity.Project) (*entity.Project, error)
	UpdateProject(ctx context.Context, project entity.Project) (*entity.Project, error)
	DeleteProject(ctx context.Context, projectID int64) (*entity.Project, error)
	GetProjectTree(ctx context.Context, projectID int64, userID uuid.UUID) (*entity.Project, error)
	GetProjects(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Project, error)
}

type UseCase struct {
	config *config.Config
	log    *slog.Logger
	db     Storage
}

func New(config *config.Config, db Storage, log *slog.Logger) *UseCase {
	return &UseCase{
		config: config,
		log:    log,
		db:     db,
	}
}
