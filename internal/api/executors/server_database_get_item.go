package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/database"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
	"github.com/marlinraker/marlinraker/internal/util"
)

type ServerDatabaseGetItemResult struct {
	Namespace string  `json:"namespace"`
	Key       *string `json:"key"`
	Value     any     `json:"value"`
}

func ServerDatabaseGetItem(_ *connections.Connection, _ *http.Request, params Params) (any, error) {
	namespace, err := params.RequireString("namespace")
	if err != nil {
		return nil, err
	}

	key, _ := params.GetString("key")

	value, err := database.GetItem(namespace, key, false)
	if err != nil {
		return nil, err
	}

	return ServerDatabaseGetItemResult{
		Namespace: namespace,
		Key:       util.StringOrNil(key),
		Value:     value,
	}, nil
}
