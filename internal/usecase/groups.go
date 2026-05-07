package usecase

import (
	"bernard/internal/domain/entity"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (u *UseCase) CreateGroup(ctx context.Context, group entity.Group) (*entity.Group, error) {
	const op = "usecase.CreateGroup"

	createdGroup, err := u.DB.CreateGroup(ctx, group)
	if err != nil {
		u.Log.Error("error while creating group", "op", op, "error", err)
		
		return nil, err
	}

	return createdGroup, nil
}

func (u *UseCase) UpdateGroup(ctx context.Context, group entity.UpdateGroupRequest, userID uuid.UUID, groupID int64) (*entity.Group, error) {
	const op = "usecase.UpdateGroup"

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

		return fmt.Errorf("invalid group id")
	}

	err := u.DB.DeleteGroup(ctx, groupID, userID)
	if err != nil {
		u.Log.Error("error while deleting group", "op", op, "error", err)
		
		return err
	}

	return nil
}

func (u *UseCase) GetGroup(ctx context.Context, groupID int64, userID uuid.UUID, limit, offset uint64) (*entity.Group, error) {
	const op = "usecase.GetGroup"
		
	group, err := u.DB.GetGroup(ctx, userID, groupID)
	if err != nil {
		u.Log.Error("cannot get group", "op", op, "error", err)
		return nil, err
	}

	group.Tasks, err = u.GetGroupTasks(ctx, group.ID, userID, limit, offset)
	if err != nil {
		u.Log.Error("cannot get group tasks", "op", op, "error", err)
		return nil, err
	}
	
	return group, nil
}

func (u *UseCase) GetProjectGroups(ctx context.Context, projectID int64, userID uuid.UUID, limit, offset uint64) ([]*entity.Group, error) {
	const op = "usecase.GetGroups"

	groups, err := u.DB.GetProjectGroups(ctx, userID, projectID, limit, offset)
	if err != nil {
		u.Log.Error("cannot get groups", "op", op, "error", err)
		return nil, err 
	}

	for _, group := range groups {
		group.Tasks, err = u.GetGroupTasks(ctx, group.ID, userID, limit, offset)
		if err != nil {
			u.Log.Error("error getting group tasks", "op", op, "error", err)
			return nil, err
		}
	}

	return groups, nil
}