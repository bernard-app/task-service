package usecase

import (
	"bernard/internal/domain/entity"
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type TxManager interface {
	ReadWrite(ctx context.Context, fn func(ctxTx context.Context) error) error
}

type Storage interface {
	CreateGroup(ctx context.Context, group entity.Group) (*entity.Group, error)
	UpdateGroup(ctx context.Context, group entity.UpdateGroupRequest, userID uuid.UUID, groupID int64) (*entity.Group, error)
	DeleteGroup(ctx context.Context, groupID int64, userID uuid.UUID) error
	GetGroup(ctx context.Context, userID uuid.UUID, groupID int64) (*entity.Group, error)
	GetProjectGroups(ctx context.Context, userID uuid.UUID, projectID int64, limit, offset uint64) ([]*entity.Group, error)

	CreateTask(ctx context.Context, task entity.Task) (*entity.Task, error)
	UpdateTask(ctx context.Context, task entity.UpdateTaskRequest, userID uuid.UUID, taskID int64) (*entity.Task, error)
	DeleteTask(ctx context.Context, userID uuid.UUID, id int64) error
	GetTask(ctx context.Context, id int64, userID uuid.UUID) (*entity.Task, error)
	ListTasks(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Task, error)
	GetTasksByPriority(ctx context.Context, userID uuid.UUID, priority int, limit, offset uint64) ([]*entity.Task, error)
	GetTasksByDate(ctx context.Context, userID uuid.UUID, from, to time.Time, limit, offset uint64) ([]*entity.Task, error)
	GetTasksByTag(ctx context.Context, userID uuid.UUID, tagID int64, limit, offset uint64) ([]*entity.Task, error)
	GetGroupTasks(ctx context.Context, groupID int64, userID uuid.UUID, limit, offset uint64) ([]*entity.Task, error)

	CreateTag(ctx context.Context, tag entity.Tag) (*entity.Tag, error)
	UpdateTag(ctx context.Context, tagID int64, name *string, color *string, userID uuid.UUID) (*entity.Tag, error)
	DeleteTag(ctx context.Context, tagID int64, userID uuid.UUID) error
	GetTag(ctx context.Context, tagID int64, userID uuid.UUID) (*entity.Tag, error)
	GetTagList(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Tag, error)
	GetTaskTags(ctx context.Context, taskID int64) ([]*entity.Tag, error)
	AddTagsToTask(ctx context.Context, tagsIDs []int64, taskID int64) error
	RemoveTagsFromTask(ctx context.Context, tagsIDs []int64, taskID int64) error
	RemoveAllTagsFromTask(ctx context.Context, taskID int64) error 

	CreateProject(ctx context.Context, project entity.Project) (*entity.Project, error)
	UpdateProject(ctx context.Context, project entity.UpdateProjectRequest, projectID int64, userID uuid.UUID) (*entity.Project, error)
	DeleteProject(ctx context.Context, projectID int64, userID uuid.UUID) error
	GetProject(ctx context.Context, projectID int64, userID uuid.UUID) (*entity.Project, error)
	GetProjects(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Project, error)

	ArchiveTask(ctx context.Context, userID uuid.UUID, taskID int64) (*entity.Task, error)
	ArchiveOldTasks(ctx context.Context) (int64, error)
}

type UseCase struct {
	Log    *slog.Logger
	DB     Storage
	Tx     TxManager
}

func New(db Storage, log *slog.Logger, tx TxManager) *UseCase {
	return &UseCase{
		Log:    log,
		DB:     db,
		Tx:     tx,
	}
}
