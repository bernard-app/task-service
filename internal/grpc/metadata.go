package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func extractUserID(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("X-User-ID")
	if len(values) == 0 {
		return uuid.Nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	userID, err := uuid.Parse(values[0])
	if err != nil {
		return uuid.Nil, status.Error(codes.InvalidArgument, "invalid metadata")
	}

	return userID, nil
}
