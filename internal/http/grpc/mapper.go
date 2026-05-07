package grpc

import (
	"bernard/internal/domain/entity"

	taskv1 "github.com/bernard-app/bernard-protos/pkg/task_v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapTask(t *entity.Task) *taskv1.Task {
	if t == nil {
		return nil
	}

	grpcTags := make([]*taskv1.Tag, 0, len(t.Tags))

	for _, tag := range t.Tags {
		grpcTags = append(grpcTags, &taskv1.Tag{
			Id:    tag.ID,
			Name:   tag.Name,
			Color:  tag.Color,
			UserId: tag.UserID.String(),
		})
	}
	intPriority := int32(*t.Priority)

	var taskTime *timestamppb.Timestamp
	if t.StartTime != nil {
		taskTime = timestamppb.New(*t.StartTime)
	}

	var taskDeadline *timestamppb.Timestamp
	if t.Deadline != nil {
		taskDeadline = timestamppb.New(*t.Deadline)
	}

	return &taskv1.Task{
		TaskId:      t.ID,
		Name:        t.Name,
		Description: t.Description,
		Priority:    &intPriority,
		Tags:        grpcTags,
		Status:      t.Status,
		GroupId:     t.GroupID,
		ProjectId:   t.ProjectID,
		UserId:      t.UserID.String(),
		StartTime:   taskTime,
		Deadline:    taskDeadline,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
}

func mapGroup(t *entity.Group) *taskv1.Group {
	if t == nil {
		return nil
	}

	grpcTasks := make([]*taskv1.Task, 0, len(t.Tasks))

	for _, task := range t.Tasks {
		grpcTasks = append(grpcTasks, mapTask(task))
	}

	return &taskv1.Group{
		Id: t.ID,
		Name: t.Name,
		Tasks: grpcTasks,
		TaskCount: int64(t.TaskCount),
		ProjectId: t.ProjectID,
	}
}