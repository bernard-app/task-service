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
		Columns("name, project_id, user_id").
		Values(&group.Name, group.ProjectID, group.UserID).
		Suffix("RETURNING id, name, project_id, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var createdGroup entity.Group

	err = s.DB.QueryRow(ctx, query, args...).Scan(&createdGroup.ID, &createdGroup.Name, &createdGroup.ProjectID, &createdGroup.UserID)
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
		Update("groups").
		Where(sq.Eq{"id": groupID, "user_id": userID})

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
		Suffix("RETURNING id, name, project_id, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var UpdatedGroup entity.Group

	err = s.DB.QueryRow(ctx, query, args...).
		Scan(&UpdatedGroup.ID, &UpdatedGroup.Name, &UpdatedGroup.ProjectID, &UpdatedGroup.UserID)
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

func (s *Storage) DeleteGroup(ctx context.Context, groupID int64, userID uuid.UUID) error {
	const op = "storage.DeleteGroup"

	query, args, err := sq.
		Delete("groups").
		Where(sq.Eq{"id": groupID, "user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	result, err := s.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: cannot delete group: %s", op, err.Error())
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("%s: group does not exist", op)
	}

	return nil
}
