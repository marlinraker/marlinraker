package notification

type SimpleNotification struct {
	method string
	params []any
}

func (notification SimpleNotification) Method() string {
	return notification.method
}

func (notification SimpleNotification) Params() []any {
	return notification.params
}

func New(method string, params []any) SimpleNotification {
	return SimpleNotification{method, params}
}
