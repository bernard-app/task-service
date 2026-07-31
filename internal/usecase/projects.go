package usecase

import (
	"bernard/internal/domain/entity"
	"bernard/pkg/response"
	"context"

	"github.com/google/uuid"
)

func (u *UseCase) CreateProject(ctx context.Context, project entity.Project) (*entity.Project, error) {
	const op = "storage.CreateProject"

	createdProject, err := u.DB.CreateProject(ctx, project)
	if err != nil {
		u.Log.Error("error creating project", "op", op, "error", err)

		return nil, err
	}

	err = u.Redis.InvalidateProjectCache(ctx, project.UserID, project.ID)
	if err != nil {
		u.Log.Error("error invalidating cache", "op", op, "error", err)
	}

	return createdProject, nil
}

func (u *UseCase) UpdateProject(ctx context.Context, project entity.UpdateProjectRequest, projectID int64, userID uuid.UUID) (*entity.Project, error) {
	const op = "storage.UpdateProject"

	updatedProject, err := u.DB.UpdateProject(ctx, project, projectID, userID)
	if err != nil {
		u.Log.Error("error updating project", "op", op, "error", err)

		return nil, err
	}

	err = u.Redis.InvalidateProjectCache(ctx, userID, projectID)
	if err != nil {
		u.Log.Error("error invalidating cache", "op", op, "error", err)
	}

	return updatedProject, nil
}

func (u *UseCase) DeleteProject(ctx context.Context, projectID int64, userID uuid.UUID) error {
	const op = "storage.DeleteProject"

	if projectID <= 0 {
		u.Log.Error("invalid projectID", "op", op, "projectID", projectID)

		return response.ErrInvalidArgument
	}

	err := u.DB.DeleteProject(ctx, projectID, userID)
	if err != nil {
		u.Log.Error("error deleting project", "op", op, "error", err)

		return err
	}

	err = u.Redis.InvalidateProjectCache(ctx, userID, projectID)
	if err != nil {
		u.Log.Error("error invalidating cache", "op", op, "error", err)
	}

	return nil
}

func (u *UseCase) GetProject(ctx context.Context, projectID int64, userID uuid.UUID, limit, offset uint64) (*entity.Project, error) {
	const op = "usecase.GetProject"

	isCache := true

	cachedProject, err := u.Redis.GetTempProject(ctx, userID, projectID)
	if err != nil {
		isCache = false
	}

	if isCache {
		return cachedProject, nil
	}

	project, err := u.DB.GetProject(ctx, projectID, userID)
	if err != nil {
		u.Log.Error("error getting project", "op", op, "error", err)

		return nil, err
	}

	groups, err := u.DB.GetProjectGroups(ctx, userID, projectID, limit, offset)
	if err != nil {
		u.Log.Error("error getting projet groups", "op", op, "error", err)

		return nil, err
	}

	tasks, err := u.DB.GetTasksByGroupIDs(ctx, extractGroupIDs(groups))
	if err != nil {
		u.Log.Error("error getting tasks", "op", op, "error", err)
		
		return nil, err
	}
	
	tags, err := u.DB.GetTagsByTaskIDs(ctx, extractTaskID(tasks))
	if err != nil {
		u.Log.Error("error getting tags", "op", op, "error", err)

		return nil, err
	}

	tagsByTask := make(map[int64][]*entity.Tag)
	for _, tag := range tags {
	    tagsByTask[tag.TaskID] = append(tagsByTask[tag.TaskID], tag)
	}
	
	tasksByGroup := make(map[int64][]*entity.Task)
	for _, task := range tasks {
	    task.Tags = tagsByTask[task.ID]

		if task.GroupID != nil {
            tasksByGroup[*task.GroupID] = append(tasksByGroup[*task.GroupID], task)
        }
	}

	for i, group := range groups {
    	groups[i].Tasks = tasksByGroup[group.ID]
	}

	project.Groups = groups
	
	err = u.Redis.SetTempProject(ctx, project)
	if err != nil {
		u.Log.Error("error setting cache", "op", op, "error", err)
	}

	return project, nil
}

func (u *UseCase) GetProjects(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]*entity.Project, error) {
	const op = "storage.GetProjects"

	projects, err := u.DB.GetProjects(ctx, userID, limit, offset)
	if err != nil {
		u.Log.Error("error getting projects", "op", op, "error", err)

		return nil, err
	}

	return projects, nil
}
