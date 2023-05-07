package executors

import (
	"marlinraker/src/database"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

type ServerDatabaseDeleteItemResult struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     any    `json:"value"`
}

func ServerDatabaseDeleteItem(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	namespace, exists := params.GetString("namespace")
	if !exists {
		return nil, util.NewError("namespace param is required", 400)
	}

	key, exists := params.GetString("key")
	if !exists {
		return nil, util.NewError("key param is required", 400)
	}

	value, err := database.DeleteItem(namespace, key)
	if err != nil {
		return nil, err
	}

	return ServerDatabaseDeleteItemResult{
		Namespace: namespace,
		Key:       key,
		Value:     value,
	}, nil
}
