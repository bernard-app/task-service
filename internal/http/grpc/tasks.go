package grpc

import (
	"bernard/internal/domain/entity"
	"bernard/utils"
	"context"
	"time"

	taskv1 "github.com/bernard-app/bernard-protos/pkg/task_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (t *TaskHandler) CreateTask(ctx context.Context, req *taskv1.CreateTaskRequest) (*taskv1.CreateTaskResponse, error) {
	const op = "grpc.CreateTask"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user ID", "error", err, "operation", op)

		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	var startTime time.Time
	if req.GetStartTime() != nil {
		startTime = req.GetStartTime().AsTime()
	}

	var deadline time.Time
	if req.GetDeadline() != nil {
		deadline = req.GetDeadline().AsTime()
	}

	task := entity.Task{
		Name:        req.GetName(),
		Description: req.Description,
		Priority:    utils.Ptr(int(req.GetPriority())),
		Status:      req.Status,
		StartTime:   &startTime,
		Deadline:    &deadline,
		GroupID:     req.GroupId,
		ProjectID:   req.ProjectId,
		UserID:      userID,
	}

	createdTask, err := t.uc.CreateTask(ctx, task, req.GetTagIds())
	if err != nil {
		t.log.Error("Failed to create task", "error", err, "operation", op)

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.CreateTaskResponse{
		Task: mapTask(createdTask),
	}, nil
}

func (t *TaskHandler) UpdateTask(ctx context.Context, req *taskv1.UpdateTaskRequest) (*taskv1.UpdateTaskResponse, error) {
	const op = "grpc.UpdateTask"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user ID", "error", err, "operation", op)

		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	var task entity.UpdateTaskRequest

	taskID := req.GetTaskId()

	if req.Description != nil {
		task.Description = req.Description
	}

	if req.Priority != nil {
		priority := int(req.GetPriority())
		task.Priority = &priority
	}

	if req.Status != nil {
		task.Status = req.Status
	}

	if req.TagIds != nil {
		tags := req.GetTagIds()
		task.TagsIDs = &tags
	}

	if req.GroupId != nil {
		task.GroupID = req.GroupId
	}

	if req.StartTime != nil {
		startTime := req.GetStartTime().AsTime()
		task.StartTime = &startTime
	}

	if req.Deadline != nil {
		deadline := req.GetDeadline().AsTime()
		task.Deadline = &deadline
	}

	updatedTask, err := t.uc.UpdateTask(ctx, task, userID, taskID)
	if err != nil {
		t.log.Error("Failed to update task", "error", err, "operation", op)

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.UpdateTaskResponse{
		Task: mapTask(updatedTask),
	}, nil
}

func (t *TaskHandler) DeleteTask(ctx context.Context, req *taskv1.DeleteTaskRequest) (*taskv1.DeleteTaskResponse, error) {
	const op = "grpc.DeleteTask"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user ID", "error", err, "operation", op)

		return &taskv1.DeleteTaskResponse{Success: false}, status.Error(codes.Unauthenticated, err.Error())
	}

	taskID := req.GetTaskId()

	err = t.uc.DeleteTask(ctx, userID, taskID)
	if err != nil {
		t.log.Error("Failed to delete task", "error", err, "operation", op)

		return &taskv1.DeleteTaskResponse{Success: false}, status.Error(codes.InvalidArgument, err.Error())
	}

	return &taskv1.DeleteTaskResponse{
		Success: true,
	}, nil
}

func (t *TaskHandler) GetTask(ctx context.Context, req *taskv1.GetTaskRequest) (*taskv1.GetTaskResponse, error) {
	const op = "grpc.GetTask"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user ID", "error", err, "operation", op)

		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	taskID := req.GetTaskId()

	task, err := t.uc.GetTask(ctx, taskID, userID)
	if err != nil {
		t.log.Error("failed to get task", "operation", op, "error", err)

		return nil, status.Errorf(codes.Internal, "failed to get task: %v", err)
	}

	return &taskv1.GetTaskResponse{
		Task: mapTask(task),
	}, nil
}

func (t *TaskHandler) GetListTask(ctx context.Context, req *taskv1.GetListTaskRequest) (*taskv1.GetListTaskResponse, error) {
	const op = "grpc.GetListTask"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user ID", "error", err, "operation", op)

		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	var taskFilter entity.TasksFilter

	taskFilter.UserID = userID
	taskFilter.Limit = req.Limit
	taskFilter.Offset = req.Offset

	if req.Priority != nil {
		priority := int(req.GetPriority())
		taskFilter.Priority = &priority
	}

	if req.TagIds != nil {
		taskFilter.Tag = req.TagIds
	}
	if req.From != nil {
		from := req.From.AsTime()
		taskFilter.From = &from
	}

	if req.To != nil {
		to := req.To.AsTime()
		taskFilter.To = &to
	}

	tasks, err := t.uc.GetListTask(ctx, taskFilter)
	if err != nil {
		t.log.Error("error getting tasks", "operation", op, "error", err)

		return nil, status.Errorf(codes.Internal, "failed to get tasks: %v", err)
	}

	tasksResponse := make([]*taskv1.Task, 0, len(tasks))
	for _, userTask := range tasks {
		tasksResponse = append(tasksResponse, mapTask(userTask))
	}

	return &taskv1.GetListTaskResponse{
		Tasks: tasksResponse,
	}, nil
}

func (t *TaskHandler) GetGroupTasks(ctx context.Context, req *taskv1.GetGroupTasksRequest) (*taskv1.GetGroupTasksResponse, error) {
	const op = "grpc.GetGroupTasks"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user ID", "error", err, "operation", op)

		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	tasks, err := t.uc.GetGroupTasks(ctx, req.GetGroupId(), userID, req.GetLimit(), req.GetOffset())
	if err != nil {
		t.log.Error("Failed to get group tasks", "op", op, "error", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	var grpcTasks []*taskv1.Task

	for _, task := range tasks {
		grpcTasks = append(grpcTasks, mapTask(task))
	}

	return &taskv1.GetGroupTasksResponse{
		Tasks: grpcTasks,
	}, nil
}

func (t *TaskHandler) ArchiveTask(ctx context.Context, req *taskv1.ArchiveTaskRequest) (*taskv1.ArchiveTaskResponse, error) {
	const op = "grpc.ArchiveTask"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user ID", "error", err, "operation", op)

		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	taskID := req.TaskId

	task, err := t.uc.ArchiveTask(ctx, userID, taskID)
	if err != nil {
		t.log.Error("failed to archive task", "error", err, "operation", op)

		return nil, status.Errorf(codes.Internal, "failed to archive task: %v", err)
	}

	return &taskv1.ArchiveTaskResponse{
		Task: mapTask(task),
	}, nil
}
