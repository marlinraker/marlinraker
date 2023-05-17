package util

import (
	log "github.com/sirupsen/logrus"
	"runtime/debug"
	"strings"
)

type ExecutorError struct {
	Message string
	Code    int
}

func (err *ExecutorError) Error() string {
	return err.Message
}

func NewError(message string, code int) error {
	return &ExecutorError{message, code}
}

func LogError(err error) {
	stack := string(debug.Stack())
	lines := strings.Split(stack, "\n")
	lines = append(lines[:1], lines[5:len(lines)-1]...)
	stack = strings.Join(lines, "\n")
	log.Error("Error: " + err.Error() + "\n" + stack)
}
