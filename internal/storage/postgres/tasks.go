package postgres

import (
	"bernard/internal/domain/entity"
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Storage) CreateTask(ctx context.Context, task entity.Task) (*entity.Task, error) {
	const op = "storage.CreateTask"

	query, args, err := sq.
		Insert("tasks").
		Columns("name", "description", "tags", "priority", "status", "start_time", "deadline", "group_id", "user_id", "created_at", "updated_at").
		Values(&task.Name, &task.Description, &task.Tags, &task.Priority, &task.Status, &task.StartTime, &task.Deadline, &task.GroupID, &task.UserID, time.Now(), time.Now()).
		Suffix("RETURNING id, name, description, tags, priority, status, start_time, deadline, group_id, user_id, created_at, updated_at, is_archived").
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

func (s *Storage) UpdateTask(ctx context.Context, task entity.UpdateTaskRequest, userID uuid.UUID, taskID int64) (*entity.Task, error) {
	const op = "storage.UpdateTask"

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
	if task.Tags != nil {
		builder = builder.Set("tags", *task.Tags)
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
	if task.StartTime != nil {
		builder = builder.Set("start_time", task.StartTime.Time())
		hasUpdate = true
	}
	if task.Deadline != nil {
		builder = builder.Set("deadline", task.Deadline.Time())
		hasUpdate = true
	}

	if !hasUpdate {
		return nil, errors.New(fmt.Sprintf("%s: cannot update task: %s", op, "no update available"))
	}

	builder = builder.Set("updated_at", time.Now())

	query, args, err := builder.
		Suffix("RETURNING id, name, description, tags, priority, status, start_time, deadline, group_id, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var UpdatedTask entity.Task

	err = s.DB.QueryRow(ctx, query, args...).
		Scan(&UpdatedTask.ID, &UpdatedTask.Name, &UpdatedTask.Description, &UpdatedTask.Tags, &UpdatedTask.Priority, &UpdatedTask.Status, &UpdatedTask.StartTime, &UpdatedTask.Deadline, &UpdatedTask.GroupID, &UpdatedTask.UserID)
	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New(fmt.Sprintf("%s: cannot update task: %s", op, err.Error()))
	}

	return &UpdatedTask, nil
}

func (s *Storage) DeleteTask(ctx context.Context, userID uuid.UUID, id int64) (*entity.Task, error) {
	const op = "storage.DeleteTask"

	query, args, err := sq.
		Delete("tasks").
		Where(sq.Eq{"id": id, "user_id": userID}).
		Suffix("RETURNING id, name, description, tags, priority, status, start_time, deadline, group_id, user_id, created_at, updated_at, is_archived").
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
		Select("id", "name", "tags", "description", "priority", "status", "start_time", "deadline", "group_id", "user_id", "created_at", "updated_at, is_archived").
		From("tasks").
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"is_archived": false}).
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
		From("tasks t").
		Join("groups g ON t.group_id = g.id").
		Join("projects p ON g.project_id = p.id").
		Where(sq.Eq{"t.is_archived": false}).
		Where(sq.Eq{"t.user_id": userID}).
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
		Where(sq.Eq{"is_archived": false}).
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
		Where(sq.Eq{"is_archived": false}).
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
			sq.Eq{"is_archived": false},
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

func (s *Storage) ArchiveTask(ctx context.Context, userID uuid.UUID, taskID int64) (*entity.Task, error) {
	const op = "storage.ArchiveTask"

	query, args, err := sq.
		Update("tasks").
		Set("is_archived", true).
		Where(sq.Eq{"id": taskID, "user_id": userID}).
		Suffix("RETURNING id, name, description, tags, priority, status, start_time, deadline, group_id, user_id, created_at, updated_at, is_archived").
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

func (s *Storage) ArchiveOldTasks(ctx context.Context) (int64, error) {
	const op = "storage.ArchiveOldTasks"

	query, args, err := sq.
		Update("tasks").
		Set("is_archived", true).
		Where(sq.Eq{"status": "done"}).
		Where(sq.Eq{"is_archived": false}).
		Where(sq.LtOrEq{"updated_at": time.Now().AddDate(0, 0, -7)}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return 0, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	result, err := s.DB.Exec(ctx, query, args...)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("%s: cannot get result: %s", op, err.Error()))
	}

	rowsAffected := result.RowsAffected()

	return rowsAffected, nil
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
		&result.IsArchived,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot exec query: %s", op, err.Error()))
	}

	return &result, nil
}
