package apperr

import "errors"

type messageError struct {
	error
	message string
}

func (e messageError) Unwrap() error { return e.error }

func WithMessage(err error, message string) error {
	return messageError{
		error:   err,
		message: message,
	}
}

func Message(err error) string {
	if err == nil {
		return ""
	}
	if err, ok := errors.AsType[messageError](err); ok {
		return err.message
	}
	return "The operation could not complete successfully. Please contact you administrator."
}
