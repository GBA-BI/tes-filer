package error

import (
	"errors"
	"fmt"
)

type Error struct {
	Code    string
	Message string
	Inner   error
}

func (e *Error) Error() string {
	return fmt.Sprintf("error code: %s, message: %s", e.Code, e.Message)
}

func (e *Error) Is(target error) bool {
	if e == nil {
		return false
	}
	return errors.Is(e.Inner, target)
}

type ErrorCode int

const (
	ErrUnknown ErrorCode = iota + 1000001
	ErrInvalidArgument
	ErrNotFound
	ErrPermissionDenied
	// Add more error codes here...
)

func wrapError(code ErrorCode, msg string, err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    fmt.Sprintf("%d", code),
		Message: fmt.Sprintf("%s: %s", msg, err.Error()),
		Inner:   err,
	}
}

// NewInvalidArgumentError ...
func NewInvalidArgumentError(param, content string) *Error {
	return wrapError(ErrInvalidArgument, "invalid argument", fmt.Errorf("%s %s is invalid", param, content))
}

// NewNotFoundError ...
func NewNotFoundError(param, content string) *Error {
	return wrapError(ErrNotFound, "object not found", fmt.Errorf("%s %s not found", param, content))
}

// NewInternalError ...
func NewInternalError(err error) *Error {
	return wrapError(ErrUnknown, "server internal error", err)
}

// NewPermissionDeniedError ...
func NewPermissionDeniedError(param, content string) *Error {
	return wrapError(ErrPermissionDenied, "no permission to do", fmt.Errorf("%s %s not permission", param, content))
}
