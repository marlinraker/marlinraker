package notification

import (
	"encoding/json"
	"marlinraker/src/marlinraker/connections"
)

type Notification interface {
	Method() string
	Params() []any
}

type rpc struct {
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params,omitempty"`
}

var Testing = false

func Publish(notification Notification) error {
	if Testing {
		return nil
	}

	jsonBytes, err := json.Marshal(rpc{
		JsonRpc: "2.0",
		Method:  notification.Method(),
		Params:  notification.Params(),
	})
	if err != nil {
		return err
	}

	for _, connection := range connections.GetConnections() {
		err = connection.WriteText(jsonBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func Send(connection *connections.Connection, notification Notification) error {
	if Testing {
		return nil
	}

	jsonBytes, err := json.Marshal(rpc{
		JsonRpc: "2.0",
		Method:  notification.Method(),
		Params:  notification.Params(),
	})
	if err != nil {
		return err
	}

	return connection.WriteText(jsonBytes)
}
