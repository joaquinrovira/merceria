package apperr

import (
	"errors"
	"net/http"
)

type overrideError struct {
	error
	fn func(w http.ResponseWriter, r *http.Request)
}

func (e overrideError) Unwrap() error { return e.error }

func WithOverride(err error, fn func(w http.ResponseWriter, r *http.Request)) error {
	return overrideError{error: err, fn: fn}
}

func Override(err error) func(w http.ResponseWriter, r *http.Request) {
	if internal, ok := errors.AsType[overrideError](err); ok {
		return internal.fn
	}
	return nil
}
