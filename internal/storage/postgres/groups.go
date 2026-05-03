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

func (s *Storage) CreateGroup(ctx context.Context, group entity.Group) (*entity.Group, error) {
	const op = "storage.CreateGroup"

	query, args, err := sq.
		Insert("groups").
		Columns("name, project_id").
		Values(&group.Name, group.ProjectID).
		Suffix("RETURNING id, name, project_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var createdGroup entity.Group

	err = s.DB.QueryRow(ctx, query, args...).Scan(&createdGroup.ID, &createdGroup.Name, &createdGroup.ProjectID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New(fmt.Sprintf("%s: group already exists", op))
		}

		return nil, errors.New(fmt.Sprintf("%s: cannot create group: %s", op, err.Error()))
	}

	return &createdGroup, nil
}

func (s *Storage) UpdateGroup(ctx context.Context, group entity.UpdateGroupRequest, userID uuid.UUID, groupID int64) (*entity.Group, error) {
	const op = "storage.UpdateGroup"

	builder := sq.
		Update("groups g").
		Where(sq.Eq{"id": groupID})

	hasUpdate := false

	if group.Name != nil {
		builder = builder.Set("name", *group.Name)
		hasUpdate = true
	}
	if group.ProjectID != nil {
		builder = builder.Set("project_id", *group.ProjectID)
		hasUpdate = true
	}

	if !hasUpdate {
		return nil, errors.New(fmt.Sprintf("%s: cannot update group: %s", op, "no update available"))
	}

	query, args, err := builder.
		Suffix("RETURNING id, name, project_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var UpdatedGroup entity.Group

	err = s.DB.QueryRow(ctx, query, args...).
		Scan(&UpdatedGroup.ID, &UpdatedGroup.Name, &UpdatedGroup.ProjectID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			if pgErr.ConstraintName == "groups_project_id_fkey" {
				return nil, fmt.Errorf("%s: project does not exist", op)
			}
		}

		return nil, errors.New(fmt.Sprintf("%s: cannot update group: %s", op, err.Error()))
	}

	return &UpdatedGroup, nil
}

func (s *Storage) DeleteGroup(ctx context.Context, groupID int64) (*entity.Group, error) {
	const op = "storage.DeleteGroup"

	query, args, err := sq.
		Delete("groups").
		Where(sq.Eq{"id": groupID}).
		Suffix("RETURNING id, name, project_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var deletedGroup entity.Group

	err = s.DB.QueryRow(ctx, query, args...).Scan(&deletedGroup.ID, &deletedGroup.Name, &deletedGroup.ProjectID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot delete group: %s", op, err.Error()))
	}

	return &deletedGroup, nil
}
