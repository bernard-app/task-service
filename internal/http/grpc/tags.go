package grpc

import (
	"bernard/internal/domain/entity"
	"context"

	taskv1 "github.com/bernard-app/bernard-protos/pkg/task_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (t *TaskHandler) CreateTag(ctx context.Context, req *taskv1.CreateTagRequest) (*taskv1.CreateTagResponse, error) {	
	userID, err := extractUserID(ctx)
	
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	if req.Color == "" {
		return nil, status.Error(codes.InvalidArgument, "color is required")
	}

	if req.ProjectId == 0 {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	
	tag := entity.Tag{
		Name:      req.GetName(),
		Color:     req.GetColor(),
		UserID:    userID,
		ProjectID: req.ProjectId,
	}

	createdTag, err := t.uc.CreateTag(ctx, tag)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "internal server error")
	}

	return &taskv1.CreateTagResponse{
		Tag: &taskv1.Tag{
			Id:       createdTag.ID,
			Name:      createdTag.Name,
			Color:     createdTag.Color,
			UserId:    createdTag.UserID.String(),
			ProjectId: createdTag.ProjectID,
		},
	}, nil
}

func (t *TaskHandler) UpdateTag(ctx context.Context, req *taskv1.UpdateTagRequest) (*taskv1.UpdateTagResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	tag, err := t.uc.UpdateTag(ctx, req.GetId(), req.Name, req.Color, userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "internal server error")
	}

	return &taskv1.UpdateTagResponse{
		Tag: &taskv1.Tag{
			Id:    tag.ID,
			Name:   tag.Name,
			Color:  tag.Color,
			UserId: tag.UserID.String(),
		},
	}, nil
}

func (t *TaskHandler) DeleteTag(ctx context.Context, req *taskv1.DeleteTagRequest) (*taskv1.DeleteTagResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	err = t.uc.DeleteTag(ctx, req.GetId(), userID)
	if err != nil {
		return &taskv1.DeleteTagResponse{Success: false}, status.Error(codes.InvalidArgument, "internal server error")
	}

	return &taskv1.DeleteTagResponse{Success: true}, nil
}

func (t *TaskHandler) GetTag(ctx context.Context, req *taskv1.GetTagRequest) (*taskv1.GetTagResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	task, err := t.uc.GetTag(ctx, req.GetId(), userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "interanl server error")
	}

	return &taskv1.GetTagResponse{
		Tag: &taskv1.Tag{
			Id:    task.ID,
			Name:   task.Name,
			Color:  task.Color,
			UserId: task.UserID.String(),
		},
	}, nil
}

func (t *TaskHandler) GetTagList(ctx context.Context, req *taskv1.GetTagListRequest) (*taskv1.GetTagListResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	tags, err := t.uc.GetTagList(ctx, userID, req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var grpcTags []*taskv1.Tag

	for _, tag := range tags {
		grpcTags = append(grpcTags, &taskv1.Tag{
			Id:    tag.ID,
			Name:   tag.Name,
			Color:  tag.Color,
			UserId: tag.UserID.String(),
		})
	}

	return &taskv1.GetTagListResponse{
		Tags: grpcTags,
	}, nil
}

func (t *TaskHandler) GetTaskTags(ctx context.Context, req *taskv1.GetTaskTagsRequest) (*taskv1.GetTaskTagsResponse, error) {
	tags, err := t.uc.GetTaskTags(ctx, req.GetTaskId())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var grpcTags []*taskv1.Tag

	for _, tag := range tags {
		grpcTags = append(grpcTags, &taskv1.Tag{
			Id:       tag.ID,
			Name:      tag.Name,
			Color:     tag.Color,
			ProjectId: tag.ProjectID,
			UserId:    tag.UserID.String(),
		})
	}

	return &taskv1.GetTaskTagsResponse{
		Tags: grpcTags,
	}, nil
}

func (t *TaskHandler) AddTagsToTask(ctx context.Context, req *taskv1.AddTagsToTaskRequest) (*taskv1.AddTagsToTaskResponse, error) {
	_, err := extractUserID(ctx)
	if err != nil {
		return &taskv1.AddTagsToTaskResponse{Succes: false}, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	err = t.uc.AddTagsToTask(ctx, req.GetTagsIds(), req.GetTaskId())
	if err != nil {
		return &taskv1.AddTagsToTaskResponse{Succes: false}, status.Error(codes.Internal, "internal server error")
	}

	return &taskv1.AddTagsToTaskResponse{Succes: true}, nil
}

func (t *TaskHandler) RemoveTagsFromTask(ctx context.Context, req *taskv1.RemoveTagsFromTaskRequest) (*taskv1.RemoveTagsFromTaskResponse, error) {
	_, err := extractUserID(ctx)
	if err != nil {
		return &taskv1.RemoveTagsFromTaskResponse{Success: false}, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	err = t.uc.RemoveTagsFromTask(ctx, req.GetTagsIds(), req.GetTaskId())
	if err != nil {
		return &taskv1.RemoveTagsFromTaskResponse{Success: false}, status.Error(codes.Internal, "internal server error")
	}

	return &taskv1.RemoveTagsFromTaskResponse{Success: true}, nil
}

func (t *TaskHandler) RemoveAllTagsFromTask(ctx context.Context, req *taskv1.RemoveAllTagsFromTaskRequest) (*taskv1.RemoveAllTagsFromTaskResponse, error) {
	_, err := extractUserID(ctx)
	if err != nil {
		return &taskv1.RemoveAllTagsFromTaskResponse{Success: false}, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	err = t.uc.RemoveAllTagsFromTask(ctx, req.GetTaskId())
	if err != nil {
		return &taskv1.RemoveAllTagsFromTaskResponse{Success: false}, status.Error(codes.Internal, "internal server error")
	}

	return &taskv1.RemoveAllTagsFromTaskResponse{Success: true}, nil
}