package postgres

import (
	"bernard/internal/domain/entity"
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateGroup(ctx context.Context, group entity.Group) (*entity.Group, error) {
	const op = "storage.CreateGroup"

	query, args, err := sq.
		Insert("groups").
		Columns("name, project_id").
		Values(&group.Name, group.ProjectID).
		Prefix("RETURNING id, name, project_id").
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

func (s *Storage) UpdateGroup(ctx context.Context, group entity.Group) (*entity.Group, error) {
	const op = "storage.UpdateGroup"

	query, args, err := sq.
		Update("groups").
		SetMap(map[string]interface{}{"name": &group.Name, "project_id": group.ProjectID}).
		Where(sq.Eq{"id": group.ID}).
		Prefix("RETURNING id, name, project_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var updatedGroup entity.Group

	err = s.DB.QueryRow(ctx, query, args...).Scan(&updatedGroup.ID, &updatedGroup.Name, &updatedGroup.ProjectID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot update group: %s", op, err.Error()))
	}

	return &updatedGroup, nil
}

func (s *Storage) DeleteGroup(ctx context.Context, groupID int64) (*entity.Group, error) {
	const op = "storage.DeleteGroup"

	query, args, err := sq.
		Delete("groups").
		Where(sq.Eq{"id": groupID}).
		Prefix("RETURNING id, name, project_id").
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
