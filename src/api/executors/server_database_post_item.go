package executors

import (
	"marlinraker/src/database"
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type ServerDatabasePostItemResult struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     any    `json:"value"`
}

func ServerDatabasePostItem(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	namespace, err := params.RequireString("namespace")
	if err != nil {
		return nil, err
	}

	key, err := params.RequireString("key")
	if err != nil {
		return nil, err
	}

	value, err := params.RequireAny("value")
	if err != nil {
		return nil, err
	}

	if _, err := database.PostItem(namespace, key, value, false); err != nil {
		return nil, err
	}

	return ServerDatabasePostItemResult{
		Namespace: namespace,
		Key:       key,
		Value:     value,
	}, nil
}
