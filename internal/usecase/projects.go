package usecase

import (
	"bernard/internal/domain/entity"
	"context"
	"errors"

	"github.com/google/uuid"
)

func (u *UseCase) CreateProject(ctx context.Context, project entity.Project) (*entity.Project, error) {
	const op = "storage.CreateProject"

	createdProject, err := u.DB.CreateProject(ctx, project)
	if err != nil {
		u.Log.Error("error creating project", "op", op, "error", err)
		return nil, err
	}

	return createdProject, nil
}

func (u *UseCase) UpdateProject(ctx context.Context, project entity.UpdateProjectRequest, projectID int64, userID uuid.UUID) (*entity.Project, error) {
	const op = "storage.UpdateProject"

	updatedProject, err := u.DB.UpdateProject(ctx, project, projectID, userID)
	if err != nil {
		u.Log.Error("error updating project", "op", op, "error", err)
		return nil, err
	}

	return updatedProject, nil
}

func (u *UseCase) DeleteProject(ctx context.Context, projectID int64, userID uuid.UUID) error {
	const op = "storage.DeleteProject"

	if projectID <= 0 {
		u.Log.Error("invalid projectID", "op", op, "projectID", projectID)
		return errors.New("invalid project id")
	}

	err := u.DB.DeleteProject(ctx, projectID, userID)
	if err != nil {
		u.Log.Error("error deleting project", "op", op, "error", err)
		return err
	}

	return nil
}

func (u *UseCase) GetProjectTree(ctx context.Context, projectID int64, userID uuid.UUID) (*entity.Project, error) {
	const op = "storage.GetProject"

	if projectID <= 0 {
		u.Log.Error("invalid projectID", "op", op, "projectID", projectID)
		return nil, errors.New("invalid project id")
	}

	project, err := u.DB.GetProjectTree(ctx, projectID, userID)
	if err != nil {
		u.Log.Error("error getting project", "op", op, "error", err)
		return nil, err
	}

	return project, nil
}

func (u *UseCase) GetProjects(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Project, error) {
	const op = "storage.GetProjects"

	if userID == uuid.Nil {
		u.Log.Error("invalid userID", "op", op, "userID", userID)
		return nil, errors.New("invalid userID")
	}

	projects, err := u.DB.GetProjects(ctx, userID, limit, offset)
	if err != nil {
		u.Log.Error("error getting projects", "op", op, "error", err)
		return nil, err
	}

	return projects, nil
}
