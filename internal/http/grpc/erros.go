package grpc

import (
	"bernard/pkg/response"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleError(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := status.FromError(err); ok {
		return err
	}

	switch {
	case errors.Is(err, response.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, response.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())

	case errors.Is(err, response.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, response.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
