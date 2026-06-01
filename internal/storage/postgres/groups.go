package postgres

import (
	"bernard/internal/domain/entity"
	"bernard/pkg/response"
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateGroup(ctx context.Context, group entity.Group) (*entity.Group, error) {
	const op = "storage.CreateGroup"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Insert("groups").
		Columns("name, project_id, user_id").
		Values(&group.Name, group.ProjectID, group.UserID).
		Suffix("RETURNING id, name, project_id, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var createdGroup entity.Group

	err = tx.QueryRow(ctx, query, args...).Scan(&createdGroup.ID, &createdGroup.Name, &createdGroup.ProjectID, &createdGroup.UserID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%s: %w", op, response.ErrAlreadyExists)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &createdGroup, nil
}

func (s *Storage) UpdateGroup(ctx context.Context, group entity.UpdateGroupRequest, userID uuid.UUID, groupID int64) (*entity.Group, error) {
	const op = "storage.UpdateGroup"

	tx := s.getEngine(ctx)

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
		return nil, fmt.Errorf("%s: cannot update group: %s", op, "no update available")
	}

	query, args, err := builder.
		Suffix("RETURNING id, name, project_id, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var UpdatedGroup entity.Group

	err = tx.QueryRow(ctx, query, args...).
		Scan(&UpdatedGroup.ID, &UpdatedGroup.Name, &UpdatedGroup.ProjectID, &UpdatedGroup.UserID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("%s: %w", op, response.ErrNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &UpdatedGroup, nil
}

func (s *Storage) DeleteGroup(ctx context.Context, groupID int64, userID uuid.UUID) error {
	const op = "storage.DeleteGroup"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Delete("groups").
		Where(sq.Eq{"id": groupID, "user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: cannot delete group: %s", op, err.Error())
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return response.ErrNotFound
	}

	return nil
}

func (s *Storage) GetGroup(ctx context.Context, userID uuid.UUID, groupID int64) (*entity.Group, error) {
	const op = "storage.GetGroup"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Select("id", "name", "project_id").
		From("groups").
		Where(sq.Eq{"user_id": userID, "id": groupID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var group entity.Group

	err = tx.QueryRow(ctx, query, args...).Scan(&group.ID, &group.Name, &group.ProjectID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("%s: %w", op, response.ErrNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &group, nil
}

func (s *Storage) GetProjectGroups(ctx context.Context, userID uuid.UUID, projectID int64, limit, offset uint64) ([]*entity.Group, error) {
	const op = "storage.GetProjectGroup"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Select("id", "name", "project_id").
		From("groups").
		Where(sq.Eq{"project_id": projectID, "user_id": userID}).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var groups []*entity.Group

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	for rows.Next() {
		var group entity.Group

		err = rows.Scan(&group.ID, &group.Name, &group.ProjectID)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, fmt.Errorf("%s: %w", op, response.ErrNotFound)
			}

			return nil, fmt.Errorf("%s: %w", op, err)
		}

		groups = append(groups, &group)
	}

	return groups, nil
}

func (s *Storage) CheckGroupOwnership(ctx context.Context, userID uuid.UUID, groupID int64) (bool, error) {
	const op = "storage.CheckGroupOwnership"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Select("1").
		From("groups").
		Where(sq.Eq{"id": groupID, "user_id": userID}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	var dummy int
	err = tx.QueryRow(ctx, query, args...).Scan(&dummy)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}
