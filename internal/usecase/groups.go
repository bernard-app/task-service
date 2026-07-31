package usecase

import (
	"bernard/internal/domain/entity"
	"bernard/pkg/response"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (u *UseCase) CreateGroup(ctx context.Context, group entity.Group) (*entity.Group, error) {
	const op = "usecase.CreateGroup"

	ok, err := u.DB.CheckProjectOwnership(ctx, group.UserID, group.ProjectID)
	if !ok || err != nil {
		u.Log.Warn("permission denied", "op", op, "error", err)

		return nil, fmt.Errorf("%s: %w", op, response.ErrPermissionDenied)
	}
	
	createdGroup, err := u.DB.CreateGroup(ctx, group)
	if err != nil {
		u.Log.Error("error while creating group", "op", op, "error", err)
		
		return nil, err
	}

	return createdGroup, nil
}

func (u *UseCase) UpdateGroup(ctx context.Context, group entity.UpdateGroupRequest, userID uuid.UUID, groupID int64) (*entity.Group, error) {
	const op = "usecase.UpdateGroup"

	if group.ProjectID != nil {
		ok, err := u.DB.CheckProjectOwnership(ctx, userID, *group.ProjectID)
		if !ok || err != nil {
			u.Log.Warn("permission denied", "op", op, "error", err)

			return nil, fmt.Errorf("%s: %w", op, response.ErrPermissionDenied)
		}
	}
	
	updatedGroup, err := u.DB.UpdateGroup(ctx, group, userID, groupID)
	if err != nil {
		u.Log.Error("error while updating group", "op", op, "error", err)
		
		return nil, err
	}

	return updatedGroup, nil
}

func (u *UseCase) DeleteGroup(ctx context.Context, groupID int64, userID uuid.UUID) error {
	const op = "usecase.DeleteGroup"

	if groupID <= 0 {
		u.Log.Error("invalid group id")

		return response.ErrInvalidArgument
	}

	err := u.DB.DeleteGroup(ctx, groupID, userID)
	if err != nil {
		u.Log.Error("error while deleting group", "op", op, "error", err)
		
		return err
	}

	return nil
}

func (u *UseCase) GetGroup(ctx context.Context, groupID int64, userID uuid.UUID) (*entity.Group, error) {
	const op = "usecase.GetGroup"
		
	group, err := u.DB.GetGroup(ctx, userID, groupID)
	if err != nil {
		u.Log.Error("cannot get group", "op", op, "error", err)
		
		return nil, err
	}

	tasks, err := u.DB.GetTasksByGroupIDs(ctx, []int64{group.ID})
	if err != nil {
		u.Log.Error("error getting tasks", "op", op, "error", err)
		
		return nil, err
	}

	tags, err := u.DB.GetTagsByTaskIDs(ctx, extractTaskID(tasks))
	if err != nil {
		u.Log.Error("error getting tags", "op", op, "error", err)

		return nil, err
	}

	tagsByTask := make(map[int64][]*entity.Tag)
	for _, tag := range tags {
	    tagsByTask[tag.TaskID] = append(tagsByTask[tag.TaskID], tag)
	}
	
	tasksByGroup := make(map[int64][]*entity.Task)
	for _, task := range tasks {
	    task.Tags = tagsByTask[task.ID]

		if task.GroupID != nil {
            tasksByGroup[*task.GroupID] = append(tasksByGroup[*task.GroupID], task)
        }
	}

	group.Tasks = tasksByGroup[group.ID]
	
	return group, nil
}

func (u *UseCase) GetProjectGroups(ctx context.Context, projectID int64, userID uuid.UUID, limit, offset uint64) ([]*entity.Group, error) {
	const op = "usecase.GetGroups"

	ok, err := u.DB.CheckProjectOwnership(ctx, userID, projectID)
	if !ok || err != nil {
		u.Log.Warn("permission denied", "op", op, "error", err)

		return nil, fmt.Errorf("%s: %w", op, response.ErrPermissionDenied)
	}
	
	groups, err := u.DB.GetProjectGroups(ctx, userID, projectID, limit, offset)
	if err != nil {
		u.Log.Error("cannot get groups", "op", op, "error", err)
		
		return nil, err 
	}

	return groups, nil
}

func extractGroupIDs(groups []*entity.Group) []int64 {
	ids := make([]int64, 0, len(groups))

	for _, group := range groups {
		ids = append(ids, group.ID)
	}

	return ids
}