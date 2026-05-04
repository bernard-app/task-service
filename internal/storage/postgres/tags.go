package postgres

import (
	"bernard/internal/domain/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateTag(ctx context.Context, tag entity.Tag) (*entity.Tag, error) {
	const op = "storage.CreateTag"

	query, args, err := sq.
		Insert("tags").
		Columns("name", "color", "user_id").
		Values(tag.Name, tag.Color, tag.UserID).
		Suffix("RETURNING id, name, color, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	var createdTag entity.Tag

	err = s.DB.QueryRow(ctx, query, args...).Scan(&createdTag.ID, &createdTag.Name, &createdTag.Color, &createdTag.UserID)
	if err != nil {
		return nil, fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	return &createdTag, nil
}

func (s *Storage) UpdateTag(ctx context.Context, tagID int64, name *string, color *string, userID uuid.UUID) (*entity.Tag, error) {
	const op = "storage.UpdateTag"

	builder := sq.Update("tags")

	if name != nil {
		builder = builder.Set("name", *name)
	}

	if color != nil {
		builder = builder.Set("color", *color)
	}

	query, args, err := builder.
		Where(sq.And{
			sq.Eq{"id": tagID},
			sq.Eq{"user_id": userID},
		}).
		Suffix("RETURNING id, name, color, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	var updatedTag entity.Tag

	err = s.DB.QueryRow(ctx, query, args...).Scan(&updatedTag.ID, &updatedTag.Name, &updatedTag.Color, &updatedTag.UserID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf(`tag "%d" not found, err: %w, op: %s`, tagID, err, op)
		}

		return nil, fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	return &updatedTag, nil
}

func (s *Storage) DeleteTag(ctx context.Context, tagID int64, userID uuid.UUID) error {
	const op = "storage.DeleteTag"

	query, args, err := sq.
		Delete("tags").
		Where(sq.Eq{"id": tagID, "user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	err = s.DB.QueryRow(ctx, query, args...).Scan(&struct{}{})
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("tag does not exist")
		}

		return fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	return nil
}

func (s *Storage) GetTag(ctx context.Context, tagID int64, userID uuid.UUID) (*entity.Tag, error) {
	const op = "storage.GetTag"

	query, args, err := sq.
		Select("id", "name", "color", "user_id").
		From("tags").
		Where(sq.And{
			sq.Eq{"id": tagID},
			sq.Eq{"user_id": userID},
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	var tag entity.Tag

	err = s.DB.QueryRow(ctx, query, args...).Scan(&tag.ID, tag.Name, tag.Color, tag.UserID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf(`tag "%d" not found, err: %w, op: %s`, tagID, err, op)
		}

		return nil, fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	return &tag, nil
}

func (s *Storage) GetTagList(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Tag, error) {
	const op = "storage.GetTagList"

	query, args, err := sq.
		Select("id", "name", "color", "user_id").
		From("tags").
		Where(sq.Eq{"user_id": userID}).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	defer rows.Close()

	var tags []*entity.Tag

	for rows.Next() {
		var tag entity.Tag

		err = rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.UserID)
		if err != nil {
			return nil, fmt.Errorf("cannot scan rows. op: %s, error: %w", op, err)
		}

		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate over rows. op: %s, error: %w", op, err)
	}

	return tags, nil
}

func (s *Storage) AddTagToTask(ctx context.Context, tagID, taskID int64, userID uuid.UUID) error {
	const op = "storage.AddTagToTask"

	// Transaction start
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("cannot start transaction. op: %s, error: %w", op, err)
	}
	defer tx.Rollback(ctx)

	// Check for user's access
	query, args, err := sq.
		Select("1").
		From("tasks t, tags tag").
		Where(sq.Eq{"t.id": taskID, "t.user_id": userID, "tag.id": tagID, "tag.user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	var exists int
	err = tx.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: task not found or access denied", op)
		}
		return fmt.Errorf("%s: cannot check task ownership: %w", op, err)
	}

	// Adds tag
	query, args, err = sq.
		Insert("tasks_tags").
		Columns("task_id", "tag_id").
		Values(taskID, tagID).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf(`tag "%d" already exists, err: %w, op: %s`, tagID, err, op)
		}

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return fmt.Errorf(`tag "%d" or tasks "%d" does not exist, err: %w, op: %s`, tagID, taskID, err, op)
		}

		return fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("cannot commit transaction. op: %s, error: %w", op, err)
	}

	return nil
}

func (s *Storage) RemoveTagFromTask(ctx context.Context, tagID, taskID int64, userID uuid.UUID) error {
	const op = "storage.RemoveTagFromTask"

	// Transaction start
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("cannot start transaction. op: %s, error: %w", op, err)
	}
	defer tx.Rollback(ctx)

	// Check for user's access
	query, args, err := sq.
		Select("1").
		From("tasks t, tags tag").
		Where(sq.Eq{"t.id": taskID, "t.user_id": userID, "tag.id": tagID, "tag.user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	var exists int
	err = tx.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: task not found or access denied", op)
		}
		return fmt.Errorf("%s: cannot check task ownership: %w", op, err)
	}

	// Removing tag
	query, args, err = sq.
		Delete("tasks_tags").
		Where(sq.And{
			sq.Eq{"task_id": taskID},
			sq.Eq{"tag_id": tagID},
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return fmt.Errorf("tag does not exist, err: %d, op: %w", tagID, err)
		}

		return fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("cannot commit transaction. op: %s, error: %w", op, err)
	}

	return nil
}
