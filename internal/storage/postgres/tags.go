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

func (s *Storage) CreateTag(ctx context.Context, tag entity.Tag) (*entity.Tag, error) {
	const op = "storage.CreateTag"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Insert("tags").
		Columns("name", "color", "user_id", "project_id").
		Values(tag.Name, tag.Color, tag.UserID, tag.ProjectID).
		Suffix("RETURNING id, name, color, user_id, project_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	var createdTag entity.Tag

	err = tx.QueryRow(ctx, query, args...).Scan(&createdTag.ID, &createdTag.Name, &createdTag.Color, &createdTag.UserID, &createdTag.ProjectID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%s: %w", op, response.ErrAlreadyExists)
		}

		return nil, fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	return &createdTag, nil
}

func (s *Storage) UpdateTag(ctx context.Context, tagID int64, name *string, color *string, userID uuid.UUID) (*entity.Tag, error) {
	const op = "storage.UpdateTag"

	tx := s.getEngine(ctx)

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

	err = tx.QueryRow(ctx, query, args...).Scan(&updatedTag.ID, &updatedTag.Name, &updatedTag.Color, &updatedTag.UserID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("%s: %w", op, response.ErrNotFound)
		}

		return nil, fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	return &updatedTag, nil
}

func (s *Storage) DeleteTag(ctx context.Context, tagID int64, userID uuid.UUID) error {
	const op = "storage.DeleteTag"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Delete("tags").
		Where(sq.Eq{"id": tagID, "user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return response.ErrNotFound
		}

		return fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	return nil
}

func (s *Storage) GetTag(ctx context.Context, tagID int64, userID uuid.UUID) (*entity.Tag, error) {
	const op = "storage.GetTag"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Select("id", "name", "color", "user_id", "project_id").
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

	err = tx.QueryRow(ctx, query, args...).Scan(&tag.ID, &tag.Name, &tag.Color, &tag.UserID, &tag.ProjectID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("%s: %w", op, response.ErrNotFound)
		}

		return nil, fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	return &tag, nil
}

func (s *Storage) GetTagList(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Tag, error) {
	const op = "storage.GetTagList"

	tx := s.getEngine(ctx)

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

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	defer rows.Close()

	var tags []*entity.Tag

	for rows.Next() {
		var tag entity.Tag

		err = rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.UserID)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, fmt.Errorf("%s: %w", op, response.ErrNotFound)
			}

			return nil, fmt.Errorf("cannot scan rows. op: %s, error: %w", op, err)
		}

		tags = append(tags, &tag)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("cannot iterate over rows. op: %s, error: %w", op, err)
	}

	return tags, nil
}

func (s *Storage) GetTaskTags(ctx context.Context, taskID int64) ([]*entity.Tag, error) {
	const op = "storage.GetTaskTags"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Select(
			"t.id", "t.name", "t.color", "t.user_id", "t.project_id",
		).
		From("tags t").
		Join("tasks_tags tt ON t.id = tt.tag_id").
		Where(sq.Eq{"tt.task_id": taskID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %w", op, err)
	}

	var tags []*entity.Tag

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot query rows: %w", op, err)
	}

	defer rows.Close()

	for rows.Next() {
		var tag entity.Tag

		err = rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.UserID, &tag.ProjectID)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, fmt.Errorf("%s: %w", op, response.ErrNotFound)
			}

			return nil, fmt.Errorf("%s: cannot scan row: %w", op, err)
		}

		tags = append(tags, &tag)
	}

	return tags, nil
}

func (s *Storage) AddTagsToTask(ctx context.Context, tagsIDs []int64, taskID int64) error {
	const op = "storage.AddTagToTask"

	if len(tagsIDs) == 0 {
		return nil
	}

	tx := s.getEngine(ctx)

	tagsBuilder := sq.Insert("tasks_tags").Columns("task_id", "tag_id")

	for _, id := range tagsIDs {
		tagsBuilder = tagsBuilder.Values(taskID, id)
	}

	query, args, err := tagsBuilder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%s: %w", op, response.ErrAlreadyExists)
		}

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return fmt.Errorf("%s: %w", op, response.ErrNotFound)
		}

		return fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	return nil
}

func (s *Storage) RemoveTagsFromTask(ctx context.Context, tagsIDs []int64, taskID int64) error {
	const op = "storage.RemoveTagFromTask"

	tx := s.getEngine(ctx)

	builder := sq.Delete("tasks_tags").Where(sq.Eq{"task_id": taskID, "tag_id": tagsIDs})

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		return fmt.Errorf("cannot build query. op: %s, error: %w", op, err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("cannot execute query. op: %s, error: %w", op, err)
	}

	return nil
}

func (s *Storage) RemoveAllTagsFromTask(ctx context.Context, taskID int64) error {
	const op = "storage.RemoveAllTagsFromTask"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Delete("tasks_tags").
		Where(sq.Eq{"task_id": taskID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: cannot build query: %w", op, err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: cannot exec query: %w", op, err)
	}

	return nil
}

func (s *Storage) CheckTagOwnership(ctx context.Context, userID uuid.UUID, tagID int64) (bool, error) {
	const op = "storage.CheckTagOwnership"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Select("1").
		From("tags").
		Where(sq.Eq{"id": tagID, "user_id": userID}).
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
