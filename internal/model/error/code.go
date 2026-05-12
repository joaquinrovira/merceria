package apperr

import "errors"

type codeError struct {
	error
	code string
}

func (e codeError) Unwrap() error { return e.error }

func WithCode(err error, code string) error {
	return codeError{
		error: err,
		code:  code,
	}
}

func Code(err error) string {
	if err == nil {
		return ""
	}
	if err, ok := errors.AsType[codeError](err); ok {
		return err.code
	}
	return "UNKNOWN"
}
