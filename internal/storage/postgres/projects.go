package postgres

import (
	"bernard/internal/domain/entity"
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateProject(ctx context.Context, project entity.Project) (*entity.Project, error) {
	const op = "storage.CreateProject"

	tx := s.getEngine(ctx)
	
	query, args, err := sq.
		Insert("projects").
		Columns("name", "description", "user_id").
		Values(&project.Name, &project.Description, &project.UserID).
		Suffix("RETURNING id, name, description, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var createdProject entity.Project

	err = tx.QueryRow(ctx, query, args...).Scan(&createdProject.ID, &createdProject.Name, &createdProject.Description, &createdProject.UserID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%s: project already exists", op)
		}

		return nil, fmt.Errorf("%s: cannot create project: %s", op, err.Error())
	}

	return &createdProject, nil
}

func (s *Storage) UpdateProject(ctx context.Context, project entity.UpdateProjectRequest, projectID int64, userID uuid.UUID) (*entity.Project, error) {
	const op = "storage.UpdateProject"

	tx := s.getEngine(ctx)
	
	builder := sq.
		Update("projects").
		Where(sq.Eq{"id": projectID, "user_id": userID})

	hasUpdate := false

	if project.Name != nil {
		builder = builder.Set("name", *project.Name)
		hasUpdate = true
	}

	if project.Description != nil {
		builder = builder.Set("description", *project.Description)
		hasUpdate = true
	}

	if !hasUpdate {
		return nil, fmt.Errorf("%s: nothing to update", op)
	}

	query, args, err := builder.
		Suffix("RETURNING id, name, description, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var updatedProject entity.Project

	err = tx.QueryRow(ctx, query, args...).Scan(&updatedProject.ID, &updatedProject.Name, updatedProject.Description, &updatedProject.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot update project: %s", op, err.Error())
	}

	return &updatedProject, nil
}

func (s *Storage) DeleteProject(ctx context.Context, projectID int64, userID uuid.UUID) error {
	const op = "storage.DeleteProject"

	tx := s.getEngine(ctx)
	
	query, args, err := sq.
		Delete("groups").
		Where(sq.Eq{"id": projectID, "user_id": userID}).
		Suffix("RETURNING id, name, description, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var deletedProject entity.Project

	err = tx.QueryRow(ctx, query, args...).Scan(&deletedProject.ID, &deletedProject.Name, deletedProject.Description, &deletedProject.UserID)
	if err != nil {
		return fmt.Errorf("%s: cannot delete project: %s", op, err.Error())
	}

	return nil
}

func (s *Storage) GetProject(ctx context.Context, projectID int64, userID uuid.UUID) (*entity.Project, error) {
	const op = "storage.GetProject"

	tx := s.getEngine(ctx)
	
	query, args, err := sq.
		Select("id", "name", "description", "user_id",).
		From("projects").
		Where(sq.Eq{"id": projectID, "user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var project entity.Project
	
	err = tx.QueryRow(ctx, query, args...).Scan(&project.ID, &project.Name, &project.Description, &project.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot query row: %w", op, err)
	}

	return &project, nil
}

func (s *Storage) GetProjects(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Project, error) {
	const op = "storage.GetProjects"

	tx := s.getEngine(ctx)
	
	query, args, err := sq.
		Select("id", "name", "description", "user_id").
		From("projects").
		Where(sq.Eq{"user_id": userID}).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var projects []*entity.Project

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot read projects: %s", op, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var project entity.Project
		
		err:= rows.Scan(&project.ID, &project.Name, &project.Description, &project.UserID)
		if err != nil {
			return nil, fmt.Errorf("%s: cannot read project: %s", op, err.Error())
		}

		projects = append(projects, &project)
	}

	return projects, nil
}
