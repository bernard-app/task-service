package grpc

import (
	"bernard/internal/usecase"
	"log/slog"

	taskv1 "github.com/bernard-app/bernard-protos/pkg/task_v1"
	"google.golang.org/grpc"
)

type TaskHandler struct {
	taskv1.UnimplementedTaskServiceServer
	uc  *usecase.UseCase
	log *slog.Logger
}

func Register(gRPC *grpc.Server, uc *usecase.UseCase, log *slog.Logger) *TaskHandler {
	th := &TaskHandler{
		uc:  uc,
		log: log,
	}
	taskv1.RegisterTaskServiceServer(gRPC, th)
	return th
}
