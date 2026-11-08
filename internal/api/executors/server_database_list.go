package executors

import (
	"net/http"

	"github.com/marlinraker/marlinraker/internal/database"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerDatabaseListResult struct {
	Namespaces []string `json:"namespaces"`
}

func ServerDatabaseList(*connections.Connection, *http.Request, Params) (any, error) {
	return ServerDatabaseListResult{
		Namespaces: database.ListNamespaces(),
	}, nil
}
