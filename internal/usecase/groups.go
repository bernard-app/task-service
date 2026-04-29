package usecase

import (
	"bernard/internal/domain/entity"
	"context"
	"errors"
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

func (u *UseCase) UpdateGroup(ctx context.Context, group entity.Group) (*entity.Group, error) {
	const op = "usecase.UpdateGroup"

	updatedGroup, err := u.DB.UpdateGroup(ctx, group)
	if err != nil {
		u.Log.Error("error while updating group", "op", op, "error", err)
		return nil, err
	}

	return updatedGroup, nil
}

func (u *UseCase) DeleteGroup(ctx context.Context, groupID int64) (*entity.Group, error) {
	const op = "usecase.DeleteGroup"

	if groupID <= 0 {
		return nil, errors.New("invalid group id")
	}

	deletedGroup, err := u.DB.DeleteGroup(ctx, groupID)
	if err != nil {
		u.Log.Error("error while deleting group", "op", op, "error", err)
		return nil, err
	}

	return deletedGroup, nil
}
