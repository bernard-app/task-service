package grpc

import (
	"bernard/internal/domain/entity"
	"context"

	taskv1 "github.com/bernard-app/bernard-protos/pkg/task_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
			Description: createdProject.Description,
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
			Description: updatedProject.Description,
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

func (t *TaskHandler) GetProject(ctx context.Context, req *taskv1.GetProjectRequest) (*taskv1.GetProjectResponse, error) {
	const op = "grpc.GetProject"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("cannot extract userID from ctx", "op", op, "error", err)

		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	project, err := t.uc.GetProject(ctx, req.GetProjectId(), userID, req.GetLimit(), req.GetOffset())
	if err != nil {
		t.log.Error("cannot get project tree", "op", op, "error", err)
	
		return nil, status.Error(codes.Internal, err.Error())
	}

	var grpcGroups []*taskv1.Group

	for _, group := range project.Groups {
		grpcGroups = append(grpcGroups, mapGroup(group))
	}

	grpcProject := &taskv1.Project{
		Id:        project.ID,
		Name:      project.Name,
		UserId:    userID.String(),
		TaskCount: int64(project.TaskCount),
		Groups:    grpcGroups,
	}

	if project.Description != nil {
		grpcProject.Description = project.Description
	}

	return &taskv1.GetProjectResponse{
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
			Description: project.Description,
			UserId:      userID.String(),
		}

		grpcProjects = append(grpcProjects, grpcProject)
	}

	return &taskv1.GetProjectsResponse{
		Projects: grpcProjects,
	}, nil
}
