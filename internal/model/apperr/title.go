package apperr

import "errors"

type titleError struct {
	error
	title string
}

func (e titleError) Unwrap() error { return e.error }

func WithTitle(err error, title string) error {
	return titleError{
		error: err,
		title: title,
	}
}

func Title(err error) string {
	if err == nil {
		return ""
	}
	if err, ok := errors.AsType[titleError](err); ok {
		return err.title
	}
	return "internal error"
}
