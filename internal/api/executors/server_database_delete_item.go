package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/database"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerDatabaseDeleteItemResult struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     any    `json:"value"`
}

func ServerDatabaseDeleteItem(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	namespace, err := params.RequireString("namespace")
	if err != nil {
		return nil, err
	}

	key, err := params.RequireString("key")
	if err != nil {
		return nil, err
	}

	value, err := database.DeleteItem(namespace, key, false)
	if err != nil {
		return nil, err
	}

	return ServerDatabaseDeleteItemResult{
		Namespace: namespace,
		Key:       key,
		Value:     value,
	}, nil
}
