package grpc

import (
	"bernard/internal/domain/entity"
	"context"

	taskv1 "github.com/bernard-app/bernard-protos/pkg/task_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (t *TaskHandler) CreateGroup(ctx context.Context, req *taskv1.CreateGroupRequest) (*taskv1.CreateGroupResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	if req.GetProjectId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	
	group := entity.Group{
		Name:      req.GetName(),
		ProjectID: req.GetProjectId(),
		UserID:    userID,
	}

	createdGroup, err := t.uc.CreateGroup(ctx, group)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &taskv1.CreateGroupResponse{
		Group: &taskv1.Group{
			Id:        createdGroup.ID,
			Name:      createdGroup.Name,
			ProjectId: createdGroup.ProjectID,
		},
	}, nil
}

func (t *TaskHandler) UpdateGroup(ctx context.Context, req *taskv1.UpdateGroupRequest) (*taskv1.UpdateGroupResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	group := entity.UpdateGroupRequest{
		Name:      req.Name,
		ProjectID: req.ProjectId,
	}

	updatedGroup, err := t.uc.UpdateGroup(ctx, group, userID, req.GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &taskv1.UpdateGroupResponse{
		Group: &taskv1.Group{
			Id:        updatedGroup.ID,
			Name:      updatedGroup.Name,
			ProjectId: updatedGroup.ProjectID,
		},
	}, nil
}

func (t *TaskHandler) DeleteGroup(ctx context.Context, req *taskv1.DeleteGroupRequest) (*taskv1.DeleteGroupResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return &taskv1.DeleteGroupResponse{Success: false}, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	err = t.uc.DeleteGroup(ctx, req.GetId(), userID)
	if err != nil {
		return &taskv1.DeleteGroupResponse{Success: false}, status.Error(codes.Internal, "internal server error")
	}

	return &taskv1.DeleteGroupResponse{Success: true}, nil
}

func (t *TaskHandler) GetGroup(ctx context.Context, req *taskv1.GetGroupRequest) (*taskv1.GetGroupResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.GetGroupId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	if req.GetLimit() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "limit is required")
	}
	
	groupID := req.GetGroupId()
	limit := req.GetLimit()
	offset := req.GetOffset()

	group, err := t.uc.GetGroup(ctx, groupID, userID, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &taskv1.GetGroupResponse{
		Group: &taskv1.Group{
			Id:        group.ID,
			Name:      group.Name,
			TaskCount: int64(group.TaskCount),
			ProjectId: group.ProjectID,
		},
	}, nil
}

func (t *TaskHandler) GetProjectGroups(ctx context.Context, req *taskv1.GetProjectGroupsRequest) (*taskv1.GetProjectGroupsResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.GetProjectId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	if req.GetLimit() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "limit is required")
	}
	
	projectID := req.GetProjectId()
	limit := req.GetLimit()
	offset := req.GetOffset()

	groups, err := t.uc.GetProjectGroups(ctx, projectID, userID, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "internals server error")
	}

	var grpcGroups []*taskv1.Group

	for _, group := range groups {
		grpcGroups = append(grpcGroups, mapGroup(group))
	}

	return &taskv1.GetProjectGroupsResponse{
		Groups: grpcGroups,
	}, nil
}
