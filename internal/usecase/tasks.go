package usecase

import (
	"bernard/internal/domain/entity"
	"bernard/pkg/response"
	"context"

	"github.com/google/uuid"
)

func (u *UseCase) CreateTask(ctx context.Context, task entity.Task, tagsIDs []int64) (*entity.Task, error) {
	const op = "usecase.CreateTask"

	if task.ProjectID != nil {
		ok, err := u.DB.CheckProjectOwnership(ctx, task.UserID, *task.ProjectID)
		if !ok || err != nil {
			u.Log.Warn("permission denied", "op", op, "error", err)

			return nil,	response.ErrPermissionDenied
		}
	}

	if task.GroupID != nil {
		ok, err := u.DB.CheckGroupOwnership(ctx, task.UserID, *task.GroupID)
		if !ok || err != nil {
			u.Log.Warn("permission denied", "op", op, "error", err)

			return nil, response.ErrPermissionDenied
		}
	}
	
	var createdTask *entity.Task

	err := u.Tx.ReadWrite(ctx, func(ctxTx context.Context) error {
		var err error

		createdTask, err = u.DB.CreateTask(ctxTx, task)
		if err != nil {
			return err
		}

		err = u.DB.AddTagsToTask(ctxTx, tagsIDs, createdTask.ID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		u.Log.Error("error creating task", "op", op, "error", err)

		return nil, err
	}

	return createdTask, nil
}

func HasBaseUpdates(req entity.UpdateTaskRequest) bool {
	return req.Name != nil ||
		req.Description != nil ||
		req.Priority != nil ||
		req.Status != nil ||
		req.Deadline != nil ||
		req.GroupID != nil ||
		req.ProjectID != nil
}

func (u *UseCase) UpdateTask(ctx context.Context, task entity.UpdateTaskRequest, userID uuid.UUID, taskID int64) (*entity.Task, error) {
	const op = "usecase.UpdateTask"

	if task.ProjectID != nil {
		ok, err := u.DB.CheckProjectOwnership(ctx, userID, *task.ProjectID)
		if !ok || err != nil {
			u.Log.Warn("permission denied", "op", op, "error", err)

			return nil, response.ErrPermissionDenied
		}
	}

	if task.GroupID != nil {
		ok, err := u.DB.CheckGroupOwnership(ctx, userID, *task.GroupID)
		if !ok || err != nil {
			u.Log.Warn("permission denied", "op", op, "error", err)

			return nil, response.ErrPermissionDenied
		}
	}
	
	updatedTask := &entity.Task{}

	err := u.Tx.ReadWrite(ctx, func(ctxTx context.Context) error {
		var err error

		if HasBaseUpdates(task) {
			updatedTask, err = u.DB.UpdateTask(ctxTx, task, userID, taskID)
			if err != nil {
				return err
			}
		} else {
			updatedTask, err = u.DB.GetTask(ctxTx, taskID, userID)
			if err != nil {
				return err
			}
		}

		if task.TagsIDs != nil {
			err = u.DB.RemoveAllTagsFromTask(ctxTx, taskID)
			if err != nil {
				return err
			}

			err = u.DB.AddTagsToTask(ctxTx, *task.TagsIDs, taskID)
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	if err != nil {
		u.Log.Error("error updating task", "op", op, "error", err)

		return nil, err
	}

	updatedTask.Tags, err = u.DB.GetTaskTags(ctx, taskID)
	if err != nil {
		u.Log.Error("error getting task tags", "op", op, "error", err)

		return nil, err
	}

	return updatedTask, nil
}

func (u *UseCase) DeleteTask(ctx context.Context, userID uuid.UUID, taskID int64) error {
	const op = "usecase.DeleteTask"

	err := u.DB.DeleteTask(ctx, userID, taskID)
	if err != nil {
		u.Log.Error("error deleting task", "op", op, "error", err)

		return err
	}

	return nil
}

func (u *UseCase) GetTask(ctx context.Context, taskID int64, userID uuid.UUID) (*entity.Task, error) {
	const op = "usecase.GetTask"

	task, err := u.DB.GetTask(ctx, taskID, userID)
	if err != nil {
		u.Log.Error("error getting task", "op", op, "error", err)

		return nil, err
	}

	task.Tags, err = u.DB.GetTaskTags(ctx, task.ID)
	if err != nil {
		u.Log.Error("error getting task tags", "op", op, "error", err)

		return nil, err
	}

	return task, nil
}

func (u *UseCase) GetListTask(ctx context.Context, taskFilter entity.TasksFilter) ([]*entity.Task, error) {
	const op = "usecase.GetTasks"

	var tasks []*entity.Task
	var err error

	if taskFilter.UserID == uuid.Nil {
		u.Log.Error("invalid userID", "op", op)

		return nil, response.ErrInvalidArgument
	}

	taskFilter.FilterType = entity.None
	
	if taskFilter.Priority != nil {
		taskFilter.FilterType = entity.PriorityFilter
	} 

	if taskFilter.Tag != nil {
		taskFilter.FilterType = entity.TagFilter
	} 

	if (taskFilter.From != nil && taskFilter.To == nil) || (taskFilter.From == nil && taskFilter.To != nil) {
		return nil, response.ErrInvalidArgument
	}

	if taskFilter.From != nil && taskFilter.To != nil {
		taskFilter.FilterType = entity.DateFilter
	}
	
	switch taskFilter.FilterType {
	case entity.TagFilter:
		tasks, err = u.DB.GetTasksByTag(ctx, taskFilter.UserID, *taskFilter.Tag, taskFilter.Limit, taskFilter.Offset)
		if err != nil {
			u.Log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	case entity.PriorityFilter:
		tasks, err = u.DB.GetTasksByPriority(ctx, taskFilter.UserID, *taskFilter.Priority, taskFilter.Limit, taskFilter.Offset)
		if err != nil {
			u.Log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	case entity.DateFilter:
		tasks, err = u.DB.GetTasksByDate(ctx, taskFilter.UserID, *taskFilter.From, *taskFilter.To, taskFilter.Limit, taskFilter.Offset)
		if err != nil {
			u.Log.Error("error getting tasks", "op", op, "error", err)

			return nil, err
		}
	case entity.None:
		tasks, err = u.DB.ListTasks(ctx, taskFilter.UserID, taskFilter.Limit, taskFilter.Offset)
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

func extractTaskID(tasks []*entity.Task) []int64 {
	ids := make([]int64, 0, len(tasks))

	for _, task := range tasks {
		ids = append(ids, task.ID)
	}

	return ids
}