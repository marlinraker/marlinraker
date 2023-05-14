package util

import (
	log "github.com/sirupsen/logrus"
	"runtime/debug"
	"strings"
	"sync"
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

func StringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type ThreadSafe[T any] interface {
	Load() T
	Store(T) T
	Do(func(T) T) T
	Get() T
	Set(T) T
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type threadSafe[T any] struct {
	mutex sync.RWMutex
	value T
}

func NewThreadSafe[T any](value T) ThreadSafe[T] {
	return &threadSafe[T]{value: value}
}

func (ts *threadSafe[T]) Load() T {
	ts.RLock()
	defer ts.RUnlock()
	return ts.Get()
}

func (ts *threadSafe[T]) Store(value T) T {
	ts.Lock()
	defer ts.Unlock()
	ts.Set(value)
	return value
}

func (ts *threadSafe[T]) Do(f func(T) T) T {
	ts.Lock()
	defer ts.Unlock()
	return ts.Set(f(ts.Get()))
}

func (ts *threadSafe[T]) Get() T {
	return ts.value
}

func (ts *threadSafe[T]) Set(value T) T {
	ts.value = value
	return value
}

func (ts *threadSafe[T]) Lock() {
	ts.mutex.Lock()
}

func (ts *threadSafe[T]) Unlock() {
	ts.mutex.Unlock()
}

func (ts *threadSafe[T]) RLock() {
	ts.mutex.RLock()
}

func (ts *threadSafe[T]) RUnlock() {
	ts.mutex.RUnlock()
}

func LogError(err error) {
	stack := string(debug.Stack())
	lines := strings.Split(stack, "\n")
	lines = append(lines[:1], lines[5:len(lines)-1]...)
	stack = strings.Join(lines, "\n")
	log.Error("Error: " + err.Error() + "\n" + stack)
}
