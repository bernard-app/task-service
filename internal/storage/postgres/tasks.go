package postgres

import (
	"bernard/internal/domain/entity"
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateTask(ctx context.Context, task entity.Task) (*entity.Task, error) {
	const op = "storage.CreateTask"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Insert("tasks").
		Columns("name", "description", "priority", "status", "start_time", "deadline", "group_id", "project_id", "user_id", "created_at", "updated_at").
		Values(&task.Name, &task.Description, &task.Priority, &task.Status, &task.StartTime, &task.Deadline, &task.GroupID, &task.ProjectID, &task.UserID, time.Now(), time.Now()).
		Suffix("RETURNING id, name, description, priority, status, start_time, deadline, group_id, project_id, user_id, created_at, updated_at, is_archived").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var createdTask entity.Task

	err = tx.QueryRow(ctx, query, args...).
		Scan(
			&createdTask.ID, &createdTask.Name, &createdTask.Description, &createdTask.Priority,
			&createdTask.Status, &createdTask.StartTime, &createdTask.Deadline, &createdTask.GroupID, &createdTask.ProjectID,
			&createdTask.UserID, &createdTask.CreatedAt, &createdTask.UpdatedAt, &createdTask.IsArchived,
		)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("%s: group or project does not exists", op)
		}

		return nil, fmt.Errorf("%s: cannot scan row: %s", op, err.Error())
	}

	return &createdTask, nil
}

func updateBuilder(task entity.UpdateTaskRequest, taskID int64, userID uuid.UUID) (sq.UpdateBuilder, error) {
	builder := sq.
		Update("tasks").
		Where(sq.Eq{"id": taskID}).
		Where(sq.Eq{"user_id": userID})

	hasUpdate := false

	if task.Name != nil {
		builder = builder.Set("name", *task.Name)
		hasUpdate = true
	}
	if task.Description != nil {
		builder = builder.Set("description", *task.Description)
		hasUpdate = true
	}
	if task.Priority != nil {
		builder = builder.Set("priority", *task.Priority)
		hasUpdate = true
	}
	if task.Status != nil {
		builder = builder.Set("status", *task.Status)
		hasUpdate = true
	}
	if task.GroupID != nil {
		builder = builder.Set("group_id", *task.GroupID)
		hasUpdate = true
	}
	if task.ProjectID != nil {
		builder = builder.Set("project_id", *task.ProjectID)
		hasUpdate = true
	}
	if task.StartTime != nil {
		builder = builder.Set("start_time", task.StartTime)
		hasUpdate = true
	}
	if task.Deadline != nil {
		builder = builder.Set("deadline", task.Deadline)
		hasUpdate = true
	}

	if !hasUpdate {
		return builder, fmt.Errorf("nothing to update")
	}

	builder = builder.Set("updated_at", time.Now())

	return builder, nil
}

func (s *Storage) UpdateTask(ctx context.Context, task entity.UpdateTaskRequest, userID uuid.UUID, taskID int64) (*entity.Task, error) {
	const op = "storage.UpdateTask"

	tx := s.getEngine(ctx)

	builder, err := updateBuilder(task, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, op)
	}

	query, args, err := builder.
		Suffix("RETURNING id, name, description, priority, status, start_time, deadline, group_id, project_id, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err)
	}

	var UpdatedTask entity.Task

	err = tx.QueryRow(ctx, query, args...).
		Scan(&UpdatedTask.ID, &UpdatedTask.Name, &UpdatedTask.Description, &UpdatedTask.Priority, &UpdatedTask.Status, &UpdatedTask.StartTime, &UpdatedTask.Deadline, &UpdatedTask.GroupID, &UpdatedTask.ProjectID, &UpdatedTask.UserID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			if pgErr.ConstraintName == "tasks_group_id_fkey" {
				return nil, fmt.Errorf("%s: group does not exist", op)
			}
		}

		return nil, fmt.Errorf("%s: cannot scan row: %s", op, err.Error())
	}

	return &UpdatedTask, nil
}

func (s *Storage) DeleteTask(ctx context.Context, userID uuid.UUID, id int64) error {
	const op = "storage.DeleteTask"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Delete("tasks").
		Where(sq.Eq{"id": id, "user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: cannot update task: %s", op, err.Error())
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s: task not found or access denied", op)
	}

	return nil
}

func (s *Storage) GetTask(ctx context.Context, id int64, userID uuid.UUID) (*entity.Task, error) {
	const op = "storage.GetTask"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Select(
			"id", "name", "description", "priority", "status", "start_time", "deadline",
			"group_id", "project_id", "user_id", "created_at", "updated_at", "is_archived",
		).
		From("tasks").
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"is_archived": false}).
		Where(sq.Eq{"user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var task entity.Task

	err = tx.QueryRow(ctx, query, args...).Scan(
		&task.ID, &task.Name, &task.Description, &task.Priority, &task.Status, &task.StartTime, &task.Deadline,
		&task.GroupID, &task.ProjectID, &task.UserID, &task.CreatedAt, &task.UpdatedAt, &task.IsArchived,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot query rows: %s", op, err.Error())
	}

	return &task, nil
}

func (s *Storage) baseListTasksQuery() sq.SelectBuilder {
	return sq.Select(
		"t.id", "t.name", "t.description", "t.priority", "t.status", "t.start_time", "t.deadline", "t.group_id", "t.project_id", "t.user_id", "t.created_at", "t.updated_at", "t.is_archived",
	).
		From("tasks t").
		LeftJoin("tasks_tags tt ON t.id = tt.task_id").
		LeftJoin("tags tag ON tt.tag_id = tag.id").
		Where(sq.Eq{"t.is_archived": false})
}

func (s *Storage) ListTasks(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Task, error) {
	const op = "storage.ListTasks"

	query, args, err := s.
		baseListTasksQuery().
		Where(sq.Eq{"t.user_id": userID}).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	return s.getTasks(ctx, query, args)
}

func (s *Storage) GetTasksByPriority(ctx context.Context, userID uuid.UUID, priority int, limit, offset uint64) ([]*entity.Task, error) {
	const op = "storage.GetTasksByPriority"

	query, args, err := s.
		baseListTasksQuery().
		Where(sq.And{
			sq.Eq{"t.user_id": userID},
			sq.Eq{"t.priority": priority},
		}).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	return s.getTasks(ctx, query, args)
}

func (s *Storage) GetTasksByDate(ctx context.Context, userID uuid.UUID, from, to time.Time, limit, offset uint64) ([]*entity.Task, error) {
	const op = "storage.GetTasksByDate"

	query, args, err := s.
		baseListTasksQuery().
		Where(sq.And{
			sq.Eq{"t.user_id": userID},
			sq.Eq{"t.start_time": from},
			sq.Eq{"t.start_time": to}}).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	return s.getTasks(ctx, query, args)
}

func (s *Storage) GetTasksByTag(ctx context.Context, userID uuid.UUID, tagID int64, limit, offset uint64) ([]*entity.Task, error) {
	const op = "storage.GetTasksByTag"

	query, args, err := s.
		baseListTasksQuery().
		Where(sq.Eq{"t.user_id": userID}).
		Where(sq.Eq{"tt.tag_id": tagID}).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	return s.getTasks(ctx, query, args)
}

func (s *Storage) getTasks(ctx context.Context, query string, args []any) ([]*entity.Task, error) {
	const op = "storage.GetTasks"

	tx := s.getEngine(ctx)

	var tasks []*entity.Task

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot query rows: %w", op, err)
	}

	defer rows.Close()

	for rows.Next() {
		var task entity.Task

		err = rows.Scan(
			&task.ID, &task.Name, &task.Description, &task.Priority, &task.Status, &task.StartTime, &task.Deadline,
			&task.GroupID, &task.ProjectID, &task.UserID, &task.CreatedAt, &task.UpdatedAt, &task.IsArchived,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: cannot scar row: %w", op, err)
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (s *Storage) GetGroupTasks(ctx context.Context, groupID int64, userID uuid.UUID, limit, offset uint64) ([]*entity.Task, error) {
	const op = "storage.GetGroupTasks"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Select(
			"id", "name", "description", "priority", "status", "start_time", "deadline",
			"group_id", "project_id", "user_id", "created_at", "updated_at", "is_archived",
		).
		From("tasks").
		Where(sq.Eq{"group_id": groupID, "user_id": userID}).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %w", op, err)
	}

	var tasks []*entity.Task

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot query rows: %w", op, err)
	}

	defer rows.Close()

	for rows.Next() {
		var task entity.Task

		err = rows.Scan(
			&task.ID, &task.Name, &task.Description, &task.Priority, &task.Status, &task.StartTime, &task.Deadline,
			&task.GroupID, &task.ProjectID, &task.UserID, &task.CreatedAt, &task.UpdatedAt, &task.IsArchived,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: cannot scan row: %w", op, err)
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (s *Storage) ArchiveTask(ctx context.Context, userID uuid.UUID, taskID int64) (*entity.Task, error) {
	const op = "storage.ArchiveTask"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Update("tasks").
		Set("is_archived", true).
		Where(sq.Eq{"id": taskID, "user_id": userID}).
		Suffix("RETURNING id, name, description, priority, status, start_time, deadline, group_id, user_id, created_at, updated_at, is_archived").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var result entity.Task

	err = tx.QueryRow(ctx, query, args...).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.Priority,
		&result.Status,
		&result.StartTime,
		&result.Deadline,
		&result.GroupID,
		&result.UserID,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.IsArchived,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: cannot exec query: %s", op, err.Error())
	}

	return &result, nil
}

func (s *Storage) ArchiveOldTasks(ctx context.Context) (int64, error) {
	const op = "storage.ArchiveOldTasks"

	tx := s.getEngine(ctx)

	query, args, err := sq.
		Update("tasks").
		Set("is_archived", true).
		Where(sq.Eq{"status": "done"}).
		Where(sq.Eq{"is_archived": false}).
		Where(sq.LtOrEq{"updated_at": time.Now().AddDate(0, 0, -7)}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("%s: cannot get result: %s", op, err.Error())
	}

	rowsAffected := result.RowsAffected()

	return rowsAffected, nil
}
