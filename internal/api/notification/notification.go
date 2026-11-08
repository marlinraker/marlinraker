package notification

import (
	"encoding/json"
	"errors"

	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
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

	// Deliver to every connection even if writing to one of them fails, so a
	// single dead socket cannot starve the remaining clients.
	var errs []error
	for _, connection := range connections.GetConnections() {
		if err := connection.WriteText(jsonBytes); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
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
