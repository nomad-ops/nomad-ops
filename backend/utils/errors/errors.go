package errors

import (
	"errors"
)

var (
	// A list of common simple errors

	// ErrNotImplemented ...
	ErrNotImplemented = errors.New("not implemented")
	// ErrAbort ...
	ErrAbort = errors.New("aborted")
	// ErrShutdown ...
	ErrShutdown = errors.New("shutdown")
	// ErrNotFound ...
	ErrNotFound = errors.New("not found")
	// ErrCallerError ...
	ErrCallerError = errors.New("client made an error")
	// ErrReceiverError ...
	ErrReceiverError = errors.New("receiver made an error")
	// ErrInvalid ...
	ErrInvalid = errors.New("invalid")
)

// TemporaryError ...
type TemporaryError interface {
	error
	IsTemporary() bool
}

type tempError struct {
	err  error
	temp bool
}

func (e *tempError) Error() string {

	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

// IsTemporary ...
func IsTemporary(err error) bool {
	if t, ok := err.(TemporaryError); ok {
		return t.IsTemporary()
	}
	return false
}

func (e *tempError) IsTemporary() bool {
	if e == nil {
		return false
	}
	return e.temp
}

// CreateTemporaryError ...
func CreateTemporaryError(err error) TemporaryError {
	if err == nil {
		return nil
	}
	return &tempError{
		err:  err,
		temp: true,
	}
}
