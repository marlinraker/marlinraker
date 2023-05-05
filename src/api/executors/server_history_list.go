// TODO

package executors

import (
	"marlinraker/src/marlinraker/connections"
	"net/http"
)

type ServerHistoryListResult struct {
	Count int   `json:"count"`
	Jobs  []any `json:"jobs"`
}

func ServerHistoryList(*connections.Connection, *http.Request, Params) (any, error) {
	return ServerHistoryListResult{
		Count: 0, Jobs: make([]any, 0),
	}, nil
}
