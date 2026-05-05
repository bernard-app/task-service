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

func (s *Storage) CreateProject(ctx context.Context, project entity.Project) (*entity.Project, error) {
	const op = "storage.CreateProject"

	query, args, err := sq.
		Insert("projects").
		Columns("name", "description", "user_id").
		Values(&project.Name, &project.Description, &project.UserID).
		Suffix("RETURNING id, name, description, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var createdProject entity.Project

	err = s.DB.QueryRow(ctx, query, args...).Scan(&createdProject.ID, &createdProject.Name, &createdProject.Description, &createdProject.UserID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%s: project already exists", op)
		}

		return nil, fmt.Errorf("%s: cannot create project: %s", op, err.Error())
	}

	return &createdProject, nil
}

func (s *Storage) UpdateProject(ctx context.Context, project entity.UpdateProjectRequest, projectID int64, userID uuid.UUID) (*entity.Project, error) {
	const op = "storage.UpdateProject"

	builder := sq.
		Update("projects").
		Where(sq.Eq{"id": projectID, "user_id": userID})

	hasUpdate := false

	if project.Name != nil {
		builder = builder.Set("name", *project.Name)
		hasUpdate = true
	}

	if project.Description != nil {
		builder = builder.Set("description", *project.Description)
		hasUpdate = true
	}

	if !hasUpdate {
		return nil, fmt.Errorf("%s: nothing to update", op)
	}

	query, args, err := builder.
		Suffix("RETURNING id, name, description, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var updatedProject entity.Project

	err = s.DB.QueryRow(ctx, query, args...).Scan(&updatedProject.ID, &updatedProject.Name, updatedProject.Description, &updatedProject.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot update project: %s", op, err.Error())
	}

	return &updatedProject, nil
}

func (s *Storage) DeleteProject(ctx context.Context, projectID int64, userID uuid.UUID) error {
	const op = "storage.DeleteProject"

	query, args, err := sq.
		Delete("groups").
		Where(sq.Eq{"id": projectID, "user_id": userID}).
		Suffix("RETURNING id, name, description, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var deletedProject entity.Project

	err = s.DB.QueryRow(ctx, query, args...).Scan(&deletedProject.ID, &deletedProject.Name, deletedProject.Description, &deletedProject.UserID)
	if err != nil {
		return fmt.Errorf("%s: cannot delete project: %s", op, err.Error())
	}

	return nil
}

func (s *Storage) GetProjectTree(ctx context.Context, projectID int64, userID uuid.UUID) (*entity.Project, error) {
	const op = "storage.GetProject"

	query, args, err := sq.
		Select(
			"p.id", "p.name", "p.description", "p.user_id",
			"g.id", "g.name", "g.project_id",
			"t.id", "t.name", "t.description", "t.priority", "t.status", "t.start_time", "t.deadline", "t.group_id", "t.created_at", "t.updated_at",
			"tag.id", "tag.name", "tag.color",
		).
		From("projects p").
		LeftJoin("groups g ON p.id = g.project_id").
		LeftJoin("tasks t ON g.id = t.group_id AND t.is_archived = false AND t.user_id = p.user_id").
		LeftJoin("tasks_tags tt ON t.id = tt.task_id").
		LeftJoin("tags tag ON tt.tag_id = tags.id").
		Where(sq.And{
			sq.Eq{"p.id": projectID},
			sq.Eq{"p.user_id": userID},
		}).
		OrderBy("g.id ASC", "t.created_at DESC").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var project entity.Project
	groupsMap := make(map[int64]*entity.Group)
	var groupIDsOrder []int64
	projectTaskCount := 0
	tasksMap := make(map[int64]*entity.Task)
	hasProject := false

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot query project: %s", op, err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		hasProject = true

		var gID, gProjectID *int64
		var gName *string

		var tID, tGroupID *int64
		var tName, tDesc, tStatus *string
		var tPriority *int
		var tStartTime, tDeadline, tCreatedAt, tUpdatedAt *time.Time

		var tagID *int64
		var tagName, tagColor *string

		err = rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.UserID,
			&gID, &gName, &gProjectID,
			&tID, &tName, &tDesc, &tPriority, &tStatus, &tStartTime, &tDeadline, &tGroupID, &tCreatedAt, &tUpdatedAt,
			&tagID, &tagName, &tagColor,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: cannot scan project: %s", op, err.Error())
		}

		if gID != nil {
			group, exists := groupsMap[*gID]

			if !exists {
				group = &entity.Group{
					ID:        *gID,
					Name:      *gName,
					ProjectID: *gProjectID,
					TaskCount: 0,
					Tasks:     make([]*entity.Task, 0),
				}
				groupsMap[*gID] = group
				groupIDsOrder = append(groupIDsOrder, *gID)
			}

			if tID != nil {
				task, ok := tasksMap[*tID]

				if !ok {
					task = &entity.Task{
						ID:          *tID,
						Name:        *tName,
						Description: tDesc,
						Priority:    tPriority,
						Status:      tStatus,
						StartTime:   tStartTime,
						Deadline:    tDeadline,
						CreatedAt:   *tCreatedAt,
						UpdatedAt:   *tUpdatedAt,
						GroupID:     *tGroupID,
					}

					tasksMap[*tID] = task

					group.Tasks = append(group.Tasks, task)
					group.TaskCount++
					projectTaskCount++
				}

				if tagID != nil {
					tag := &entity.Tag{
						ID:    *tagID,
						Name:  *tagName,
						Color: *tagColor,
					}

					task.Tags = append(task.Tags, tag)
				}
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: cannot iterate projects: %s", op, err.Error())
	}

	if !hasProject {
		return nil, fmt.Errorf("%s: project not found", op)
	}

	for _, id := range groupIDsOrder {
		project.Groups = append(project.Groups, groupsMap[id])
	}

	project.TaskCount = projectTaskCount

	return &project, nil
}

func (s *Storage) GetProjects(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Project, error) {
	const op = "storage.GetProjects"

	query, args, err := sq.
		Select("id", "name", "description", "user_id").
		From("projects").
		Where(sq.Eq{"user_id": userID}).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: cannot build query: %s", op, err.Error())
	}

	var projects []*entity.Project

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot read projects: %s", op, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var project entity.Project

		if err = rows.Scan(&project.ID, &project.Name, &project.Description, &project.UserID); err != nil {
			return nil, fmt.Errorf("%s: cannot read project: %s", op, err.Error())
		}

		projects = append(projects, &project)
	}

	return projects, nil
}
