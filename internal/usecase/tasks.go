package usecase

import (
	"bernard/internal/domain/entity"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func (u *UseCase) CreateTask(ctx context.Context, task entity.Task) (*entity.Task, error) {
	const op = "usecase.CreateTask"

	createdTask, err := u.DB.CreateTask(ctx, task)
	if err != nil {
		u.Log.Error("error creating task", "op", op, "error", err)
		return nil, err
	}

	return createdTask, nil
}

func (u *UseCase) UpdateTask(ctx context.Context, task entity.Task) (*entity.Task, error) {
	const op = "usecase.UpdateTask"

	updatedTask, err := u.DB.UpdateTask(ctx, task)
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

func (u *UseCase) GetListTask(ctx context.Context, userID uuid.UUID, priority, tag, from, to string) ([]*entity.UserTasksTab, error) {
	const op = "usecase.GetTasks"

	var tasks []*entity.UserTasksTab
	var err error

	if userID == uuid.Nil {
		u.Log.Error("invalid userID", "op", op, "userID", userID, "priority", priority, "tag", tag)
		return nil, errors.New("invalid userID")
	}

	taskFilter := &entity.TasksFilter{}

	if priority != "" {
		taskFilter.Priority, err = strconv.Atoi(priority)
		if err != nil {
			u.Log.Error("invalid priority", "op", op, "error", err)
			return nil, errors.New("invalid priority")
		}

		taskFilter.FilterType = entity.Priority
	} else if tag != "" {
		taskFilter.Tag = tag
		taskFilter.FilterType = entity.Tag
	} else if from != "" {
		taskFilter.From, err = time.Parse(time.DateTime, from)
		if err != nil {
			u.Log.Error("invalid from", "op", op, "error", err)
			return nil, errors.New("invalid from")
		}

		taskFilter.To, err = time.Parse(time.RFC3339, to)
		if err != nil {
			u.Log.Error("invalid to", "op", op, "error", err)
			return nil, errors.New("invalid to")
		}

		taskFilter.FilterType = entity.Date
	} else {
		taskFilter.FilterType = entity.None
	}

	switch taskFilter.FilterType {
	case entity.Tag:
		tasks, err = u.DB.GetTasksByTag(ctx, taskFilter.UserID, taskFilter.Tag)
		if err != nil {
			u.Log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	case entity.Priority:
		tasks, err = u.DB.GetTasksByPriority(ctx, taskFilter.UserID, taskFilter.Priority)
		if err != nil {
			u.Log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	case entity.Date:
		tasks, err = u.DB.GetTasksByDate(ctx, taskFilter.UserID, taskFilter.From, taskFilter.To)
		if err != nil {
			u.Log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	case entity.None:
		tasks, err = u.DB.ListTasks(ctx, taskFilter.UserID)
		if err != nil {
			u.Log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	}

	return tasks, nil
}
