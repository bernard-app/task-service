package grpc

import (
	"bernard/internal/domain/entity"
	"context"

	taskv1 "github.com/bernard-app/bernard-protos/pkg/task_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (t *TaskHandler) CreateProject(ctx context.Context, req *taskv1.CreateProjectRequest) (*taskv1.CreateProjectResponse, error) {
	const op = "grpc.CreateProject"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("cannot extract userID from ctx", "op", op, "error", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	project := entity.Project{
		Name:        req.GetName(),
		Description: req.Description,
		UserID:      userID,
	}

	createdProject, err := t.uc.CreateProject(ctx, project)
	if err != nil {
		t.log.Error("cannot create project", "op", op, "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.CreateProjectResponse{
		Project: &taskv1.Project{
			Id:          createdProject.ID,
			Name:        createdProject.Name,
			Description: *createdProject.Description,
			UserId:      userID.String(),
		},
	}, nil
}

func (t *TaskHandler) UpdateProject(ctx context.Context, req *taskv1.UpdateProjectRequest) (*taskv1.UpdateProjectResponse, error) {
	const op = "grpc.UpdateProject"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("cannot extract userID from ctx", "op", op, "error", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	project := entity.UpdateProjectRequest{
		Name:        req.Name,
		Description: req.Description,
	}

	updatedProject, err := t.uc.UpdateProject(ctx, project, req.GetId(), userID)
	if err != nil {
		t.log.Error("cannot update project", "op", op, "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.UpdateProjectResponse{
		Project: &taskv1.Project{
			Id:          updatedProject.ID,
			Name:        updatedProject.Name,
			Description: *updatedProject.Description,
			UserId:      userID.String(),
		},
	}, nil
}

func (t *TaskHandler) DeleteProject(ctx context.Context, req *taskv1.DeleteProjectRequest) (*taskv1.DeleteProjectResponse, error) {
	const op = "grpc.DeleteProject"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("cannot extract userID from ctx", "op", op, "error", err)
		return &taskv1.DeleteProjectResponse{Success: false}, status.Error(codes.Unauthenticated, err.Error())
	}

	err = t.uc.DeleteProject(ctx, req.GetId(), userID)
	if err != nil {
		t.log.Error("cannot delete project", "op", op, "error", err)
		return &taskv1.DeleteProjectResponse{Success: false}, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.DeleteProjectResponse{Success: true}, nil
}

func (t *TaskHandler) GetProjectTree(ctx context.Context, req *taskv1.GetProjectTreeRequest) (*taskv1.GetProjectTreeResponse, error) {
	const op = "grpc.GetProjectTree"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("cannot extract userID from ctx", "op", op, "error", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	projectTree, err := t.uc.GetProjectTree(ctx, req.GetId(), userID)
	if err != nil {
		t.log.Error("cannot get project tree", "op", op, "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	var grpcGroups []*taskv1.Group

	for _, group := range projectTree.Groups {
		var grpcTasks []*taskv1.Task

		for _, task := range group.Tasks {
			grpcTask := &taskv1.Task{
				TaskId:    task.ID,
				Name:      task.Name,
				GroupId:   task.GroupID,
				UserId:    userID.String(),
				CreatedAt: timestamppb.New(task.CreatedAt),
				UpdatedAt: timestamppb.New(task.UpdatedAt),
			}

			if task.Description != nil {
				grpcTask.Description = task.Description
			}
			if task.Status != nil {
				grpcTask.Status = task.Status
			}
			if task.Priority != nil {
				priority := int32(*task.Priority)
				grpcTask.Priority = &priority
			}
			if task.StartTime != nil {
				grpcTask.StartTime = timestamppb.New(*task.StartTime)
			}
			if task.Deadline != nil {
				grpcTask.Deadline = timestamppb.New(*task.Deadline)
			}

			for _, tag := range task.Tags {
				grpcTag := &taskv1.Tag{
					Tag:    tag.ID,
					Name:   tag.Name,
					Color:  tag.Color,
					UserId: userID.String(),
				}
				grpcTask.Tags = append(grpcTask.Tags, grpcTag)
			}

			grpcTasks = append(grpcTasks, grpcTask)
		}

		grpcGroup := &taskv1.Group{
			Id:        group.ID,
			Name:      group.Name,
			ProjectId: group.ProjectID,
			TaskCount: int32(group.TaskCount),
			Tasks:     grpcTasks,
		}

		grpcGroups = append(grpcGroups, grpcGroup)
	}

	grpcProject := &taskv1.Project{
		Id:        projectTree.ID,
		Name:      projectTree.Name,
		UserId:    userID.String(),
		TaskCount: int32(projectTree.TaskCount),
		Groups:    grpcGroups,
	}

	if projectTree.Description != nil {
		grpcProject.Description = *projectTree.Description
	}

	return &taskv1.GetProjectTreeResponse{
		Project: grpcProject,
	}, nil
}

func (t *TaskHandler) GetProjects(ctx context.Context, req *taskv1.GetProjectsRequest) (*taskv1.GetProjectsResponse, error) {
	const op = "grpc.GetProjects"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("cannot extract userID from ctx", "op", op, "error", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	projects, err := t.uc.GetProjects(ctx, userID, req.GetLimit(), req.GetOffset())
	if err != nil {
		t.log.Error("cannot get projects", "op", op, "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var grpcProjects []*taskv1.Project

	for _, project := range projects {
		grpcProject := &taskv1.Project{
			Id:          project.ID,
			Name:        project.Name,
			Description: *project.Description,
			UserId:      userID.String(),
		}

		grpcProjects = append(grpcProjects, grpcProject)
	}

	return &taskv1.GetProjectsResponse{
		Projects: grpcProjects,
	}, nil
}
