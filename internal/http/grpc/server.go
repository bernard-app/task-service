package grpc

import (
	"bernard/internal/usecase"
	"log/slog"

	taskv1 "github.com/bernard-app/bernard-protos/pkg/task_v1"
	"google.golang.org/grpc"
)

type TaskHandler struct {
	taskv1.UnimplementedTaskServiceServer
	uc *usecase.UseCase
}

func Register(gRPC *grpc.Server, uc *usecase.UseCase, log *slog.Logger) {
	th := &TaskHandler{
		uc: uc,
	}

	taskv1.RegisterTaskServiceServer(gRPC, th)
}
