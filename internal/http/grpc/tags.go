package grpc

import (
	"bernard/internal/domain/entity"
	"context"

	taskv1 "github.com/bernard-app/bernard-protos/pkg/task_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (t *TaskHandler) CreateTag(ctx context.Context, req *taskv1.CreateTagRequest) (*taskv1.CreateTagResponse, error) {
	const op = "grpc.CreateTag"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user id from ctx", "operation", op, "error", err, "user_id", userID)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	tag := entity.Tag{
		Name:   req.GetName(),
		Color:  req.GetColor(),
		UserID: userID,
	}

	createdTag, err := t.uc.CreateTag(ctx, tag)
	if err != nil {
		t.log.Error("Failed to create tag", "operation", op, "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.CreateTagResponse{
		Tag: &taskv1.Tag{
			Tag:    createdTag.ID,
			Name:   createdTag.Name,
			Color:  createdTag.Color,
			UserId: createdTag.UserID.String(),
		},
	}, nil
}

func (t *TaskHandler) UpdateTag(ctx context.Context, req *taskv1.UpdateTagRequest) (*taskv1.UpdateTagResponse, error) {
	const op = "grpc.UpdateTag"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user id from ctx", "operation", op, "error", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	tag, err := t.uc.UpdateTag(ctx, req.GetId(), req.Name, req.Color, userID)
	if err != nil {
		t.log.Error("Failed to update tag", "operation", op, "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.UpdateTagResponse{
		Tag: &taskv1.Tag{
			Tag:    tag.ID,
			Name:   tag.Name,
			Color:  tag.Color,
			UserId: tag.UserID.String(),
		},
	}, nil
}

func (t *TaskHandler) DeleteTag(ctx context.Context, req *taskv1.DeleteTagRequest) (*taskv1.DeleteTagResponse, error) {
	const op = "grpc.DeleteTag"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user id from ctx", "operation", op, "error", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = t.uc.DeleteTag(ctx, req.GetId(), userID)
	if err != nil {
		t.log.Error("Failed to delete tag", "operation", op, "error", err, "user_id", userID)
		return &taskv1.DeleteTagResponse{Success: false}, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.DeleteTagResponse{Success: true}, nil
}

func (t *TaskHandler) GetTag(ctx context.Context, req *taskv1.GetTagRequest) (*taskv1.GetTagResponse, error) {
	const op = "grpc.GetTag"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user id from ctx", "operation", op, "error", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	task, err := t.uc.GetTag(ctx, req.GetId(), userID)
	if err != nil {
		t.log.Error("Failed to get tag", "operation", op, "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.GetTagResponse{
		Tag: &taskv1.Tag{
			Tag:    task.ID,
			Name:   task.Name,
			Color:  task.Color,
			UserId: task.UserID.String(),
		},
	}, nil
}

func (t *TaskHandler) GetTagList(ctx context.Context, req *taskv1.GetTagListRequest) (*taskv1.GetTagListResponse, error) {
	const op = "grpc.GetTagList"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user id from ctx", "operation", op, "error", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	tags, err := t.uc.GetTagList(ctx, userID, req.GetLimit(), req.GetOffset())
	if err != nil {
		t.log.Error("Failed to get tags", "operation", op, "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var grpcTags []*taskv1.Tag

	for _, tag := range tags {
		grpcTags = append(grpcTags, &taskv1.Tag{
			Tag:    tag.ID,
			Name:   tag.Name,
			Color:  tag.Color,
			UserId: tag.UserID.String(),
		})
	}

	return &taskv1.GetTagListResponse{
		Tags: grpcTags,
	}, nil
}

func (t *TaskHandler) AddTagToTask(ctx context.Context, req *taskv1.AddTagToTaskRequest) (*taskv1.AddTagToTaskResponse, error) {
	const op = "grpc.AddTagToTask"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user id from ctx", "operation", op, "error", err)
		return &taskv1.AddTagToTaskResponse{Success: false}, status.Error(codes.Unauthenticated, err.Error())
	}

	err = t.uc.AddTagToTask(ctx, req.GetTaskId(), req.GetTagId(), userID)
	if err != nil {
		t.log.Error("Failed to add tag to task", "operation", op, "error", err, "user_id", userID)
		return &taskv1.AddTagToTaskResponse{Success: false}, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.AddTagToTaskResponse{Success: true}, nil
}

func (t *TaskHandler) RemoveTagFromTask(ctx context.Context, req *taskv1.RemoveTagFromTaskRequest) (*taskv1.RemoveTagFromTaskResponse, error) {
	const op = "grpc.RemoveTagFromTask"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user id from ctx", "operation", op, "error", err)
		return &taskv1.RemoveTagFromTaskResponse{Success: false}, status.Error(codes.Unauthenticated, err.Error())
	}

	err = t.uc.RemoveTagFromTask(ctx, req.GetTaskId(), req.GetTagId(), userID)
	if err != nil {
		t.log.Error("Failed to remove tag from task", "operation", op, "error", err, "user_id", userID)
		return &taskv1.RemoveTagFromTaskResponse{Success: false}, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.RemoveTagFromTaskResponse{Success: true}, nil
}
