package response

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrInternal            = errors.New("internal error")
)