package executors

import (
	"marlinraker-go/src/database"
	"marlinraker-go/src/marlinraker/connections"
	"net/http"
)

type ServerDatabaseListResult struct {
	Namespaces []string `json:"namespaces"`
}

func ServerDatabaseList(*connections.Connection, *http.Request, Params) (any, error) {
	return ServerDatabaseListResult{
		Namespaces: database.ListNamespaces(),
	}, nil
}
