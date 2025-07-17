package token

import "errors"

// UnexpectedError represents an unexpected error so that we could differentiate it from expected errors.
type UnexpectedError struct {
	err error
}

// Error returns a string representation for an unexpected error.
func (ue *UnexpectedError) Error() string {
	return ue.err.Error()
}

// IsUnexpectedError returns true if its argument is an unexpected error.
func IsUnexpectedError(err error) bool {
	var unexpectedError *UnexpectedError
	return errors.As(err, &unexpectedError)
}

func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}

func recoverPanics(
	err *error, //nolint:gocritic // we need the pointer as we replace the error with a panic
) {
	if r := recover(); r != nil {
		*err = &UnexpectedError{err: r.(error)}
	}
}
