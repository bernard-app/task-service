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

func (u *UseCase) UpdateTask(ctx context.Context, task entity.UpdateTaskRequest, userID uuid.UUID, taskID int64) (*entity.Task, error) {
	const op = "usecase.UpdateTask"

	var updatedTask *entity.Task

	err := u.Tx.ReadWrite(ctx, func(ctxTx context.Context) error {
		var err error

		updatedTask, err = u.DB.UpdateTask(ctxTx, task, userID, taskID)
		if err != nil {
			return err
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

	updatedTask.Tags, err = u.DB.GetTaskTags(ctx, updatedTask.ID)
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

func (u *UseCase) GetGroupTasks(ctx context.Context, groupID int64, userID uuid.UUID, limit, offset uint64) ([]*entity.Task, error) {
	const op = "usecase.GetGroupTasks"

	tasks, err := u.DB.GetGroupTasks(ctx, groupID, userID, limit, offset)
	if err != nil {
		u.Log.Error("error getting group tasks", "op", op, "error", err)

		return nil, err
	}

	for _, task := range tasks {
		task.Tags, err = u.DB.GetTaskTags(ctx, task.ID)
		if err != nil {
			u.Log.Error("error gettig task tags", "op", op, "error", err)

			return nil, err
		}
	}
	return tasks, nil
}

func (u *UseCase) GetListTask(ctx context.Context, userID uuid.UUID, priority *int, tagID *int64, from, to *time.Time, limit, offset uint64) ([]*entity.Task, error) {
	const op = "usecase.GetTasks"

	var tasks []*entity.Task
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
	}

	if tagID != nil {
		taskFilter.Tag = tagID
		taskFilter.FilterType = entity.TagFilter
	}

	if from != nil {
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
