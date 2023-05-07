package executors

import (
	"marlinraker/src/database"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

type ServerDatabasePostItemResult struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     any    `json:"value"`
}

func ServerDatabasePostItem(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	namespace, exists := params.GetString("namespace")
	if !exists {
		return nil, util.NewError("namespace param is required", 400)
	}

	key, exists := params.GetString("key")
	if !exists {
		return nil, util.NewError("key param is required", 400)
	}

	value, exists := params["value"]
	if !exists {
		return nil, util.NewError("value param is required", 400)
	}

	_, err := database.PostItem(namespace, key, value)
	if err != nil {
		return nil, err
	}

	return ServerDatabasePostItemResult{
		Namespace: namespace,
		Key:       key,
		Value:     value,
	}, nil
}
