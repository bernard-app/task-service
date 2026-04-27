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
		Prefix("RETURNING id, name, description, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var createdProject entity.Project

	err = s.DB.QueryRow(ctx, query, args...).Scan(&createdProject.ID, &createdProject.Name, createdProject.Description, &createdProject.UserID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New(fmt.Sprintf("%s: project already exists", op))
		}

		return nil, errors.New(fmt.Sprintf("%s: cannot create project: %s", op, err.Error()))
	}

	return &createdProject, nil
}

func (s *Storage) UpdateProject(ctx context.Context, project entity.Project) (*entity.Project, error) {
	const op = "storage.UpdateProject"

	query, args, err := sq.
		Update("projects").
		SetMap(map[string]interface{}{"name": &project.Name, "description": &project.Description}).
		Where(sq.Eq{"id": &project.ID}).
		Prefix("RETURNING id, name, description, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var updatedProject entity.Project

	err = s.DB.QueryRow(ctx, query, args...).Scan(&updatedProject.ID, &updatedProject.Name, updatedProject.Description, &updatedProject.UserID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot update project: %s", op, err.Error()))
	}

	return &updatedProject, nil
}

func (s *Storage) DeleteProject(ctx context.Context, projectID int64) (*entity.Project, error) {
	const op = "storage.DeleteProject"

	query, args, err := sq.
		Delete("groups").
		Where(sq.Eq{"id": projectID}).
		Prefix("RETURNING id, name, description, user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var deletedProject entity.Project

	err = s.DB.QueryRow(ctx, query, args...).Scan(&deletedProject.ID, &deletedProject.Name, deletedProject.Description, &deletedProject.UserID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot delete project: %s", op, err.Error()))
	}

	return &deletedProject, nil
}

func (s *Storage) GetProjectTree(ctx context.Context, projectID int64, userID uuid.UUID) (*entity.Project, error) {
	const op = "storage.GetProject"

	query, args, err := sq.
		Select(
			"p.id", "p.name", "p.description", "p.user_id",
			"g.id", "g.name", "g.project_id",
			"t.id", "t.name", "t.description", "t.tags", "t.priority", "t.status", "t.start_time", "t.deadline", "t.group_id", "t.created_at", "t.updated_at",
		).
		From("projects p").
		LeftJoin("groups g ON p.id = g.project_id").
		LeftJoin("tasks t ON g.id = t.group_id").
		Where(sq.Eq{"p.id": projectID, "p.user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var project entity.Project
	groupsMap := make(map[int64]*entity.Group)
	var groupIDsOrder []int64
	projectTaskCount := 0

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot query project: %s", op, err.Error()))
	}
	defer rows.Close()

	for rows.Next() {
		var gID, gProjectID *int64
		var gName *string

		var tID, tGroupID *int64
		var tName, tDesc, tTags, tStatus *string
		var tPriority *int
		var tStartTime, tDeadline, tCreatedAt, tUpdatedAt *time.Time

		err = rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.UserID,
			&gID, &gName, &gProjectID,
			&tID, &tName, &tDesc, &tTags, &tPriority, &tStatus, &tGroupID, &tStartTime, &tDeadline, &tCreatedAt, &tUpdatedAt,
		)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s: cannot scan project: %s", op, err.Error()))
		}

		if gID != nil {
			group, exists := groupsMap[*gID]

			if !exists {
				group = &entity.Group{
					ID:        *gID,
					Name:      *gName,
					ProjectID: *gProjectID,
					TaskCount: 0,
					Tasks:     make([]entity.Task, 0),
				}
				groupsMap[*gID] = group
				groupIDsOrder = append(groupIDsOrder, *gID)
			}

			if tID != nil {
				task := entity.Task{
					ID:          *tID,
					Name:        *tName,
					Description: *tDesc,
					Tags:        *tTags,
					Priority:    *tPriority,
					Status:      *tStatus,
					StartTime:   *tStartTime,
					Deadline:    *tDeadline,
					CreatedAt:   *tCreatedAt,
					UpdatedAt:   *tUpdatedAt,
					GroupID:     *tGroupID,
				}

				group.Tasks = append(group.Tasks, task)
				group.TaskCount++
				projectTaskCount++
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot iterate projects: %s", op, err.Error()))
	}

	for _, id := range groupIDsOrder {
		project.Groups = append(project.Groups, *groupsMap[id])
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
		return nil, errors.New(fmt.Sprintf("%s: cannot build query: %s", op, err.Error()))
	}

	var projects []*entity.Project

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: cannot read projects: %s", op, err.Error()))
	}

	defer rows.Close()

	for rows.Next() {
		var project entity.Project

		if err = rows.Scan(&project.ID, &project.Name, &project.Description, &project.UserID); err != nil {
			return nil, errors.New(fmt.Sprintf("%s: cannot read project: %s", op, err.Error()))
		}

		projects = append(projects, &project)
	}

	return projects, nil
}
