package executors

import (
	"marlinraker/src/database"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
)

type ServerDatabaseGetItemResult struct {
	Namespace string  `json:"namespace"`
	Key       *string `json:"key"`
	Value     any     `json:"value"`
}

func ServerDatabaseGetItem(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	namespace, exists := params.GetString("namespace")
	if !exists {
		return nil, util.NewError("namespace param is required", 400)
	}

	key, _ := params.GetString("key")

	value, err := database.GetItem(namespace, key)
	if err != nil {
		return nil, err
	}

	return ServerDatabaseGetItemResult{
		Namespace: namespace,
		Key:       util.StringOrNil(key),
		Value:     value,
	}, nil
}
