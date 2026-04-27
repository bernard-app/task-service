package usecase

import (
	"bernard/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

func (u *UseCase) CreateProject(ctx context.Context, project entity.Project) (*entity.Project, error) {
	const op = "storage.CreateProject"

	createdProject, err := u.db.CreateProject(ctx, project)
	if err != nil {
		u.log.Error("error creating project", "op", op, "error", err)
		return nil, err
	}

	return createdProject, nil
}

func (u *UseCase) UpdateProject(ctx context.Context, project entity.Project) (*entity.Project, error) {
	const op = "storage.UpdateProject"

	updatedProject, err := u.db.UpdateProject(ctx, project)
	if err != nil {
		u.log.Error("error updating project", "op", op, "error", err)
		return nil, err
	}

	return updatedProject, nil
}

func (u *UseCase) DeleteProject(ctx context.Context, projectID int64) (*entity.Project, error) {
	const op = "storage.DeleteProject"

	project, err := u.db.DeleteProject(ctx, projectID)
	if err != nil {
		u.log.Error("error deleting project", "op", op, "error", err)
		return nil, err
	}

	return project, nil
}

func (u *UseCase) GetProjectTree(ctx context.Context, projectID int64, userID uuid.UUID) (*entity.Project, error) {
	const op = "storage.GetProject"

	project, err := u.db.GetProjectTree(ctx, projectID, userID)
	if err != nil {
		u.log.Error("error getting project", "op", op, "error", err)
		return nil, err
	}

	return project, nil
}

func (u *UseCase) GetProjects(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Project, error) {
	const op = "storage.GetProjects"

	projects, err := u.db.GetProjects(ctx, userID, limit, offset)
	if err != nil {
		u.log.Error("error getting projects", "op", op, "error", err)
		return nil, err
	}

	return projects, nil
}
