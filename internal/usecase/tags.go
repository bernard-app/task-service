package usecase

import (
	"bernard/internal/domain/entity"
	"context"
	"fmt"
	
	"github.com/google/uuid"
)

func (u *UseCase) CreateTag(ctx context.Context, tag entity.Tag) (*entity.Tag, error) {
	const op = "usecase.CreateTag"

	createdTag, err := u.DB.CreateTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return createdTag, nil
}

func (u *UseCase) UpdateTag(ctx context.Context, tagID int64, tagName, tagColor *string, userID uuid.UUID) (*entity.Tag, error) {
	const op = "usecase.UpdateTag"

	updatedTag, err := u.DB.UpdateTag(ctx, tagID, tagName, tagColor, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return updatedTag, nil
}

func (u *UseCase) DeleteTag(ctx context.Context, tagID int64, userID uuid.UUID) error {
	const op = "usecase.DeleteTag"

	err := u.DB.DeleteTag(ctx, tagID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *UseCase) GetTag(ctx context.Context, tagID int64, userID uuid.UUID) (*entity.Tag, error) {
	const op = "usecase.GetTag"

	tag, err := u.DB.GetTag(ctx, tagID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tag, nil
}

func (u *UseCase) GetTagList(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Tag, error) {
	const op = "usecase.GetTagList"

	tags, err := u.DB.GetTagList(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tags, nil
}

func (u *UseCase) GetTaskTags(ctx context.Context, taskID int64) ([]*entity.Tag, error) {
	const op = "usecase.GetTaskTags"

	tags, err := u.DB.GetTaskTags(ctx, taskID)
	if err != nil {
		u.Log.Error("error getting task tags", "op", op, "error", err)
		return nil, err
	}

	return tags, nil
}

func (u *UseCase) AddTagsToTask(ctx context.Context, tagsIDs []int64, taskID int64) error {
	const op = "usecase.AddTagToTask"

	err := u.DB.AddTagsToTask(ctx, tagsIDs, taskID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *UseCase) RemoveTagsFromTask(ctx context.Context, tagsIDs []int64, taskID int64) error {
	const op = "usecase.RemoveTagFromTask"

	err := u.DB.RemoveTagsFromTask(ctx, tagsIDs, taskID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}