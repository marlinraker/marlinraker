package executors

import (
	"marlinraker-go/src/database"
	"marlinraker-go/src/marlinraker/connections"
)

type ServerDatabaseListResult struct {
	Namespaces []string `json:"namespaces"`
}

func ServerDatabaseList(*connections.Connection, Params) (any, error) {
	return ServerDatabaseListResult{
		Namespaces: database.ListNamespaces(),
	}, nil
}
