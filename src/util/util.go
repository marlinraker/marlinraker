package util

import "sync"

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

func StringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type ThreadSafe[T any] interface {
	Get() T
	Set(T)
}

type threadSafe[T any] struct {
	mutex sync.RWMutex
	value T
}

func NewThreadSafe[T any](value T) ThreadSafe[T] {
	return &threadSafe[T]{value: value}
}

func (ts *threadSafe[T]) Get() T {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	return ts.value
}

func (ts *threadSafe[T]) Set(value T) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	ts.value = value
}
