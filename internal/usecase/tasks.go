package usecase

import (
	"bernard/internal/domain/entity"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

func (u *UseCase) CreateTask(ctx context.Context, task entity.Task, tagsIDs []int64) (*entity.Task, error) {
	const op = "usecase.CreateTask"

	createdTask, err := u.DB.CreateTask(ctx, task, tagsIDs)
	if err != nil {
		u.Log.Error("error creating task", "op", op, "error", err)
		return nil, err
	}

	return createdTask, nil
}

func (u *UseCase) UpdateTask(ctx context.Context, task entity.UpdateTaskRequest, userID uuid.UUID, taskID int64) (*entity.Task, error) {
	const op = "usecase.UpdateTask"

	updatedTask, err := u.DB.UpdateTask(ctx, task, userID, taskID)
	if err != nil {
		u.Log.Error("error updating task", "op", op, "error", err)
		return nil, err
	}

	return updatedTask, nil
}

func (u *UseCase) DeleteTask(ctx context.Context, userID uuid.UUID, taskID int64) (*entity.Task, error) {
	const op = "usecase.DeleteTask"

	if userID == uuid.Nil {
		u.Log.Error("invalid userID", "op", op, "userID", userID, "taskID", taskID)
		return nil, errors.New("invalid userID")
	}

	if taskID <= 0 {
		u.Log.Error("invalid taskID", "op", op, "taskID", taskID)
		return nil, errors.New("invalid taskID")
	}

	deletedTask, err := u.DB.DeleteTask(ctx, userID, taskID)
	if err != nil {
		u.Log.Error("error deleting task", "op", op, "error", err)
		return nil, err
	}

	return deletedTask, nil
}

func (u *UseCase) GetTask(ctx context.Context, taskID int64) (*entity.Task, error) {
	const op = "usecase.GetTask"

	if taskID <= 0 {
		u.Log.Error("invalid taskID", "op", op, "taskID", taskID)
		return nil, errors.New("invalid taskID")
	}

	task, err := u.DB.GetTask(ctx, taskID)
	if err != nil {
		u.Log.Error("error getting task", "op", op, "error", err)
		return nil, err
	}

	return task, nil
}

func (u *UseCase) GetListTask(ctx context.Context, userID uuid.UUID, priority *int, tagID *int64, from, to *time.Time, limit, offset uint64) ([]*entity.UserTasksTab, error) {
	const op = "usecase.GetTasks"

	var tasks []*entity.UserTasksTab
	var err error

	if userID == uuid.Nil {
		u.Log.Error("invalid userID", "op", op, "userID", userID, "priority", priority, "tag", tagID)
		return nil, errors.New("invalid userID")
	}

	taskFilter := &entity.TasksFilter{}
	taskFilter.UserID = userID

	if priority != nil {
		taskFilter.Priority = priority
		taskFilter.FilterType = entity.PriorityFilter
	} else if tagID != nil {
		taskFilter.Tag = tagID
		taskFilter.FilterType = entity.TagFilter
	} else if from != nil {
		taskFilter.From = from

		if to != nil {
			taskFilter.To = to
		} else {
			return nil, errors.New("invalid data")
		}

		taskFilter.FilterType = entity.DateFilter
	} else {
		taskFilter.FilterType = entity.None
	}

	switch taskFilter.FilterType {
	case entity.TagFilter:
		tasks, err = u.DB.GetTasksByTag(ctx, taskFilter.UserID, *taskFilter.Tag, limit, offset)
		if err != nil {
			u.Log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	case entity.PriorityFilter:
		tasks, err = u.DB.GetTasksByPriority(ctx, taskFilter.UserID, *taskFilter.Priority, limit, offset)
		if err != nil {
			u.Log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	case entity.DateFilter:
		tasks, err = u.DB.GetTasksByDate(ctx, taskFilter.UserID, *taskFilter.From, *taskFilter.To, limit, offset)
		if err != nil {
			u.Log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	case entity.None:
		tasks, err = u.DB.ListTasks(ctx, taskFilter.UserID, limit, offset)
		if err != nil {
			u.Log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	}

	return tasks, nil
}

func (u *UseCase) ArchiveTask(ctx context.Context, userID uuid.UUID, id int64) (*entity.Task, error) {
	const op = "usecase.ArchiveTask"

	task, err := u.DB.ArchiveTask(ctx, userID, id)
	if err != nil {
		u.Log.Error("cannot archive task", "user_id", userID, "operation", op, "error", err)

		return nil, err
	}

	return task, err
}
