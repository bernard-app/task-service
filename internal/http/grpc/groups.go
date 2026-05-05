package grpc

import (
	"bernard/internal/domain/entity"
	"context"

	taskv1 "github.com/bernard-app/bernard-protos/pkg/task_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (t *TaskHandler) CreateGroup(ctx context.Context, req *taskv1.CreateGroupRequest) (*taskv1.CreateGroupResponse, error) {
	const op = "grpc.CreateGroup"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user id", "op", op, "error", err, "user_id", userID)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	group := entity.Group{
		Name:      req.GetName(),
		ProjectID: req.GetProjectId(),
		UserID:    userID,
	}

	createdGroup, err := t.uc.CreateGroup(ctx, group)
	if err != nil {
		t.log.Error("Failed to create group", "op", op, "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, err.Error())
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
	const op = "grpc.UpdateGroup"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user id", "op", op, "error", err, "user_id", userID)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	group := entity.UpdateGroupRequest{
		Name:      req.Name,
		ProjectID: req.ProjectId,
	}

	updatedGroup, err := t.uc.UpdateGroup(ctx, group, userID, req.GetId())
	if err != nil {
		t.log.Error("Failed to update group", "op", op, "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, err.Error())
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
	const op = "grpc.DeleteGroup"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user id", "op", op, "error", err, "user_id", userID)
		return &taskv1.DeleteGroupResponse{Success: false}, status.Error(codes.Unauthenticated, err.Error())
	}

	err = t.uc.DeleteGroup(ctx, req.GetId(), userID)
	if err != nil {
		t.log.Error("Failed to delete group", "op", op, "error", err, "user_id", userID)
		return &taskv1.DeleteGroupResponse{Success: false}, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.DeleteGroupResponse{Success: true}, nil
}
