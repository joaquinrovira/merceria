package apperr

import "errors"

type publicError struct {
	error
}

func (e publicError) Unwrap() error { return e.error }

func WithPublic(err error) error {
	return publicError{error: err}
}

func Public(err error) error {
	if err == nil {
		return nil
	}
	if internal, ok := errors.AsType[publicError](err); ok {
		return internal.Unwrap()
	}
	return nil
}
