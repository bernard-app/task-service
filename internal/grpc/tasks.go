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
		GroupID:     req.GetGroupId(),
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

	taskID := req.GetTaskId()

	var startTime time.Time
	if req.GetStartTime() != nil {
		startTime = req.GetStartTime().AsTime()
	}

	var deadline time.Time
	if req.GetDeadline() != nil {
		deadline = req.GetDeadline().AsTime()
	}

	task := entity.UpdateTaskRequest{
		Name:        req.Name,
		Description: req.Description,
		Priority:    utils.Ptr(int(req.GetPriority())),
		Status:      req.Status,
		GroupID:     req.GroupId,
		StartTime:   &startTime,
		Deadline:    &deadline,
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

func (t *TaskHandler) GetListTask(ctx context.Context, req *taskv1.GetListTasksRequest) (*taskv1.GetListTasksResponse, error) {
	const op = "grpc.GetListTask"

	userID, err := extractUserID(ctx)
	if err != nil {
		t.log.Error("Failed to extract user ID", "error", err, "operation", op)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	priority := int(req.GetPriority())
	tag := req.GetTagIds()
	from := req.GetFrom().AsTime()
	to := req.GetTo().AsTime()
	limit := req.GetLimit()
	offset := req.GetOffset()

	tasks, err := t.uc.GetListTask(ctx, userID, &priority, &tag, &from, &to, limit, offset)
	if err != nil {
		t.log.Error("error getting tasks", "operation", op, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get tasks: %v", err)
	}

	tasksResponse := make([]*taskv1.TasksTab, 0, len(tasks))
	for _, userTask := range tasks {
		taskTab := taskv1.TasksTab{
			Task:        mapTask(&userTask.Task),
			GroupName:   userTask.GroupName,
			ProjectId:   userTask.ProjectID,
			ProjectName: userTask.ProjectName,
		}
		tasksResponse = append(tasksResponse, &taskTab)
	}

	return &taskv1.GetListTasksResponse{
		Tasks: tasksResponse,
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
