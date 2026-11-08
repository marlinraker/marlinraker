package util

import "fmt"

type ExecutorError struct {
	Message string
	Code    int
}

func (err *ExecutorError) Error() string {
	return err.Message
}

func NewError(code int, message string) error {
	return &ExecutorError{message, code}
}

func NewErrorf(code int, format string, a ...interface{}) error {
	return NewError(code, fmt.Sprintf(format, a...))
}
