package util

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
