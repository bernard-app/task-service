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
	UpdateGroup(ctx context.Context, group entity.UpdateGroupRequest, userID uuid.UUID, groupID int64) (*entity.Group, error)
	DeleteGroup(ctx context.Context, groupID int64, userID uuid.UUID) error

	CreateTask(ctx context.Context, task entity.Task, tasksIDs []int64) (*entity.Task, error)
	UpdateTask(ctx context.Context, task entity.UpdateTaskRequest, userID uuid.UUID, taskID int64) (*entity.Task, error)
	DeleteTask(ctx context.Context, userID uuid.UUID, id int64) error
	GetTask(ctx context.Context, id int64, userID uuid.UUID) (*entity.Task, error)
	ListTasks(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.UserTasksTab, error)
	GetTasksByPriority(ctx context.Context, userID uuid.UUID, priority int, limit, offset uint64) ([]*entity.UserTasksTab, error)
	GetTasksByDate(ctx context.Context, userID uuid.UUID, from, to time.Time, limit, offset uint64) ([]*entity.UserTasksTab, error)
	GetTasksByTag(ctx context.Context, userID uuid.UUID, tagID int64, limit, offset uint64) ([]*entity.UserTasksTab, error)

	CreateTag(ctx context.Context, tag entity.Tag) (*entity.Tag, error)
	UpdateTag(ctx context.Context, tagID int64, name *string, color *string, userID uuid.UUID) (*entity.Tag, error)
	DeleteTag(ctx context.Context, tagID int64, userID uuid.UUID) error
	GetTag(ctx context.Context, tagID int64, userID uuid.UUID) (*entity.Tag, error)
	GetTagList(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Tag, error)
	AddTagToTask(ctx context.Context, tagID, taskID int64, userID uuid.UUID) error
	RemoveTagFromTask(ctx context.Context, tagID, taskID int64, userID uuid.UUID) error

	CreateProject(ctx context.Context, project entity.Project) (*entity.Project, error)
	UpdateProject(ctx context.Context, project entity.UpdateProjectRequest, projectID int64, userID uuid.UUID) (*entity.Project, error)
	DeleteProject(ctx context.Context, projectID int64, userID uuid.UUID) error
	GetProjectTree(ctx context.Context, projectID int64, userID uuid.UUID) (*entity.Project, error)
	GetProjects(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Project, error)

	ArchiveTask(ctx context.Context, userID uuid.UUID, taskID int64) (*entity.Task, error)
	ArchiveOldTasks(ctx context.Context) (int64, error)
}

type UseCase struct {
	Config *config.Config
	Log    *slog.Logger
	DB     Storage
}

func New(config *config.Config, db Storage, log *slog.Logger) *UseCase {
	return &UseCase{
		Config: config,
		Log:    log,
		DB:     db,
	}
}
