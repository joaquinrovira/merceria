package apperr

import (
	"errors"
	"net/http"
)

type statusError struct {
	error
	status int
}

func (e statusError) Unwrap() error { return e.error }

func WithStatus(err error, status int) error {
	return statusError{
		error:  err,
		status: status,
	}
}

func Status(err error) int {
	if err == nil {
		return 0
	}
	if err, ok := errors.AsType[statusError](err); ok {
		return err.status
	}
	return http.StatusInternalServerError
}
