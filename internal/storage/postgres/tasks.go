package postgres

import (
	"bernard/internal/domain/entity"
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Storage) CreateTask(ctx context.Context, task entity.Task) (*entity.Task, error) {
	const op = "storage.CreateTask"

	query, args, err := sq.
		Insert("tasks").
		Columns("id", "name", "description", "tags", "priority", "status", "start_time", "deadline", "group_id", "user_id", "created_at", "updated_at").
		Values(&task.ID, &task.Name, &task.Description, &task.Tags, &task.Priority, &task.Status, &task.StartTime, &task.Deadline, &task.GroupID, &task.UserID, time.Now(), time.Now()).
		Prefix("RETURNING id, name, description, tags, priority, status, start_time, deadline, group_id, user_id, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	createdTask, err := execTask(s.DB, ctx, query, args)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot create task: %s", op, err.Error()))
	}

	return createdTask, nil
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
			"start_time":  &task.StartTime,
			"deadline":    &task.Deadline,
			"group_id":    &task.GroupID,
			"updated_at":  time.Now(),
		}).
		Where(sq.Eq{"id": task.ID, "user_id": task.UserID}).
		Prefix("RETURNING id, name, description, tags, priority, status, start_time, deadline, group_id, user_id, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	updatedTask, err := execTask(s.DB, ctx, query, args)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot update task: %s", op, err.Error()))
	}

	return updatedTask, nil
}

func (s *Storage) DeleteTask(ctx context.Context, userID uuid.UUID, id int64) (*entity.Task, error) {
	const op = "storage.DeleteTask"

	query, args, err := sq.
		Delete("tasks").
		Where(sq.Eq{"id": id, "user_id": userID}).
		Prefix("RETURNING id, name, description, tags, priority, status, start_time, deadline, group_id, user_id, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	deletedTask, err := execTask(s.DB, ctx, query, args)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot delete task: %s", op, err.Error()))
	}

	return deletedTask, nil
}

func (s *Storage) GetTask(ctx context.Context, id int64) (*entity.Task, error) {
	const op = "storage.GetTask"

	query, args, err := sq.
		Select("id", "name", "description", "priority", "status", "start_time", "deadline", "group_id", "user_id", "created_at", "updated_at").
		From("tasks").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	task, err := execTask(s.DB, ctx, query, args)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot get task: %s", op, err.Error()))
	}

	return task, nil
}

func (s *Storage) ListTasks(ctx context.Context, userID uuid.UUID) ([]*entity.UserTasksTab, error) {
	const op = "storage.ListTasks"

	query, args, err := sq.
		Select(
			"t.id", "t.name", "t.description", "t.priority", "t.status", "t.start_time", "t.deadline", "t.group_id", "t.user_id", "t.created_at", "t.updated_at",
			"g.name",
			"p.id", "p.name",
		).
		From("tasks").
		Join("group g ON t.group_id = g.id").
		Join("project p ON g.project_id = p.id").
		Where(sq.Eq{"user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	tasks, err := getTasks(s.DB, ctx, query, args)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot get tasks: %s", op, err.Error()))
	}

	return tasks, nil
}

func (s *Storage) GetTasksByPriority(ctx context.Context, userID uuid.UUID, priority int) ([]*entity.UserTasksTab, error) {
	const op = "storage.GetTasksByPriority"

	query, args, err := sq.
		Select(
			"t.id", "t.name", "t.description", "t.priority", "t.status", "t.start_time", "t.deadline", "t.group_id", "t.user_id", "t.created_at", "t.updated_at",
			"g.name",
			"p.id", "p.name",
		).
		From("tasks").
		Join("group g ON t.group_id = g.id").
		Join("project p ON g.project_id = p.id").
		Where(sq.Eq{"user_id": userID, "priority": priority}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	tasks, err := getTasks(s.DB, ctx, query, args)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot get tasks: %s", op, err.Error()))
	}

	return tasks, nil
}

func (s *Storage) GetTasksByDate(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*entity.UserTasksTab, error) {
	const op = "storage.GetTasksByDate"

	query, args, err := sq.
		Select(
			"t.id", "t.name", "t.description", "t.priority", "t.status", "t.start_time", "t.deadline", "t.group_id", "t.user_id", "t.created_at", "t.updated_at",
			"g.name",
			"p.id", "p.name",
		).
		From("tasks").
		Join("group g ON t.group_id = g.id").
		Join("project p ON g.project_id = p.id").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.GtOrEq{"start_time": from}).
		Where(sq.LtOrEq{"start_time": to}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	tasks, err := getTasks(s.DB, ctx, query, args)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot get tasks: %s", op, err.Error()))
	}

	return tasks, nil
}

func (s *Storage) GetTasksByTag(ctx context.Context, userID uuid.UUID, tag string) ([]*entity.UserTasksTab, error) {
	const op = "storage.GetTasksByTag"

	query, args, err := sq.
		Select(
			"t.id", "t.name", "t.description", "t.priority", "t.status", "t.start_time", "t.deadline", "t.group_id", "t.user_id", "t.created_at", "t.updated_at",
			"g.name",
			"p.id", "p.name",
		).
		From("tasks").
		Join("group g ON t.group_id = g.id").
		Join("project p ON g.project_id = p.id").
		Where(sq.And{
			sq.Eq{"user_id": userID},
			sq.Expr("tags LIKE ?", "%"+tag+"%"),
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	tasks, err := getTasks(s.DB, ctx, query, args)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot get tasks: %s", op, err.Error()))
	}

	return tasks, nil
}

func getTasks(DB *pgxpool.Pool, ctx context.Context, query string, args []interface{}) ([]*entity.UserTasksTab, error) {
	const op = "storage.GetTasks"

	rows, err := DB.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot list tasks: %s", op, err.Error()))
	}
	defer rows.Close()

	var tasks []*entity.UserTasksTab

	for rows.Next() {
		var task entity.UserTasksTab

		err = rows.Scan(
			&task.Task.ID, &task.Task.Name, &task.Task.Description, &task.Task.Priority, &task.Task.Status, &task.Task.StartTime, &task.Task.Deadline, &task.Task.GroupID, &task.Task.UserID, &task.Task.CreatedAt, &task.Task.UpdatedAt,
			&task.GroupName, &task.ProjectID, &task.ProjectName,
		)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s: cannot scan row: %s", op, err.Error()))
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func execTask(DB *pgxpool.Pool, ctx context.Context, query string, args []interface{}) (*entity.Task, error) {
	const op = "storage.ExecTask"

	var result entity.Task

	err := DB.QueryRow(ctx, query, args...).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.Tags,
		&result.Priority,
		&result.Status,
		&result.StartTime,
		&result.Deadline,
		&result.GroupID,
		&result.UserID,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot exec query: %s", op, err.Error()))
	}

	return &result, nil
}
