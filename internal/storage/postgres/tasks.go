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
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Storage) CreateTask(ctx context.Context, task entity.Task, tasksIDs []int64) (*entity.Task, error) {
	const op = "storage.CreateTask"

	query, args, err := sq.
		Insert("tasks").
		Columns("name", "description", "priority", "status", "start_time", "deadline", "group_id", "user_id", "created_at", "updated_at").
		Values(&task.Name, &task.Description, &task.Priority, &task.Status, &task.StartTime, &task.Deadline, &task.GroupID, &task.UserID, time.Now(), time.Now()).
		Suffix("RETURNING id, name, description, priority, status, start_time, deadline, group_id, user_id, created_at, updated_at, is_archived").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot start transaction: %s", op, err.Error())
	}
	defer tx.Rollback(ctx)

	var createdTask entity.Task

	err = tx.
		QueryRow(ctx, query, args...).
		Scan(
			&createdTask.ID, &createdTask.Name, &createdTask.Description, &createdTask.Priority,
			&createdTask.Status, &createdTask.StartTime, &createdTask.Deadline, &createdTask.GroupID,
			&createdTask.UserID, &createdTask.CreatedAt, &createdTask.UpdatedAt, &createdTask.IsArchived,
		)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("%s: group does not exists", op)
		}

		return nil, fmt.Errorf("%s: cannot scan row: %s", op, err.Error())
	}

	if len(tasksIDs) > 0 {
		tagsBuilder := sq.Insert("tasks_tags").Columns("task_id", "tag_id")
		for _, id := range task.Tags {
			tagsBuilder = tagsBuilder.Values(createdTask.ID, id)
		}

		query, args, err = tagsBuilder.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
		}

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, fmt.Errorf("%s: tag does not exists", op)
			}

			return nil, fmt.Errorf("%s: cannot scan row: %w", op, err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot commit transaction: %s", op, err.Error())
	}

	return &createdTask, nil
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
		builder = builder.Set("start_time", task.StartTime)
		hasUpdate = true
	}
	if task.Deadline != nil {
		builder = builder.Set("deadline", task.Deadline)
		hasUpdate = true
	}

	if !hasUpdate {
		return nil, fmt.Errorf("cannot update task. op: %s", op)
	}

	builder = builder.Set("updated_at", time.Now())

	query, args, err := builder.
		Suffix("RETURNING id, name, description, priority, status, start_time, deadline, group_id, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err)
	}

	var UpdatedTask entity.Task

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot start transaction: %s", op, err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, query, args...).
		Scan(&UpdatedTask.ID, &UpdatedTask.Name, &UpdatedTask.Description, &UpdatedTask.Priority, &UpdatedTask.Status, &UpdatedTask.StartTime, &UpdatedTask.Deadline, &UpdatedTask.GroupID, &UpdatedTask.UserID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			if pgErr.ConstraintName == "tasks_group_id_fkey" {
				return nil, fmt.Errorf("%s: group does not exist", op)
			}
		}

		return nil, fmt.Errorf("%s: cannot scan row: %s", op, err.Error())
	}

	if task.TagsIDs != nil {
		// Deleting all task tags
		query, args, err = sq.Delete("tasks_tags").Where(sq.Eq{"task_id": UpdatedTask.ID}).PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
		}

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("%s: cannot update task: %s", op, err.Error())
		}

		// Add new task's tags
		if len(*task.TagsIDs) > 0 {
			tagBuilder := sq.Insert("tasks_tags").Columns("task_id", "tag_id")

			for _, tag := range *task.TagsIDs {
				tagBuilder = tagBuilder.Values(UpdatedTask.ID, tag)
			}

			query, args, err = tagBuilder.PlaceholderFormat(sq.Dollar).ToSql()
			if err != nil {
				return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
			}

			_, err = tx.Exec(ctx, query, args...)
			if err != nil {
				var pgErr *pgconn.PgError

				if errors.As(err, &pgErr) && pgErr.Code == "23503" {
					return nil, fmt.Errorf("%s: tag does not exist", op)
				}

				return nil, fmt.Errorf("%s: cannot update task: %s", op, err.Error())
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot commit transaction: %s", op, err.Error())
	}

	return &UpdatedTask, nil
}

func (s *Storage) DeleteTask(ctx context.Context, userID uuid.UUID, id int64) error {
	const op = "storage.DeleteTask"

	query, args, err := sq.
		Delete("tasks").
		Where(sq.Eq{"id": id, "user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	result, err := s.DB.Exec(ctx, query, args...)
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

	query, args, err := sq.
		Select(
			"t.id", "t.name", "t.description", "t.priority", "t.status", "t.start_time", "t.deadline", "t.group_id", "t.user_id", "t.created_at", "t.updated_at", "t.is_archived",
			"tag.id", "tag.name", "tag.color", "tag.user_id",
		).
		From("tasks t").
		LeftJoin("tasks_tags tt ON t.id = tt.task_id").
		LeftJoin("tags tag ON tt.tag_id = tag.id").
		Where(sq.Eq{"t.id": id}).
		Where(sq.Eq{"t.is_archived": false}).
		Where(sq.Eq{"t.user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var task entity.Task
	hasTask := false

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot query rows: %s", op, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		hasTask = true

		var tDesc, tStatus *string
		var tPriority *int
		var tDeadLine, tStartTime *time.Time

		var tagID *int64
		var tagName, tagColor *string
		var tagUserID *uuid.UUID

		err = rows.Scan(
			&task.ID, &task.Name,
			&tDesc, &tPriority, &tStatus, &tStartTime, &tDeadLine, // <- Possible NULL
			&task.GroupID, &task.UserID, &task.CreatedAt, &task.UpdatedAt, &task.IsArchived,
			&tagID, &tagName, &tagColor, &tagUserID, // <- Possible NULL
		)
		if err != nil {
			return nil, fmt.Errorf("%s: cannot scan row: %s", op, err.Error())
		}

		if tDesc != nil {
			task.Description = tDesc
		}

		if tStatus != nil {
			task.Status = tStatus
		}

		if tDeadLine != nil {
			task.Deadline = tDeadLine
		}

		if tStartTime != nil {
			task.StartTime = tStartTime
		}

		if tPriority != nil {
			task.Priority = tPriority
		}

		if tagID != nil {
			tag := &entity.Tag{
				ID:     *tagID,
				Name:   *tagName,
				Color:  *tagColor,
				UserID: *tagUserID,
			}

			task.Tags = append(task.Tags, tag)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: cannot iterate rows: %s", op, err.Error())
	}

	if !hasTask {
		return nil, fmt.Errorf("%s: task not found", op)
	}

	return &task, nil
}

func (s *Storage) baseListTasksQuery() sq.SelectBuilder {
	return sq.Select(
		"t.id", "t.name", "t.description", "t.priority", "t.status", "t.start_time", "t.deadline", "t.group_id", "t.user_id", "t.created_at", "t.updated_at", "t.is_archived",
		"g.name",
		"p.id", "p.name",
		"tag.id", "tag.name", "tag.color", "tag.user_id",
	).
		From("tasks t").
		Join("groups g ON t.group_id = g.id").
		Join("projects p ON g.project_id = p.id").
		LeftJoin("tasks_tags tt ON t.id = tt.task_id").
		LeftJoin("tags tag ON tt.tag_id = tag.id").
		Where(sq.Eq{"t.is_archived": false})
}

func (s *Storage) ListTasks(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.UserTasksTab, error) {
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

	return getTasks(s.DB, ctx, query, args)
}

func (s *Storage) GetTasksByPriority(ctx context.Context, userID uuid.UUID, priority int, limit, offset uint64) ([]*entity.UserTasksTab, error) {
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

	return getTasks(s.DB, ctx, query, args)
}

func (s *Storage) GetTasksByDate(ctx context.Context, userID uuid.UUID, from, to time.Time, limit, offset uint64) ([]*entity.UserTasksTab, error) {
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

	return getTasks(s.DB, ctx, query, args)
}

func (s *Storage) GetTasksByTag(ctx context.Context, userID uuid.UUID, tagID int64, limit, offset uint64) ([]*entity.UserTasksTab, error) {
	const op = "storage.GetTasksByTag"

	query, args, err := s.
		baseListTasksQuery().
		Where(sq.Eq{"t.user_id": userID}).
		Where("t.id IN (SELECT tt2.task_id FROM tasks_tags tt2 JOIN tags tag2 ON tt2.tag_id = tag2.id WHERE tag2.id LIKE ?)", tagID).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	return getTasks(s.DB, ctx, query, args)
}

func getTasks(DB *pgxpool.Pool, ctx context.Context, query string, args []interface{}) ([]*entity.UserTasksTab, error) {
	const op = "storage.GetTasks"

	rows, err := DB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot query rows: %w", op, err)
	}
	defer rows.Close()

	tasksMap := make(map[int64]*entity.UserTasksTab)
	var tasksOrder []int64

	for rows.Next() {
		var task entity.UserTasksTab

		var tDesc, tStatus *string
		var tPriority *int
		var tDeadLine, tStartTime *time.Time

		var tagID *int64
		var tagName, tagColor *string
		var tagUserID *uuid.UUID

		err = rows.Scan(
			&task.Task.ID, &task.Task.Name, &tDesc, &tPriority, &tStatus, &tStartTime, &tDeadLine, &task.Task.GroupID, &task.Task.UserID, &task.Task.CreatedAt, &task.Task.UpdatedAt, &task.Task.IsArchived,
			&task.GroupName,
			&task.ProjectID, &task.ProjectName,
			&tagID, &tagName, &tagColor, &tagUserID,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: cannot scan row: %w", op, err)
		}

		existingTask, ok := tasksMap[*tagID]
		if !ok {
			if tDesc != nil {
				task.Task.Description = tDesc
			}
			if tStatus != nil {
				task.Task.Status = tStatus
			}
			if tDeadLine != nil {
				task.Task.Deadline = tDeadLine
			}
			if tStartTime != nil {
				task.Task.StartTime = tStartTime
			}
			if tPriority != nil {
				task.Task.Priority = tPriority
			}

			task.Task.Tags = make([]*entity.Tag, 0)

			tasksMap[task.Task.ID] = &task
			existingTask = &task
			tasksOrder = append(tasksOrder, task.Task.ID)
		}

		if tagID != nil {
			tag := &entity.Tag{
				ID:     *tagID,
				Name:   *tagName,
				Color:  *tagColor,
				UserID: *tagUserID,
			}
			existingTask.Task.Tags = append(existingTask.Task.Tags, tag)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: cannot iterate rows: %w", op, err)
	}

	result := make([]*entity.UserTasksTab, 0, len(tasksOrder))
	for _, taskID := range tasksOrder {
		result = append(result, tasksMap[taskID])
	}

	return result, nil
}

func (s *Storage) ArchiveTask(ctx context.Context, userID uuid.UUID, taskID int64) (*entity.Task, error) {
	const op = "storage.ArchiveTask"

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

	err = s.DB.QueryRow(ctx, query, args...).Scan(
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

	result, err := s.DB.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("%s: cannot get result: %s", op, err.Error())
	}

	rowsAffected := result.RowsAffected()

	return rowsAffected, nil
}
