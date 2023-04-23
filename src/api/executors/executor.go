package executors

type Params map[string]any

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
