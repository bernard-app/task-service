package postgres

import (
	"bernard/internal/domain/entity"
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateTask(ctx context.Context, task entity.Task) (*entity.Task, error) {
	const op = "storage.CreateTask"

	query, args, err := sq.
		Insert("tasks").
		Columns("id", "name", "description", "tags", "priority", "status", "group_id", "user_id", "created_at", "updated_at").
		Values(&task.ID, &task.Name, &task.Description, &task.Tags, &task.Priority, &task.Status, &task.GroupID, &task.UserID, &task.CreatedAt, &task.UpdatedAt).
		Prefix("RETURNING id, name, description, tags, priority, status, group_id, user_id, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var result entity.Task

	err = s.DB.QueryRow(ctx, query, args...).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.Tags,
		&result.Priority,
		&result.Status,
		&result.GroupID,
		&result.UserID,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New(fmt.Sprintf("%s: task %s already exists", op, result.Name))
		}

		return nil, errors.New(fmt.Sprintf("%s: cannot create task: %s", op, err.Error()))
	}

	return &result, nil
}

func (s *Storage) UpdateTask(ctx context.Context, task entity.Task) (*entity.Task, error) {
	const op = "storage.UpdateTask"

	query, args, err := sq.
		Update("tasks").
		SetMap(map[string]interface{}{
			"name":        &task.Name,
			"description": &task.Description,
			"tags":        &task.Tags,
			"priority":    &task.Priority,
			"status":      &task.Status,
			"group_id":    &task.GroupID,
			"updated_at":  &task.UpdatedAt,
		}).
		Where(sq.Eq{"id": task.ID}).
		Prefix("RETURNING id, name, description, tags, priority, status, group_id, user_id, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var returnedTask entity.Task

	err = s.DB.QueryRow(ctx, query, args...).Scan(
		&returnedTask.ID,
		&returnedTask.Name,
		&returnedTask.Description,
		&returnedTask.Tags,
		&returnedTask.Priority,
		&returnedTask.Status,
		&returnedTask.GroupID,
		&returnedTask.UserID,
		&returnedTask.CreatedAt,
		&returnedTask.UpdatedAt,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot update task: %s", op, err.Error()))
	}

	return &returnedTask, nil
}

func (s *Storage) DeleteTask(ctx context.Context, id uint64) (*entity.Task, error) {
	const op = "storage.DeleteTask"

	query, args, err := sq.
		Delete("tasks").
		Where(sq.Eq{"id": id}).
		Prefix("RETURNING id, name, description, tags, priority, status, group_id, user_id, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var task entity.Task

	err = s.DB.QueryRow(ctx, query, args...).Scan(
		&task.ID,
		&task.Name,
		&task.Description,
		&task.Tags,
		&task.Priority,
		&task.Status,
		&task.GroupID,
		&task.UserID,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(fmt.Sprintf("%s: task not exists: %s", op, err.Error()))
		}

		return nil, errors.New(fmt.Sprintf("%s: cannot delete task: %s", op, err.Error()))
	}

	return &task, nil
}

func (s *Storage) GetTask(ctx context.Context, id uint64) (*entity.Task, error) {
	const op = "storage.GetTask"

	query, args, err := sq.
		Select("id, name, description, priority, status, group_id, user_id, created_at, updated_at").
		From("tasks").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var task entity.Task

	err = s.DB.QueryRow(ctx, query, args...).Scan(
		&task.ID,
		&task.Name,
		&task.Description,
		&task.Priority,
		&task.Status,
		&task.GroupID,
		&task.UserID,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot get task: %s", op, err.Error()))
	}

	return &task, nil
}

func (s *Storage) ListTasks(ctx context.Context, userID uint64) ([]*entity.Task, error) {
	const op = "storage.ListTasks"

	query, args, err := sq.
		Select("id, name, description, priority, status, group_id, user_id, created_at, updated_at").
		From("tasks").
		Where(sq.Eq{"user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot list tasks: %s", op, err.Error()))
	}
	defer rows.Close()

	var tasks []*entity.Task

	for rows.Next() {
		var task entity.Task

		err = rows.Scan(&task.ID, &task.Name, &task.Description, &task.Priority, &task.Status, &task.GroupID, &task.UserID, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s: cannot scan row: %s", op, err.Error()))
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}
