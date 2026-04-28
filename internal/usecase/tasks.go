package usecase

import (
	"bernard/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

func (u *UseCase) CreateTask(ctx context.Context, task entity.Task) (*entity.Task, error) {
	const op = "usecase.CreateTask"

	createdTask, err := u.db.CreateTask(ctx, task)
	if err != nil {
		u.log.Error("error creating task", "op", op, "error", err)
		return nil, err
	}

	return createdTask, nil
}

func (u *UseCase) UpdateTask(ctx context.Context, userID uuid.UUID, task entity.Task) (*entity.Task, error) {
	const op = "usecase.UpdateTask"

	updatedTask, err := u.db.UpdateTask(ctx, userID, task)
	if err != nil {
		u.log.Error("error updating task", "op", op, "error", err)
		return nil, err
	}

	return updatedTask, nil
}

func (u *UseCase) DeleteTask(ctx context.Context, userID uuid.UUID, taskID int64) (*entity.Task, error) {
	const op = "usecase.DeleteTask"

	deletedTask, err := u.db.DeleteTask(ctx, userID, taskID)
	if err != nil {
		u.log.Error("error deleting task", "op", op, "error", err)
		return nil, err
	}

	return deletedTask, nil
}

func (u *UseCase) GetTask(ctx context.Context, taskID int64) (*entity.Task, error) {
	const op = "usecase.GetTask"

	task, err := u.db.GetTask(ctx, taskID)
	if err != nil {
		u.log.Error("error getting task", "op", op, "error", err)
		return nil, err
	}

	return task, nil
}

func (u *UseCase) GetListTask(ctx context.Context, tasksFilter entity.TasksFilter) ([]*entity.UserTasksTab, error) {
	const op = "usecase.GetTasks"

	var tasks []*entity.UserTasksTab
	var err error

	switch tasksFilter.FilterType {
	case entity.Tag:
		tasks, err = u.db.GetTasksByTag(ctx, tasksFilter.UserID, tasksFilter.Tag)
		if err != nil {
			u.log.Error("error getting tasks", "op", op, "error", err)
		
			return nil, err
		}
	case entity.Priority:
		tasks, err = u.db.GetTasksByPriority(ctx, tasksFilter.UserID, tasksFilter.Priority)
		if err != nil {
			u.log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	case entity.Date:
		tasks, err = u.db.GetTasksByDate(ctx, tasksFilter.UserID, tasksFilter.From, tasksFilter.To)
		if err != nil {
			u.log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	default:
		tasksFilter.FilterType = entity.None
		tasks, err = u.db.ListTasks(ctx, tasksFilter.UserID)
		if err != nil {
			u.log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	}

	return tasks, nil
}
