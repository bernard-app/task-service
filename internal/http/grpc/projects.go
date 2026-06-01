package grpc

import (
	"bernard/internal/domain/entity"
	"context"

	taskv1 "github.com/bernard-app/bernard-protos/pkg/task_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (t *TaskHandler) CreateProject(ctx context.Context, req *taskv1.CreateProjectRequest) (*taskv1.CreateProjectResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	
	project := entity.Project{
		Name:        req.GetName(),
		Description: req.Description,
		UserID:      userID,
	}

	createdProject, err := t.uc.CreateProject(ctx, project)
	if err != nil {
		return nil, HandleError(err)
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
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	project := entity.UpdateProjectRequest{
		Name:        req.Name,
		Description: req.Description,
	}

	updatedProject, err := t.uc.UpdateProject(ctx, project, req.GetId(), userID)
	if err != nil {
		return nil, HandleError(err)
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
	userID, err := extractUserID(ctx)
	if err != nil {
		return &taskv1.DeleteProjectResponse{Success: false}, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	err = t.uc.DeleteProject(ctx, req.GetId(), userID)
	if err != nil {
		return &taskv1.DeleteProjectResponse{Success: false}, HandleError(err)
	}

	return &taskv1.DeleteProjectResponse{Success: true}, nil
}

func (t *TaskHandler) GetProject(ctx context.Context, req *taskv1.GetProjectRequest) (*taskv1.GetProjectResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	project, err := t.uc.GetProject(ctx, req.GetProjectId(), userID, req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, HandleError(err)
	}

	return &taskv1.GetProjectResponse{
		Project: mapProject(project),
	}, nil
}

func (t *TaskHandler) GetProjects(ctx context.Context, req *taskv1.GetProjectsRequest) (*taskv1.GetProjectsResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	projects, err := t.uc.GetProjects(ctx, userID, req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, HandleError(err)
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
